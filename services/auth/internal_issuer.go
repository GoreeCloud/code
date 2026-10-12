// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"testing"

	"forgejo.org/modules/jwtx"
)

var internalIssuers = make(map[string]InternalIssuer)

// Authorized Integrations can verify the signature of JWTs that the application itself generated without requiring
// remote access, and in a manner that is flexible to changes in [setting.AppURL].
//
// For example, Forgejo Actions is often used to access Forgejo with a JWT, by setting `enable-openid-connect: true` in
// a workflow.  Without any special support for this internal access situation, problems would occur:
//
// 1. Forgejo would need to make an HTTP request to itself to get the valid public key for the JWT, in order to validate
// its signature.  This is a waste of resources, and introduces a self-DoS risk.
//
// 2. Forgejo would need to be available via TLS in order for Actions to make service calls to Forgejo with that JWT
// (due to the TLS requirement for public key fetching).
//
// 3. Authorized Integrations would need to be saved with the `issuer` URL of Forgejo.  If Forgejo's own
// [setting.AppURL] changed, all the persisted records in the database would become incorrect.
//
// Internal Issuers work by registering a URL suffix like "api/actions".  When a JWT is received with an issuer
// matching [setting.AppURL] and the registered URL suffix, then the [InternalIssuer] interface is used to access the
// JWT public key, and the value to be saved in the Authorized Integrations table as the issuer.
func RegisterInternalIssuer(urlSuffix string, internalIssuer InternalIssuer) {
	internalIssuers[urlSuffix] = internalIssuer
}

// Variant of RegisterInternalIssuer which removes the registration impact in test cleanup.
func RegisterInternalIssuerForTesting(t *testing.T, urlSuffix string, internalIssuer InternalIssuer) {
	orig, hadOrig := internalIssuers[urlSuffix]
	internalIssuers[urlSuffix] = internalIssuer
	t.Cleanup(func() {
		if hadOrig {
			internalIssuers[urlSuffix] = orig
		} else {
			delete(internalIssuers, urlSuffix)
		}
	})
}

// Retrieve an internal issuer, if one exists, for the provided URL suffix from a JWT token.  For example,
// "api/actions".
func GetInternalIssuerByURLSuffix(issuerSuffix string) (InternalIssuer, bool) {
	ii, ok := internalIssuers[issuerSuffix]
	return ii, ok
}

// Read access to the registered internal issuers.
func GetInternalIssuers() map[string]InternalIssuer {
	return internalIssuers
}

//mockery:generate: true
type InternalIssuer interface {
	// Signing key used to validate a JWT from this internal issuer.
	SigningKey() jwtx.SigningKey
	// Value to store in [auth_model.AuthorizedIntegration]'s Issuer field to reflect the use of this internal issuer.
	IssuerPlaceholder() string
}
