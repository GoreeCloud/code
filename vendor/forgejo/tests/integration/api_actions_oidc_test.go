// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"encoding/base64"
	"net/http"
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jwksResponse map[string][]map[string]string

type openIDConfigurationResponse struct {
	Issuer                           string   `json:"issuer"`
	JwksURI                          string   `json:"jwks_uri"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                  []string `json:"scopes_supported"`
}

func prepareTestEnvActionsOIDC(t *testing.T) func() {
	t.Helper()
	f := tests.PrepareTestEnv(t, 1)
	return f
}

func TestActionsOIDC(t *testing.T) {
	defer prepareTestEnvActionsOIDC(t)()

	// get config information
	req := NewRequest(t, "GET", "/api/actions/.well-known/openid-configuration")
	resp := MakeRequest(t, req, http.StatusOK)
	var config openIDConfigurationResponse
	DecodeJSON(t, resp, &config)
	assert.Equal(t, setting.AppURL+"api/actions", config.Issuer)
	assert.Equal(t, setting.AppURL+"api/actions/.well-known/keys", config.JwksURI)
	assert.Equal(t, []string{"public"}, config.SubjectTypesSupported)
	assert.Equal(t, []string{"id_token"}, config.ResponseTypesSupported)
	assert.Equal(t, []string{"sub", "aud", "exp", "iat", "iss", "nbf", "actor", "base_ref", "event_name", "head_ref", "ref", "ref_protected", "ref_type", "repository", "repository_owner", "run_attempt", "run_id", "run_number", "sha", "workflow", "workflow_ref"}, config.ClaimsSupported)
	assert.Equal(t, []string{"RS256"}, config.IDTokenSigningAlgValuesSupported)
	assert.Equal(t, []string{"openid"}, config.ScopesSupported)

	// get JWKs information
	req = NewRequest(t, "GET", config.JwksURI)
	resp = MakeRequest(t, req, http.StatusOK)
	var jwks jwksResponse
	DecodeJSON(t, resp, &jwks)
	require.Len(t, jwks["keys"], 1)
	key := jwks["keys"][0]
	require.Equal(t, "RSA", key["kty"])
	require.Equal(t, "RS256", key["alg"])
	require.Equal(t, "sig", key["use"])

	// Basic validation of returned exponents
	if _, err := base64.RawURLEncoding.DecodeString(key["e"]); err != nil {
		t.Fatal(err)
	}

	if _, err := base64.RawURLEncoding.DecodeString(key["n"]); err != nil {
		t.Fatal(err)
	}
}
