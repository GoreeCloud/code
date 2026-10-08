// Copyright 2025 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "forgejo.org/models/auth"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPIConvert(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	user5 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})
	repo5 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 5})
	repo4 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 4})
	org3 := "org3"

	// Get user2's token
	session := loginUser(t, user2.Name)
	token2 := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)
	// Get user5's token
	session = loginUser(t, user5.Name)
	token5 := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	req := NewRequest(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/convert", org3, repo5.Name)).AddTokenAuth(token2)
	resp := MakeRequest(t, req, http.StatusOK)
	var repo api.Repository
	DecodeJSON(t, resp, &repo)
	assert.NotNil(t, repo)
	assert.False(t, repo.Mirror)

	repo5edited := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 5})
	assert.False(t, repo5edited.IsMirror)

	// Test editing a non-existing repo return 404
	name := "repodoesnotexist"
	req = NewRequest(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/convert", org3, name)).AddTokenAuth(token2)
	_ = MakeRequest(t, req, http.StatusNotFound)

	// Test converting a repo when not owner returns 422
	req = NewRequest(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/convert", org3, repo5.Name)).AddTokenAuth(token2)
	MakeRequest(t, req, http.StatusUnprocessableEntity)

	// Tests converting a repo with no token returns 404
	req = NewRequest(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/convert", org3, repo5.Name))
	_ = MakeRequest(t, req, http.StatusNotFound)

	// Test converting a repo that is not a mirror does nothing and returns 422
	req = NewRequest(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/convert", user5.Name, repo4.Name)).AddTokenAuth(token5)
	_ = MakeRequest(t, req, http.StatusUnprocessableEntity)
	repo4edited := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 4})
	assert.False(t, repo4edited.IsMirror)
}

// This test verifies that a repo-specific access token with `write:repository` scope is not a sufficient scope to edit
// the settings of a repository that is within its repo-specific list.
func TestAPIConvertAccessTokenResources(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo5 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 5})
	org3 := "org3"

	repoSpecificToken := createFineGrainedRepoAccessToken(t, "user2",
		[]auth_model.AccessTokenScope{auth_model.AccessTokenScopeWriteRepository},
		[]int64{repo5.ID},
	)
	req := NewRequest(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/convert", org3, repo5.Name)).AddTokenAuth(repoSpecificToken)
	MakeRequest(t, req, http.StatusForbidden)
}
