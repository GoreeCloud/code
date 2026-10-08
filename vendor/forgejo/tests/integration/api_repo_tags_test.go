// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIRepoTags(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	repoName := "repo1"

	req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/tags", user.Name, repoName).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)

	var tags []*api.Tag
	DecodeJSON(t, resp, &tags)

	assert.Len(t, tags, 1)
	assert.Equal(t, "v1.1", tags[0].Name)
	assert.Equal(t, "Initial commit", tags[0].Message)
	assert.Equal(t, "65f1bf27bc3bf70f64657658635e66094edbcb4d", tags[0].Commit.SHA)
	assert.Equal(t, setting.AppURL+"api/v1/repos/user2/repo1/git/commits/65f1bf27bc3bf70f64657658635e66094edbcb4d", tags[0].Commit.URL)
	assert.Equal(t, setting.AppURL+"user2/repo1/archive/v1.1.zip", tags[0].ZipballURL)
	assert.Equal(t, setting.AppURL+"user2/repo1/archive/v1.1.tar.gz", tags[0].TarballURL)

	newTag := createNewTagUsingAPI(t, token, user.Name, repoName, "gitea/22", "", "nice!\nand some text")
	resp = MakeRequest(t, req, http.StatusOK)
	DecodeJSON(t, resp, &tags)
	assert.Len(t, tags, 2)
	for _, tag := range tags {
		if tag.Name != "v1.1" {
			assert.Equal(t, newTag.Name, tag.Name)
			assert.Equal(t, newTag.Message, tag.Message)
			assert.Equal(t, "nice!\nand some text", tag.Message)
			assert.Equal(t, newTag.Commit.SHA, tag.Commit.SHA)
		}
	}

	// get created tag
	req = NewRequestf(t, "GET", "/api/v1/repos/%s/%s/tags/%s", user.Name, repoName, newTag.Name).
		AddTokenAuth(token)
	resp = MakeRequest(t, req, http.StatusOK)
	var tag *api.Tag
	DecodeJSON(t, resp, &tag)
	assert.Equal(t, newTag, tag)

	// delete tag
	delReq := NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/tags/%s", user.Name, repoName, newTag.Name).
		AddTokenAuth(token)
	MakeRequest(t, delReq, http.StatusNoContent)

	// check if it's gone
	MakeRequest(t, req, http.StatusNotFound)
}

func createNewTagUsingAPI(t *testing.T, token, ownerName, repoName, name, target, msg string) *api.Tag {
	urlStr := fmt.Sprintf("/api/v1/repos/%s/%s/tags", ownerName, repoName)
	req := NewRequestWithJSON(t, "POST", urlStr, &api.CreateTagOption{
		TagName: name,
		Message: msg,
		Target:  target,
	}).AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusCreated)

	var respObj api.Tag
	DecodeJSON(t, resp, &respObj)
	return &respObj
}

func TestAPIGetTagArchiveDownloadCount(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	repoName := "repo1"
	tagName := "TagDownloadCount"

	createNewTagUsingAPI(t, token, user.Name, repoName, tagName, "", "")

	urlStr := fmt.Sprintf("/api/v1/repos/%s/%s/tags/%s?token=%s", user.Name, repoName, tagName, token)

	req := NewRequest(t, "GET", urlStr)
	resp := MakeRequest(t, req, http.StatusOK)

	var tagInfo *api.Tag
	DecodeJSON(t, resp, &tagInfo)

	// Check if everything defaults to 0
	assert.Equal(t, int64(0), tagInfo.ArchiveDownloadCount.TarGz)
	assert.Equal(t, int64(0), tagInfo.ArchiveDownloadCount.Zip)

	// Download the tarball to increase the count
	MakeRequest(t, NewRequest(t, "GET", tagInfo.TarballURL), http.StatusOK)

	// Check if the count has increased
	resp = MakeRequest(t, req, http.StatusOK)

	DecodeJSON(t, resp, &tagInfo)

	assert.Equal(t, int64(1), tagInfo.ArchiveDownloadCount.TarGz)
	assert.Equal(t, int64(0), tagInfo.ArchiveDownloadCount.Zip)
}

func TestAPIGetTagsPaginated(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	repoName := "repo1"
	expectedTagName := "TagDownloadCount"

	for i := range 5 {
		createNewTagUsingAPI(t, token, user.Name, repoName, expectedTagName+fmt.Sprintf("%d", i), "", "")
	}

	// List tags with pagination
	req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/tags?limit=1", user.Name, repoName).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)

	var tags []*api.Tag
	DecodeJSON(t, resp, &tags)

	assert.Len(t, tags, 1)

	assert.Equal(t, fmt.Sprintf("%s%d", expectedTagName, 0), tags[0].Name)

	// Check if Link header is present for pagination
	link := resp.Header().Get("Link")
	assert.NotEmpty(t, link, "Link header should be set for paginated responses")
	assert.Contains(t, link, "rel=\"next\"")
	assert.Contains(t, link, "page=2")
}

func TestAPIRepoTagDeleteProtection(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	// Login as User2.
	session := loginUser(t, user.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	repoName := "repo1"

	req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/tags", user.Name, repoName).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)
	var tags []*api.Tag
	DecodeJSON(t, resp, &tags)
	require.Len(t, tags, 1)
	require.Equal(t, "v1.1", tags[0].Name)

	// Create a tag protection rule for the repo so that `user2` cannot create/remove tags, even if they have write
	// perms to the repo... which they do because they own it.
	req = NewRequestWithJSON(t, "POST",
		fmt.Sprintf("/api/v1/repos/%s/%s/tag_protections", user.Name, repoName),
		&api.CreateTagProtectionOption{
			NamePattern:        "v*",
			WhitelistUsernames: []string{"user1"},
		}).
		AddTokenAuth(token)
	resp = MakeRequest(t, req, http.StatusCreated)
	var tagProtection api.TagProtection
	DecodeJSON(t, resp, &tagProtection)
	require.Equal(t, "v*", tagProtection.NamePattern)

	// Delete the release associated with v1.1, so that it's possible to delete the tag.
	delReq := NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/releases/tags/%s", user.Name, repoName, tags[0].Name).
		AddTokenAuth(token)
	MakeRequest(t, delReq, http.StatusNoContent)

	// Attempt to delete the tag, which should be denied by the tag protection rule.
	delReq = NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/tags/%s", user.Name, repoName, tags[0].Name).
		AddTokenAuth(token)
	MakeRequest(t, delReq, http.StatusUnprocessableEntity)

	// Remove the tag protection rule.
	delReq = NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/tag_protections/%d", user.Name, repoName, tagProtection.ID).
		AddTokenAuth(token)
	MakeRequest(t, delReq, http.StatusNoContent)

	// Attempt to delete the tag again, which should now be permitted.
	delReq = NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/tags/%s", user.Name, repoName, tags[0].Name).
		AddTokenAuth(token)
	MakeRequest(t, delReq, http.StatusNoContent)
}
