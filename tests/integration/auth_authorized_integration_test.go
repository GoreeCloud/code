// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIAuthWithAuthorizedIntegration(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	t.Run("all access authorized integration", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		ait := newAITester(t)
		defer ait.close()
		token := ait.signedJWT()

		req := NewRequest(t, "GET", "/api/v1/user").AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)
		var user api.User
		DecodeJSON(t, resp, &user)
		assert.Equal(t, "user2", user.LoginName)

		req = NewRequest(t, "GET", "/api/v1/repos/user2/repo1").AddTokenAuth(token)
		MakeRequest(t, req, http.StatusOK)
	})

	t.Run("scope-limited authorized integration", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		ait := newAITester(t, func(ai *auth_model.AuthorizedIntegration) {
			ai.Scope = auth_model.AccessTokenScopeReadUser
		})
		defer ait.close()
		token := ait.signedJWT()

		req := NewRequest(t, "GET", "/api/v1/user").AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)
		var user api.User
		DecodeJSON(t, resp, &user)
		assert.Equal(t, "user2", user.LoginName)

		req = NewRequest(t, "GET", "/api/v1/repos/user2/repo1").AddTokenAuth(token)
		MakeRequest(t, req, http.StatusForbidden)
	})

	// Clone of TestAPICompareCommitsAccessTokenResources, but using an authorized integration rather than an access
	// token, to test its application of an authorization reducer for public-only and repo-specific access. Any API test
	// that uses repo-specific tokens could serve as a test here; this one is just relatively simple and succinct for
	// the number of things being tested.
	t.Run("authorization reducer", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Using the compare API, will be testing that the base repo's security checks implement fine-grained access
		// controls (and baselines with all and public-only).
		testCase := func(t *testing.T, repo, token string, expectedStatus int) {
			req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/compare/master...master", repo)).AddTokenAuth(token)
			MakeRequest(t, req, expectedStatus)
		}

		t.Run("all access token", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			ait := newAITester(t, func(ai *auth_model.AuthorizedIntegration) {
				ai.Scope = auth_model.AccessTokenScopeReadRepository
			})
			defer ait.close()
			allToken := ait.signedJWT()

			testCase(t, "user2/repo1", allToken, http.StatusOK)  // public user2/repo1
			testCase(t, "org3/repo3", allToken, http.StatusOK)   // private org3/repo3
			testCase(t, "user2/repo20", allToken, http.StatusOK) // private user2/repo20
		})

		t.Run("public-only access token", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			ait := newAITester(t, func(ai *auth_model.AuthorizedIntegration) {
				ai.Scope = auth_model.AccessTokenScope(fmt.Sprintf("%s,%s", auth_model.AccessTokenScopePublicOnly, auth_model.AccessTokenScopeReadRepository))
			})
			defer ait.close()
			publicOnlyToken := ait.signedJWT()

			testCase(t, "user2/repo1", publicOnlyToken, http.StatusOK)        // public user2/repo1
			testCase(t, "org3/repo3", publicOnlyToken, http.StatusNotFound)   // private org3/repo3
			testCase(t, "user2/repo20", publicOnlyToken, http.StatusNotFound) // private user2/repo20
		})

		t.Run("specific repo access token", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			ait := newAITester(t, func(ai *auth_model.AuthorizedIntegration) {
				ai.Scope = auth_model.AccessTokenScopeReadRepository
				ai.ResourceAllRepos = false
			})
			defer ait.close()

			_, err := db.GetEngine(t.Context()).Insert(&auth_model.AuthorizedIntegResourceRepo{
				IntegID: ait.authorizedIntegration.ID,
				RepoID:  3,
			})
			require.NoError(t, err)
			repo3OnlyToken := ait.signedJWT()

			testCase(t, "user2/repo1", repo3OnlyToken, http.StatusOK)        // public user2/repo1
			testCase(t, "org3/repo3", repo3OnlyToken, http.StatusOK)         // private org3/repo3
			testCase(t, "user2/repo20", repo3OnlyToken, http.StatusNotFound) // private user2/repo20, outside of fine-grain
		})
	})
}
