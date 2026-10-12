// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

// Response structure for a JWT issuer's `${iss}/.well-known/openid-configuration` URL endpoint; this is pared down to
// the relevant entries for authorized integrations to inspect from the remote issuer.
type AuthorizedIntegrationOpenIDConfiguration struct {
	Issuer                           string   `json:"issuer"`
	JwksURI                          string   `json:"jwks_uri"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

// Response structure for a JSON Web Key Set, which is typically read from the JwksURI field of [openIDConfiguration].
type AuthorizedIntegrationOpenIDKeys struct {
	// Typically map[string]string, for fields like "kty", "alg", "use", "kid", "n", "e", but also string:any for fields
	// like x5c which are []string. We currently don't parse any fields that aren't string, but we need to Unmarshal
	// into this field successfully in those cases.
	Keys []map[string]any `json:"keys"`
}
