// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	auth_model "forgejo.org/models/auth"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/jwtx"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/util"
	auth_service "forgejo.org/services/auth"
	"forgejo.org/services/authz"

	"github.com/gobwas/glob"
	"github.com/golang-jwt/jwt/v5"
)

var (
	_ auth_service.Method = &AuthorizedIntegration{}

	errParseInternalServer = errors.New("internal server error")
)

// Authenticates incoming requests by JWTs that are issued by an authorized integration.  Authorized integrations are
// stored in the database in the [auth_model.AuthorizedIntegration] table.  Once authenticated, the request can perform
// actions as the owner of the authorized integration, with limited access defined by the scope and resources stored on
// the database record.
//
// Authorization is received from HTTP requests as a `Authorization: Bearer [...jwt...]` (or `Authorization: Token
// [...jwt...]`).
type AuthorizedIntegration struct {
	// Permit the use of `Authorization: Basic ...`, in addition to the typical bearer/token authorization header.  If
	// true, the basic password will be interpreted as a JWT token if present and valid.  The username is ignored.
	PermitBasic bool

	// For testing -- interpret JWTs with now as a fixed time.
	fixedTime *time.Time
}

func (a *AuthorizedIntegration) Verify(req *http.Request, w http.ResponseWriter, _ auth_service.SessionStore) auth_service.MethodOutput {
	hasToken, token := tokenFromAuthorizationBearer(req).Get()
	if !hasToken {
		if !a.PermitBasic {
			return &auth_service.AuthenticationNotAttempted{}
		}
		hasBasic, basicToken := tokenFromAuthorizationBasic(req).Get()
		if !hasBasic {
			return &auth_service.AuthenticationNotAttempted{}
		}
		token = basicToken
	}

	var authorizedIntegration *auth_model.AuthorizedIntegration

	parsedToken, err := jwt.ParseWithClaims(token, &flexibleClaims{},
		func(t *jwt.Token) (any, error) {
			keyID, ok := t.Header["kid"]
			if !ok {
				return nil, errors.New("failed finding key identifer (kid) in JWT headers")
			}

			issuer, err := t.Claims.GetIssuer()
			if err != nil {
				return nil, fmt.Errorf("failed getting `iss` claim: %w", err)
			} else if len(issuer) == 0 {
				return nil, fmt.Errorf("invalid `iss` claim: %q", issuer)
			}
			audienceArray, err := t.Claims.GetAudience()
			if err != nil {
				return nil, fmt.Errorf("failed getting `aud` claim: %w", err)
			} else if len(audienceArray) != 1 {
				return nil, fmt.Errorf("required one and only one `aud` claim, but received %d", len(audienceArray))
			}
			audience := audienceArray[0]
			if len(audience) == 0 {
				return nil, fmt.Errorf("invalid `aud` claim: %q", audience)
			}

			// Check if there's an internal issuer that matches the JWT's issuer, and if so, change `queryIssuer` to the
			// internal issuer's placeholder, and store `internalIssuer` for later:
			queryIssuer := issuer
			var internalIssuer auth_service.InternalIssuer
			issuerSuffix := strings.TrimPrefix(issuer, setting.AppURL)
			if issuer != issuerSuffix { // TrimPrefix will return a different string when the prefix was present
				if ii, ok := auth_service.GetInternalIssuerByURLSuffix(issuerSuffix); ok {
					internalIssuer = ii
					queryIssuer = internalIssuer.IssuerPlaceholder()
				}
			}

			authorizedIntegration, err = auth_model.GetAuthorizedIntegration(req.Context(), queryIssuer, audience)
			if errors.Is(err, util.ErrNotExist) {
				return nil, errors.New("matching authorized_integration not found")
			} else if err != nil {
				return nil, fmt.Errorf("failure reading authorized_integration: %w (%w)", err, errParseInternalServer)
			}

			// Do the claim check before accessing the issuer's OIDC metadata and JWKS, to reduce risk of resource
			// utilization attack through invalid JWTs causing remote requests.
			err = a.checkClaims(t.Claims.(*flexibleClaims), authorizedIntegration.ClaimRules)
			if err != nil {
				return nil, fmt.Errorf("claim mismatch: %w", err)
			}

			// If an internal issuer was found earlier, then we can skip the JWKS fetch and just use its in-memory
			// signing key to validate the JWT.  It is critical we do this after the `checkClaims` above so that we
			// don't miss important validation of the JWT.
			if internalIssuer != nil {
				key := internalIssuer.SigningKey().VerifyKey()
				return key, nil
			}

			issuerURL, err := url.Parse(issuer)
			if err != nil {
				return nil, fmt.Errorf("failed parsing issuer: %w", err)
			}

			// Checks implemented here a variation of validateExternalIssuer used when creating an authorized
			// integration.  Where possible, if validation changes are made on either implementation, they should be
			// kept in sync with each other.

			issuerOIDCURL := issuerURL.JoinPath(".well-known/openid-configuration")
			var oidcConfig auth_service.AuthorizedIntegrationOpenIDConfiguration
			if err := auth_service.AuthorizedIntegrationFetchJSON(issuerOIDCURL.String(), &oidcConfig); err != nil {
				return nil, fmt.Errorf("error when fetching .well-known/openid-configuration from %s: %w", issuerOIDCURL, err)
			}

			if oidcConfig.Issuer != issuer {
				return nil, fmt.Errorf("issuer mismatch; expected %q, received %q from %s", issuer, oidcConfig.Issuer, issuerOIDCURL)
			}
			if !slices.Contains(oidcConfig.IDTokenSigningAlgValuesSupported, t.Method.Alg()) {
				return nil, fmt.Errorf("issuer supports signature algorithms %#v, but received token with algorithm %s", oidcConfig.IDTokenSigningAlgValuesSupported, t.Method.Alg())
			}

			jwksURI, err := url.Parse(oidcConfig.JwksURI)
			if err != nil {
				return nil, fmt.Errorf("failed parsing jwks_uri: %w", err)
			} else if jwksURI.Host != issuerURL.Host {
				// Prevent SSRF which could occur if a malicious openid-connection response returned a jwks_uri field
				// that causes Forgejo to access other hostnames.  This could be considered a valid case as well and we
				// can rely on the config-based allowed and blocked domains for the [authorized_integration] section,
				// but until a real-world case comes up where that is needed, this is a safety-first restriction.
				return nil, fmt.Errorf("jwks_uri host mismatch: must be the same as issuer host %q, but was %q", issuerURL.Host, jwksURI.Host)
			}
			var keys auth_service.AuthorizedIntegrationOpenIDKeys
			if err := auth_service.AuthorizedIntegrationFetchJSON(oidcConfig.JwksURI, &keys); err != nil {
				return nil, fmt.Errorf("error when fetching JWKS from %s: %w", oidcConfig.JwksURI, err)
			}

			availableKeyIDs := make([]string, 0, len(keys.Keys))
			for _, key := range keys.Keys {
				if key["kid"] == keyID {
					alg, algPresent := key["alg"] // "alg" is an optional field
					if algPresent && alg != t.Method.Alg() {
						return nil, fmt.Errorf("kid %q doesn't match expected algorithm %s, was %v", keyID, t.Method.Alg(), key["alg"])
					}

					use, usePresent := key["use"] // "use" is also an optional field
					if usePresent && use != "sig" {
						return nil, fmt.Errorf("kid %q isn't designated for signing usage, was %s", keyID, key["use"])
					}

					pub, err := jwtx.ParseJWKToPublicKey(key)
					if err != nil {
						return nil, fmt.Errorf("failed to parse JWKS: %w", err)
					}
					return pub, nil
				}
				availableKeyIDs = append(availableKeyIDs, fmt.Sprintf("%s", key["kid"]))
			}

			return nil, fmt.Errorf("no key identified; JWT was signed with key %q, but JWKS only provided keys with IDs: %s", keyID, strings.Join(availableKeyIDs, ", "))
		},
		jwt.WithValidMethods(jwtx.ValidAsymmetricAlgorithms), // only asymetric algorithms, as JWKS must have a public key only
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(func() time.Time {
			if a.fixedTime != nil {
				return *a.fixedTime
			}
			return time.Now()
		}),
	)
	if err != nil && errors.Is(err, errParseInternalServer) {
		// Errors from parsing marked errParseInternalServer are AuthenticationError, not incorrect creds:
		return &auth_service.AuthenticationError{Error: err}
	} else if err != nil {
		return &auth_service.AuthenticationAttemptedIncorrectCredential{Error: fmt.Errorf("authorized integration: parse JWT error: %w", err)}
	} else if !parsedToken.Valid {
		return &auth_service.AuthenticationAttemptedIncorrectCredential{Error: errors.New("authorized integration: JWT not valid")}
	} else if authorizedIntegration == nil { // shouldn't be possible, but overly safe
		return &auth_service.AuthenticationError{Error: errors.New("authorized integration: nil authorized integration")}
	}

	u, err := user_model.GetUserByID(req.Context(), authorizedIntegration.UserID)
	if err != nil {
		return &auth_service.AuthenticationError{Error: fmt.Errorf("authorized integration: GetUserByID: %w", err)}
	}

	if err = authorizedIntegration.UpdateLastUsed(req.Context()); err != nil {
		log.Error("UpdateLastUsed:  %v", err)
	}

	reducer, err := authz.GetAuthorizationReducerForAuthorizedIntegration(req.Context(), authorizedIntegration)
	if err != nil {
		return &auth_service.AuthenticationError{Error: fmt.Errorf("authorized integration GetAuthorizationReducerForAuthorizedIntegration: %w", err)}
	}

	var optionalExp optional.Option[timeutil.TimeStamp]
	if exp, err := parsedToken.Claims.GetExpirationTime(); err != nil {
		return &auth_service.AuthenticationError{Error: fmt.Errorf("authorized integration GetExpirationTime: %w", err)}
	} else if exp != nil {
		optionalExp = optional.Some(timeutil.TimeStamp(exp.Unix()))
	}

	return &auth_service.AuthenticationSuccess{
		Result: &authorizedIntegrationAuthenticationResult{
			user:      u,
			scope:     authorizedIntegration.Scope,
			reducer:   reducer,
			expiresAt: optionalExp,
		},
	}
}

// Compare a map[string]any of incoming claims against an array of claim rules.  All rules must match successfully or
// else an error with the mismatch detail is returned.
func (a *AuthorizedIntegration) checkClaims(incomingClaims any, stored *auth_model.ClaimRules) error {
	if stored == nil {
		return nil
	}

	for _, rule := range stored.Rules {
		var lhs any

		if lhsClaim, isFlex := incomingClaims.(*flexibleClaims); isFlex {
			switch rule.Claim {
			case "iss":
				lhs = lhsClaim.Issuer
			case "sub":
				lhs = lhsClaim.Subject
			case "jti":
				lhs = lhsClaim.ID
			case "aud":
				audienceArray, err := lhsClaim.GetAudience()
				if err != nil {
					return fmt.Errorf("failed getting `aud` claim: %w", err)
				} else if len(audienceArray) != 1 {
					return fmt.Errorf("required one and only one `aud` claim, but received %d", len(audienceArray))
				}
				lhs = audienceArray[0]
			default:
				v, present := lhsClaim.other[rule.Claim]
				if !present {
					return fmt.Errorf("claim rule on %q couldn't be satisfied: claim not found", rule.Claim)
				}
				lhs = v
			}
		} else if lhsMap, isMap := incomingClaims.(map[string]any); isMap {
			v, present := lhsMap[rule.Claim]
			if !present {
				return fmt.Errorf("claim rule on %q couldn't be satisfied: claim not found", rule.Claim)
			}
			lhs = v
		} else {
			return fmt.Errorf("unexpected incoming claims type: %T", incomingClaims)
		}

		switch rule.Comparison {
		case auth_model.ClaimEqual:
			lhsStr, ok := lhs.(string)
			if !ok {
				return fmt.Errorf("claim %q must be a string, but was %T", rule.Claim, lhs)
			} else if lhsStr != rule.Value {
				return fmt.Errorf("claim %q must be %q, but was %q", rule.Claim, rule.Value, lhsStr)
			}
		case auth_model.ClaimIn:
			lhsStr, ok := lhs.(string)
			if !ok {
				return fmt.Errorf("claim %q must be a string, but was %T", rule.Claim, lhs)
			} else if !slices.Contains(rule.Values, lhsStr) {
				return fmt.Errorf("claim %q must be one of %q, but was %q", rule.Claim, rule.Values, lhsStr)
			}
		case auth_model.ClaimGlob:
			lhsStr, ok := lhs.(string)
			if !ok {
				return fmt.Errorf("claim %q must be a string, but was %T", rule.Claim, lhs)
			}
			r, err := glob.Compile(rule.Value)
			if err != nil {
				return fmt.Errorf("unable to parse glob for claim rule on %q; glob = %q, err = %w", rule.Claim, rule.Value, err)
			}
			if !r.Match(lhsStr) {
				return fmt.Errorf("claim %q must match glob %q, but value %q did not match", rule.Claim, rule.Value, lhsStr)
			}
		case auth_model.ClaimGlobIn:
			lhsStr, ok := lhs.(string)
			if !ok {
				return fmt.Errorf("claim %q must be a string, but was %T", rule.Claim, lhs)
			}
			matched := false
			for _, g := range rule.Values {
				r, err := glob.Compile(g)
				if err != nil {
					return fmt.Errorf("unable to parse glob for claim rule on %q; glob = %q, err = %w", rule.Claim, g, err)
				}
				if r.Match(lhsStr) {
					matched = true
					break
				}
			}
			if !matched {
				return fmt.Errorf("claim %q must glob match one of %q, but value %q did not match", rule.Claim, rule.Values, lhsStr)
			}
		case auth_model.ClaimNested:
			lhsMap, ok := lhs.(map[string]any)
			if !ok {
				return fmt.Errorf("claim %q must be a map, but was %T", rule.Claim, lhs)
			} else if err := a.checkClaims(lhsMap, rule.Nested); err != nil {
				return fmt.Errorf("in nested claim %q: %w", rule.Claim, err)
			}
		}
	}

	return nil
}
