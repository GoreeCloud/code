// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"testing"

	auth_model "forgejo.org/models/auth"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"
	"forgejo.org/tests/forgery"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHookContent = `#!/bin/bash

echo Hello, World!
`

const repositoryIDWithPreReceiveHook = 37

func TestAPIListGitHooks(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: repositoryIDWithPreReceiveHook})

	// user1 is an admin user
	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/hooks/git", repo.FullName()).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)
	var apiGitHooks []*api.GitHook
	DecodeJSON(t, resp, &apiGitHooks)
	assert.Len(t, apiGitHooks, 3)
	for _, apiGitHook := range apiGitHooks {
		if apiGitHook.Name == "pre-receive" {
			assert.True(t, apiGitHook.IsActive)
			assert.Equal(t, testHookContent, apiGitHook.Content)
		} else {
			assert.False(t, apiGitHook.IsActive)
			assert.Empty(t, apiGitHook.Content)
		}
	}
}

func TestAPIGetGitHook(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: repositoryIDWithPreReceiveHook})

	// user1 is an admin user
	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName()).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)
	var apiGitHook *api.GitHook
	DecodeJSON(t, resp, &apiGitHook)
	assert.True(t, apiGitHook.IsActive)
	assert.Equal(t, testHookContent, apiGitHook.Content)
}

func TestAPIDeleteGitHook(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: repositoryIDWithPreReceiveHook})

	// user1 is an admin user
	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	req := NewRequestf(t, "DELETE", "/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName()).
		AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)

	req = NewRequestf(t, "GET", "/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName()).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)
	var apiGitHook2 *api.GitHook
	DecodeJSON(t, resp, &apiGitHook2)
	assert.False(t, apiGitHook2.IsActive)
	assert.Empty(t, apiGitHook2.Content)

	// after deletion, the sample webhook should be shown in the web interface
	resp = session.MakeRequest(t, NewRequestf(t, "GET", "/%s/settings/hooks/git/pre-receive", repo.FullName()), http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)
	sampleHook := htmlDoc.doc.Find("#content").Text()
	assert.Contains(t, sampleHook, "#!/bin/bash")
}

func TestAPIGitHooksFromEmpty(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	admin := forgery.CreateUser(t, &forgery.CreateUserOptions{
		IsAdmin: true, // only admin can view git hooks
	})
	session := loginUser(t, admin.Name)

	repo := forgery.CreateRepository(t, nil, nil) // admin does not need to own the repo

	t.Run("initially empty", func(t *testing.T) {
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)

		req := NewRequestf(t, "GET", "/api/v1/repos/%s/hooks/git", repo.FullName()).
			AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)
		var apiGitHooks []*api.GitHook
		DecodeJSON(t, resp, &apiGitHooks)
		assert.Len(t, apiGitHooks, 3)
		for _, apiGitHook := range apiGitHooks {
			assert.False(t, apiGitHook.IsActive)
			assert.Empty(t, apiGitHook.Content)
		}

		// no hooks folder or description file should be present
		entries, err := os.ReadDir(repo.RepoPath())
		require.NoError(t, err)
		assert.False(t, slices.ContainsFunc(entries, func(entry os.DirEntry) bool {
			return entry.Name() == "hooks"
		}), "hooks folder should be missing")
		assert.False(t, slices.ContainsFunc(entries, func(entry os.DirEntry) bool {
			return entry.Name() == "description"
		}), "description file should be missing")
	})

	t.Run("create pre-receive", func(t *testing.T) {
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

		urlStr := fmt.Sprintf("/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName())
		req := NewRequestWithJSON(t, "PATCH", urlStr, &api.EditGitHookOption{
			Content: testHookContent,
		}).AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)
		var apiGitHook *api.GitHook
		DecodeJSON(t, resp, &apiGitHook)
		assert.True(t, apiGitHook.IsActive)
		assert.Equal(t, testHookContent, apiGitHook.Content)

		req = NewRequestf(t, "GET", "/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName()).
			AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)
		var apiGitHook2 *api.GitHook
		DecodeJSON(t, resp, &apiGitHook2)
		assert.True(t, apiGitHook2.IsActive)
		assert.Equal(t, testHookContent, apiGitHook2.Content)

		// no description file should be present
		entries, err := os.ReadDir(repo.RepoPath())
		require.NoError(t, err)
		assert.True(t, slices.ContainsFunc(entries, func(entry os.DirEntry) bool {
			return entry.Name() == "hooks"
		}), "hooks folder should be present")
		assert.False(t, slices.ContainsFunc(entries, func(entry os.DirEntry) bool {
			return entry.Name() == "description"
		}), "description file should be missing")
	})

	t.Run("delete pre-receive", func(t *testing.T) {
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

		req := NewRequestf(t, "DELETE", "/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName()).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNoContent)
	})
}

func TestAPIGitHookNoAccess(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := forgery.CreateRepository(t, nil, nil)
	owner := repo.Owner
	require.False(t, owner.IsAdmin)
	session := loginUser(t, owner.Name)

	t.Run("Get", func(t *testing.T) {
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)
		req := NewRequestf(t, "GET", "/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName()).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusForbidden)
	})

	t.Run("List", func(t *testing.T) {
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)
		req := NewRequestf(t, "GET", "/api/v1/repos/%s/hooks/git", repo.FullName()).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusForbidden)
	})

	t.Run("Edit", func(t *testing.T) {
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)
		urlStr := fmt.Sprintf("/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName())
		req := NewRequestWithJSON(t, "PATCH", urlStr, &api.EditGitHookOption{
			Content: testHookContent,
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusForbidden)
	})

	t.Run("Delete", func(t *testing.T) {
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)
		req := NewRequestf(t, "DELETE", "/api/v1/repos/%s/hooks/git/pre-receive", repo.FullName()).
			AddTokenAuth(token)
		MakeRequest(t, req, http.StatusForbidden)
	})
}
