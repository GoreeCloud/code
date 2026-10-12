// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	"forgejo.org/modules/cache"
	"forgejo.org/modules/hostmatcher"
	"forgejo.org/modules/json"
	"forgejo.org/modules/log"
	"forgejo.org/modules/proxy"
	"forgejo.org/modules/setting"
	"forgejo.org/services/authz"

	"github.com/gobwas/glob"
)

var (
	ErrAuthorizedIntegrationBadUI = errors.New("invalid authorized integration UI")
	ErrInvalidIssuer              = errors.New("invalid issuer")
	ErrInvalidClaimRules          = errors.New("invalid claim rules")

	// Authorized Integration's HTTP client for remote OIDC metadata and key fetches:
	aiHTTPClient   *http.Client
	initHTTPClient sync.Once

	// Allow mocking / overridding during tests:
	GetAuthorizedIntegrationHTTPClient = func() *http.Client {
		initHTTPClient.Do(initAuthorizedIntegrationHTTPClient)
		return aiHTTPClient
	}
	GetAuthorizedIntegrationCache = cache.GetCache
)

// Restrict document size to prevent resource exhaustion attack with a malicious authorized integration; largest
// real-world openid-configuration observed is about 1kB, largest JWKS is 6kB, so for both cases 16kB should be
// sufficient. If this needs to change in the future, it could be moved to a config setting -- but until a reason comes
// up it seems reasonable to keep microscopic settings out-of-sight.
const authorizedIntegrationRequestBodyLimit = int64(16 * 1024)

func initAuthorizedIntegrationHTTPClient() {
	blockList := hostmatcher.ParseSimpleMatchList("authorized_integration.BLOCKED_DOMAINS", setting.AuthorizedIntegration.BlockedDomains)

	allowList := hostmatcher.ParseSimpleMatchList("authorized_integration.ALLOWED_DOMAINS", setting.AuthorizedIntegration.AllowedDomains)
	if allowList.IsEmpty() {
		// the default policy is that authorized integrations can access external hosts
		allowList.AppendBuiltin(hostmatcher.MatchBuiltinExternal)
	}
	if setting.AuthorizedIntegration.AllowLocalNetworks {
		allowList.AppendBuiltin(hostmatcher.MatchBuiltinPrivate)
		allowList.AppendBuiltin(hostmatcher.MatchBuiltinLoopback)
	}

	aiHTTPClient = &http.Client{
		Timeout: setting.AuthorizedIntegration.RequestTimeout,
		Transport: &http.Transport{
			Proxy:       proxy.Proxy(),
			DialContext: hostmatcher.NewDialContext("authorized_integration", allowList, blockList, setting.Proxy.ProxyURLFixed),
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// It might be possible to come up with some reasonable capability to support redirects -- such as
			// keeping them within the same issuer host? -- but there are risks that this can be used for SSRF
			// attacks.  In the face of those risks, and with a lack of real-world use-cases, disable redirects.
			return errors.New("authorized integration: HTTP redirects are disabled")
		},
	}
}

func authorizedIntegrationCacheKey(urlString string) string {
	return fmt.Sprintf("auth-int-remote:%s", urlString)
}

func authorizedIntegrationCacheGetJSON[K any](urlString string, v *K) bool {
	conn := GetAuthorizedIntegrationCache()
	if conn == nil {
		return false
	}

	cachedAny := conn.Get(authorizedIntegrationCacheKey(urlString))
	if cachedAny == nil {
		return false
	}
	cachedBytes, ok := cachedAny.([]byte)
	if !ok {
		cachedString, ok := cachedAny.(string)
		if !ok {
			log.Error("cached content was not []byte or string, but was %T", cachedAny)
			return false
		}
		cachedBytes = []byte(cachedString)
	}

	err := json.Unmarshal(cachedBytes, &v)
	if err != nil {
		// This error case shouldn't occur, as we only store data in the cache once we're sure we could unmarshal it.
		// If it does occur, log and fallback to treating as uncached.
		log.Error("failed to Unmarshal cached content: %s", err)
		// Caller may reuse `v` in a future unmarshal/decode call, and failure here may have polluted it.
		var zeroValue K
		*v = zeroValue
		return false
	}

	return true
}

func authorizedIntegrationCacheSetJSON(urlString string, buf []byte) {
	conn := GetAuthorizedIntegrationCache()
	if conn == nil {
		return
	}
	err := conn.Put(authorizedIntegrationCacheKey(urlString), buf, int64(setting.AuthorizedIntegration.CacheTTL.Seconds()))
	if err != nil {
		log.Error("failed to put cache: %s", err)
	}
}

func AuthorizedIntegrationFetchJSON[K any](urlString string, v *K) error {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("failed parsing URL %q: %w", urlString, err)
	}
	// Fetching openid-connect or JWKS needs to come from a source that is authentic, and therefore only `https` is
	// supported.  This also protects against a trusted issuer being configured maliciously  as `file://` or a JKWS URI
	// being `file://` -- the HTTP client won't permit that, but, extra safety doesn't hurt.
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %q", parsedURL.String())
	}

	// Check our cache, save a remote HTTP interaction.
	if authorizedIntegrationCacheGetJSON(urlString, v) {
		return nil
	}

	resp, err := GetAuthorizedIntegrationHTTPClient().Get(parsedURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK response code: %s", resp.Status)
	}

	bodyReader := io.LimitReader(resp.Body, authorizedIntegrationRequestBodyLimit)
	var buf bytes.Buffer
	_, err = io.Copy(bufio.NewWriter(&buf), bodyReader)
	if err != nil {
		return fmt.Errorf("read from remote error: %w", err)
	}

	err = json.Unmarshal(buf.Bytes(), &v)
	if err != nil {
		// If a decoding error is hit, decorate with information about the limited body size so that it doesn't look
		// like the remote server provided an incomplete response. err should be something like `io.UnexpectedEOF` in
		// this case, but it actually isn't, so don't bother trying to detect precisely.
		return fmt.Errorf("failed to decode (response body restricted to %d bytes): %w", authorizedIntegrationRequestBodyLimit, err)
	}

	// Successfully decoded the response -- cache the raw bytes for later access.
	authorizedIntegrationCacheSetJSON(urlString, buf.Bytes())

	return nil
}

type MissingFieldError struct {
	Field string
}

func (e *MissingFieldError) Error() string {
	return fmt.Sprintf("missing field %s", e.Field)
}

// Validate that an authorized integration's state is valid for creation.  For example, that it doesn't have a
// conflicting set of resources (public-only and specific repositories), and other similar checks.
func ValidateAuthorizedIntegration(ai *auth_model.AuthorizedIntegration, repoResources []*auth_model.AuthorizedIntegResourceRepo) error {
	if ai.Name == "" {
		return &MissingFieldError{Field: "Name"}
	}

	switch ai.UI {
	case auth_model.AuthorizedIntegrationUIGeneric,
		auth_model.AuthorizedIntegrationUIForgejoActionsLocal:
		break
	default:
		return fmt.Errorf("%w: invalid UI: %q", ErrAuthorizedIntegrationBadUI, ai.UI)
	}

	internalIssuer := false
	for _, ii := range GetInternalIssuers() {
		if ai.Issuer == ii.IssuerPlaceholder() {
			internalIssuer = true
			break
		}
	}
	if !internalIssuer {
		if err := validateExternalIssuer(ai.Issuer); err != nil {
			return err
		}
	}

	if err := validateClaimRules(ai.ClaimRules, "root"); err != nil {
		return err
	}

	return authz.ValidateRepositoryResource(ai.ResourceAllRepos, ai.Scope, len(repoResources))
}

// Validate and insert a new authorized integration.
func InsertAuthorizedIntegration(ctx context.Context, ai *auth_model.AuthorizedIntegration, repoResources []*auth_model.AuthorizedIntegResourceRepo) error {
	ai.Name = strings.TrimSpace(ai.Name)
	ai.Description = strings.TrimSpace(ai.Description)

	if err := ValidateAuthorizedIntegration(ai, repoResources); err != nil {
		return err
	}

	return db.WithTx(ctx, func(ctx context.Context) error {
		if err := auth_model.InsertAuthorizedIntegration(ctx, ai); err != nil {
			return err
		}
		if !ai.ResourceAllRepos {
			if err := auth_model.InsertAuthorizedIntegrationResourceRepos(ctx, ai.ID, repoResources); err != nil {
				return err
			}
		}
		return nil
	})
}

func UpdateAuthorizedIntegration(ctx context.Context, ai *auth_model.AuthorizedIntegration, repoResources []*auth_model.AuthorizedIntegResourceRepo) error {
	ai.Name = strings.TrimSpace(ai.Name)
	ai.Description = strings.TrimSpace(ai.Description)

	if err := ValidateAuthorizedIntegration(ai, repoResources); err != nil {
		return err
	}

	return db.WithTx(ctx, func(ctx context.Context) error {
		if err := auth_model.UpdateAuthorizedIntegration(ctx, ai); err != nil {
			return err
		}
		return auth_model.UpdateAuthorizedIntegrationResourceRepos(ctx, ai.ID, repoResources)
	})
}

func validateExternalIssuer(issuer string) error {
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return fmt.Errorf("%w: failed parsing issuer URL: %w", ErrInvalidIssuer, err)
	}

	// Checks implemented here a variation of [AuthorizedIntegration.Verify]'s checks on the remote issuer.  Where
	// possible, if validation changes are made on either implementation, they should be kept in sync with each other.

	issuerOIDCURL := issuerURL.JoinPath(".well-known/openid-configuration")
	var oidcConfig AuthorizedIntegrationOpenIDConfiguration
	if err := AuthorizedIntegrationFetchJSON(issuerOIDCURL.String(), &oidcConfig); err != nil {
		return fmt.Errorf("%w: error when fetching .well-known/openid-configuration from %s: %w", ErrInvalidIssuer, issuerOIDCURL, err)
	}
	if oidcConfig.Issuer != issuer {
		return fmt.Errorf("%w: .well-known/openid-configuration from %s has issuer %q, but input issuer was %q", ErrInvalidIssuer, issuerOIDCURL, oidcConfig.Issuer, issuer)
	} else if len(oidcConfig.IDTokenSigningAlgValuesSupported) == 0 {
		return fmt.Errorf("%w: .well-known/openid-configuration from %s lacks required field id_token_signing_alg_values_supported", ErrInvalidIssuer, issuerOIDCURL)
	} else if oidcConfig.JwksURI == "" {
		return fmt.Errorf("%w: .well-known/openid-configuration from %s lacks required field jwks_uri", ErrInvalidIssuer, issuerOIDCURL)
	}

	jwksURI, err := url.Parse(oidcConfig.JwksURI)
	if err != nil {
		return fmt.Errorf("%w: .well-known/openid-configuration from %s has invalid jwks_uri: %w", ErrInvalidIssuer, issuerOIDCURL, err)
	} else if jwksURI.Host != issuerURL.Host {
		return fmt.Errorf("%w: .well-known/openid-configuration from %s has jwks_uri host mismatch: must be the same as issuer host %q, but was %q", ErrInvalidIssuer, issuerOIDCURL, issuerURL.Host, jwksURI.Host)
	}

	var keys AuthorizedIntegrationOpenIDKeys
	if err := AuthorizedIntegrationFetchJSON(oidcConfig.JwksURI, &keys); err != nil {
		return fmt.Errorf("%w: error when fetching JWKS from %s: %w", ErrInvalidIssuer, oidcConfig.JwksURI, err)
	} else if len(keys.Keys) == 0 {
		return fmt.Errorf("%w: fetching JWKS from %s had zero keys", ErrInvalidIssuer, oidcConfig.JwksURI)
	}

	return nil
}

func validateClaimRules(cr *auth_model.ClaimRules, path string) error {
	if cr == nil {
		return fmt.Errorf("%w: claim rules are nil at %s", ErrInvalidClaimRules, path)
	}

	for ruleIndex, r := range cr.Rules {
		if r.Claim == "" {
			return fmt.Errorf("%w: claim is missing at %s[%d]", ErrInvalidClaimRules, path, ruleIndex)
		}
		switch r.Comparison {
		case auth_model.ClaimEqual:
			if r.Value == "" {
				return fmt.Errorf("%w: claim value missing at %s[%d].value", ErrInvalidClaimRules, path, ruleIndex)
			}
		case auth_model.ClaimGlob:
			if r.Value == "" {
				return fmt.Errorf("%w: claim value missing at %s[%d].value", ErrInvalidClaimRules, path, ruleIndex)
			} else if _, err := glob.Compile(r.Value); err != nil {
				return fmt.Errorf("%w: claim glob invalid at %s[%d].value: %w", ErrInvalidClaimRules, path, ruleIndex, err)
			}
		case auth_model.ClaimIn:
			if len(r.Values) == 0 {
				return fmt.Errorf("%w: claim values missing at %s[%d].values", ErrInvalidClaimRules, path, ruleIndex)
			}
		case auth_model.ClaimGlobIn:
			if len(r.Values) == 0 {
				return fmt.Errorf("%w: claim values missing at %s[%d].values", ErrInvalidClaimRules, path, ruleIndex)
			}
			for globIndex, g := range r.Values {
				if g == "" {
					return fmt.Errorf("%w: claim glob empty string invalid, would match anything, at %s[%d].values[%d]", ErrInvalidClaimRules, path, ruleIndex, globIndex)
				} else if _, err := glob.Compile(g); err != nil {
					return fmt.Errorf("%w: claim glob invalid at %s[%d].values[%d]: %w", ErrInvalidClaimRules, path, ruleIndex, globIndex, err)
				}
			}
		case auth_model.ClaimNested:
			if err := validateClaimRules(r.Nested, fmt.Sprintf("%s.%s", path, r.Claim)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%w: compare %q is not valid at %s[%d]", ErrInvalidClaimRules, r.Comparison, path, ruleIndex)
		}
	}

	return nil
}
