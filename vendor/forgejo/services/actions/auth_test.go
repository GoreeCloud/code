// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"net/http"
	"testing"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/modules/json"
	"forgejo.org/modules/setting"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAuthorizationToken(t *testing.T) {
	task := &actions_model.ActionTask{
		ID: 23,
		Job: &actions_model.ActionRunJob{
			ID:    2,
			RunID: 1,
		},
	}

	testcases := []struct {
		name                string
		enableOpenIDConnect bool
		gitCtx              map[string]any
	}{
		{
			name:                "enableOpenIDConnect false",
			enableOpenIDConnect: false,
			gitCtx:              map[string]any{},
		},
		{
			name:                "enableOpenIDConnect true",
			enableOpenIDConnect: true,
			gitCtx: map[string]any{
				"actor":            "user1",
				"base_ref":         "master",
				"event_name":       "push",
				"head_ref":         "master",
				"ref":              "refs/heads/master",
				"ref_protected":    "false",
				"ref_type":         "branch",
				"repository":       "mpminardi/testing",
				"repository_owner": "mpminardi",
				"run_attempt":      "1",
				"run_id":           "1",
				"run_number":       "1",
				"sha":              "pretend-sha",
				"workflow":         "test.yml",
				"workflow_ref":     "pretend-ref",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := CreateAuthorizationToken(task, tc.gitCtx, tc.enableOpenIDConnect)
			require.NoError(t, err)
			assert.NotEmpty(t, token)
			claims := jwt.MapClaims{}
			_, err = jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
				return setting.GetGeneralTokenSigningSecret(), nil
			})
			require.NoError(t, err)
			scp, ok := claims["scp"]
			assert.True(t, ok, "Has scp claim in jwt token")
			assert.Contains(t, scp, "Actions.Results:1:2")
			taskIDClaim, ok := claims["TaskID"]
			assert.True(t, ok, "Has TaskID claim in jwt token")
			assert.InDelta(t, float64(task.ID), taskIDClaim, 0, "Supplied taskid must match stored one")
			acClaim, ok := claims["ac"]
			assert.True(t, ok, "Has ac claim in jwt token")
			ac, ok := acClaim.(string)
			assert.True(t, ok, "ac claim is a string for buildx gha cache")
			scopes := []actionsCacheScope{}
			err = json.Unmarshal([]byte(ac), &scopes)
			require.NoError(t, err, "ac claim is a json list for buildx gha cache")
			assert.GreaterOrEqual(t, len(scopes), 1, "Expected at least one action cache scope for buildx gha cache")

			if tc.enableOpenIDConnect {
				assert.Contains(t, scp, "generate_id_token:1:2")
				oidcSubClaim, ok := claims["oidc_sub"]
				assert.True(t, ok, "Has oidc_sub claim in jwt token")
				assert.Equal(t, "repo:mpminardi/testing:ref:refs/heads/master", oidcSubClaim)
				oidcExtraClaim, ok := claims["oidc_extra"]
				assert.True(t, ok, "Has oidc_extra claim in jwt token")
				val, err := json.Marshal(tc.gitCtx)
				require.NoError(t, err)
				assert.Equal(t, string(val), oidcExtraClaim)
			} else {
				assert.NotContains(t, scp, "generate_id_token")
				_, ok := claims["oidc_sub"]
				assert.False(t, ok, "Does not have oidc_sub claim in jwt token")
				_, ok = claims["oidc_extra"]
				assert.False(t, ok, "Does not have oidc_extra claim in jwt token")
			}
		})
	}
}

func TestParseAuthorizationToken(t *testing.T) {
	task := &actions_model.ActionTask{
		ID: 23,
		Job: &actions_model.ActionRunJob{
			ID:    2,
			RunID: 1,
		},
	}
	token, err := CreateAuthorizationToken(task, map[string]any{}, false)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)
	rTaskID, err := ParseAuthorizationToken(&http.Request{
		Header: headers,
	})
	require.NoError(t, err)
	assert.Equal(t, task.ID, rTaskID)
}

func TestParseAuthorizationTokenClaims(t *testing.T) {
	task := &actions_model.ActionTask{
		ID: 23,
		Job: &actions_model.ActionRunJob{
			ID:    2,
			RunID: 1,
		},
	}
	gitCtx := map[string]any{
		"actor":            "user1",
		"base_ref":         "master",
		"event_name":       "push",
		"head_ref":         "master",
		"ref":              "refs/heads/master",
		"ref_protected":    "false",
		"ref_type":         "branch",
		"repository":       "mpminardi/testing",
		"repository_owner": "mpminardi",
		"run_attempt":      "1",
		"run_id":           "1",
		"run_number":       "1",
		"sha":              "pretend-sha",
		"workflow":         "test.yml",
		"workflow_ref":     "pretend-ref",
	}
	token, err := CreateAuthorizationToken(task, gitCtx, true)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+token)
	tokenClaims, err := ParseAuthorizationTokenClaims(&http.Request{
		Header: headers,
	})
	require.NoError(t, err)
	assert.Equal(t, task.ID, tokenClaims.TaskID)
	assert.Equal(t, task.Job.ID, tokenClaims.JobID)
	assert.Equal(t, task.Job.RunID, tokenClaims.RunID)
	var customClaims map[string]any
	err = json.Unmarshal([]byte(tokenClaims.OIDCExtra), &customClaims)
	require.NoError(t, err)
	assert.Equal(t, gitCtx, customClaims)
	assert.Equal(t, "repo:mpminardi/testing:ref:refs/heads/master", tokenClaims.OIDCSub)
}

func TestParseAuthorizationTokenNoAuthHeader(t *testing.T) {
	headers := http.Header{}
	rTaskID, err := ParseAuthorizationToken(&http.Request{
		Header: headers,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), rTaskID)
}

func TestGenerateOIDCSub(t *testing.T) {
	t.Run("pull_request event", func(t *testing.T) {
		sub := generateOIDCSub(map[string]any{
			"event_name": "pull_request",
			"repository": "mpminardi/testing",
			"ref":        "refs/heads/master",
		})

		assert.Equal(t, "repo:mpminardi/testing:pull_request", sub)
	})

	t.Run("other event", func(t *testing.T) {
		sub := generateOIDCSub(map[string]any{
			"event_name": "random",
			"repository": "mpminardi/testing",
			"ref":        "refs/heads/master",
		})

		assert.Equal(t, "repo:mpminardi/testing:ref:refs/heads/master", sub)
	})
}
