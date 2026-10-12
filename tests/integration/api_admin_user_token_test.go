// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIAdminCreateUserAccessToken(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	adminUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user1"})
	targetUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user2"})

	token := getUserToken(t, adminUser.Name, auth_model.AccessTokenScopeWriteAdmin)

	t.Run("Create token for another user", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		urlStr := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "admin-created-token",
			Scopes: []string{"all"},
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var newToken api.AccessToken
		DecodeJSON(t, resp, &newToken)

		assert.Equal(t, "admin-created-token", newToken.Name)
		assert.NotEmpty(t, newToken.Token)
		assert.NotEmpty(t, newToken.TokenLastEight)
		assert.Contains(t, newToken.Scopes, "all")
		assert.NotZero(t, newToken.Created)

		// Verify the token exists in DB
		unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{
			ID:   newToken.ID,
			Name: newToken.Name,
			UID:  targetUser.ID,
		})
	})

	t.Run("Create token with duplicate name", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		urlStr := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "admin-created-token",
			Scopes: []string{"all"},
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusBadRequest)
	})

	t.Run("Create token without scopes", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		urlStr := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "empty-scope-token",
			Scopes: []string{},
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusBadRequest)
	})

	t.Run("Create token with invalid scope", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		urlStr := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "invalid-scope-token",
			Scopes: []string{"invalid-scope"},
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusBadRequest)
	})

	t.Run("Create token for nonexistent user", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		urlStr := "/api/v1/admin/users/nonexistentuser/tokens"
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "some-token",
			Scopes: []string{"all"},
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("Non-admin cannot create token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		normalToken := getUserToken(t, "user2", auth_model.AccessTokenScopeWriteAdmin)
		urlStr := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "unauthorized-token",
			Scopes: []string{"all"},
		}).AddTokenAuth(normalToken)
		MakeRequest(t, req, http.StatusForbidden)
	})
}

func TestAPIAdminListUserAccessTokens(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	adminUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user1"})
	targetUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user2"})

	token := getUserToken(t, adminUser.Name, auth_model.AccessTokenScopeWriteAdmin)

	// First, create a token for the target user
	createURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
	req := NewRequestWithJSON(t, "POST", createURL, api.CreateAccessTokenOption{
		Name:   "list-test-token",
		Scopes: []string{"all"},
	}).AddTokenAuth(token)
	MakeRequest(t, req, http.StatusCreated)

	t.Run("List tokens for user", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		listURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequest(t, "GET", listURL).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)

		var tokens []*api.AccessToken
		DecodeJSON(t, resp, &tokens)

		// user2 has at least the token we just created plus any fixture tokens
		require.NotEmpty(t, tokens)

		found := false
		for _, tk := range tokens {
			if tk.Name == "list-test-token" {
				found = true
				assert.NotEmpty(t, tk.TokenLastEight)
				assert.NotZero(t, tk.Created)
				break
			}
		}
		assert.True(t, found, "should find the admin-created token in the list")
	})

	t.Run("Non-admin cannot list tokens", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		normalToken := getUserToken(t, "user2", auth_model.AccessTokenScopeWriteAdmin)
		listURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequest(t, "GET", listURL).AddTokenAuth(normalToken)
		MakeRequest(t, req, http.StatusForbidden)
	})
}

func TestAPIAdminCreateRepoSpecificToken(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	adminUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user1"})
	targetUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user2"})

	token := getUserToken(t, adminUser.Name, auth_model.AccessTokenScopeWriteAdmin)

	t.Run("Create repo-specific token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		urlStr := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "admin-repo-specific-token",
			Scopes: []string{string(auth_model.AccessTokenScopeReadRepository)},
			Repositories: []*api.RepoTargetOption{
				{
					Owner: "user2",
					Name:  "repo2",
				},
			},
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var newToken api.AccessToken
		DecodeJSON(t, resp, &newToken)

		assert.Equal(t, "admin-repo-specific-token", newToken.Name)
		assert.NotEmpty(t, newToken.Token)
		assert.NotEmpty(t, newToken.Repositories)
	})

	t.Run("Create token targeting invalid repo", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		urlStr := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateAccessTokenOption{
			Name:   "admin-invalid-repo-token",
			Scopes: []string{string(auth_model.AccessTokenScopeReadRepository)},
			Repositories: []*api.RepoTargetOption{
				{
					Owner: "user10000",
					Name:  "repo70000",
				},
			},
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusBadRequest)
	})

	t.Run("List repo-specific token returns repositories", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a repo-specific token
		createURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", createURL, api.CreateAccessTokenOption{
			Name:   "admin-list-repo-token",
			Scopes: []string{string(auth_model.AccessTokenScopeReadRepository)},
			Repositories: []*api.RepoTargetOption{
				{
					Owner: "user2",
					Name:  "repo2",
				},
			},
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusCreated)

		// List tokens and verify repositories are returned
		listURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req = NewRequest(t, "GET", listURL).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)

		var tokens []*api.AccessToken
		DecodeJSON(t, resp, &tokens)

		found := false
		for _, tk := range tokens {
			if tk.Name == "admin-list-repo-token" {
				found = true
				assert.NotEmpty(t, tk.Repositories, "admin-listed token should have repositories populated")
				break
			}
		}
		assert.True(t, found, "should find the repo-specific token in the admin list")
	})
}

func TestAPIAdminDeleteUserAccessToken(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	adminUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user1"})
	targetUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{Name: "user2"})

	token := getUserToken(t, adminUser.Name, auth_model.AccessTokenScopeWriteAdmin)

	t.Run("Delete token by ID", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a token first
		createURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", createURL, api.CreateAccessTokenOption{
			Name:   "delete-by-id-token",
			Scopes: []string{"all"},
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var newToken api.AccessToken
		DecodeJSON(t, resp, &newToken)

		// Delete it
		deleteURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens/%d", targetUser.Name, newToken.ID)
		req = NewRequest(t, "DELETE", deleteURL).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNoContent)

		// Verify it's gone
		unittest.AssertNotExistsBean(t, &auth_model.AccessToken{ID: newToken.ID})
	})

	t.Run("Delete token by name", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a token first
		createURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", createURL, api.CreateAccessTokenOption{
			Name:   "delete-by-name-token",
			Scopes: []string{"all"},
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var newToken api.AccessToken
		DecodeJSON(t, resp, &newToken)

		// Delete by name
		deleteURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens/%s", targetUser.Name, "delete-by-name-token")
		req = NewRequest(t, "DELETE", deleteURL).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNoContent)

		// Verify it's gone
		unittest.AssertNotExistsBean(t, &auth_model.AccessToken{ID: newToken.ID})
	})

	t.Run("Delete nonexistent token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		deleteURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens/%d", targetUser.Name, 999999)
		req := NewRequest(t, "DELETE", deleteURL).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("Non-admin cannot delete token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Create a token as admin
		createURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens", targetUser.Name)
		req := NewRequestWithJSON(t, "POST", createURL, api.CreateAccessTokenOption{
			Name:   "non-admin-delete-token",
			Scopes: []string{"all"},
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusCreated)

		var newToken api.AccessToken
		DecodeJSON(t, resp, &newToken)

		// Try to delete as non-admin
		normalToken := getUserToken(t, "user2", auth_model.AccessTokenScopeWriteAdmin)
		deleteURL := fmt.Sprintf("/api/v1/admin/users/%s/tokens/%d", targetUser.Name, newToken.ID)
		req = NewRequest(t, "DELETE", deleteURL).AddTokenAuth(normalToken)
		MakeRequest(t, req, http.StatusForbidden)
	})
}
