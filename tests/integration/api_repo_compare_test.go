// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/git"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPICompareCommits(t *testing.T) {
	forEachObjectFormat(t, testAPICompareCommits)
}

func testAPICompareCommits(t *testing.T, objectFormat git.ObjectFormat) {
	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		newBranchAndFile := func(ctx APITestContext, user *user_model.User, branch, filename string) func(*testing.T) {
			return func(t *testing.T) {
				doAPICreateFile(ctx, filename, &api.CreateFileOptions{
					FileOptions: api.FileOptions{
						NewBranchName: branch,
						Message:       "create " + filename,
						Author: api.Identity{
							Name:  user.Name,
							Email: user.Email,
						},
						Committer: api.Identity{
							Name:  user.Name,
							Email: user.Email,
						},
						Dates: api.CommitDateOptions{
							Author:    time.Now(),
							Committer: time.Now(),
						},
					},
					ContentBase64: base64.StdEncoding.EncodeToString([]byte("content " + filename)),
				})(t)
			}
		}

		requireErrorContains := func(t *testing.T, resp *httptest.ResponseRecorder, expected string) {
			t.Helper()

			type response struct {
				Message string   `json:"message"`
				Errors  []string `json:"errors"`
			}
			var bodyResp response
			DecodeJSON(t, resp, &bodyResp)

			if strings.Contains(bodyResp.Message, expected) {
				return
			}
			for _, error := range bodyResp.Errors {
				if strings.Contains(error, expected) {
					return
				}
			}
			t.Log(fmt.Sprintf("expected %s in %+v", expected, bodyResp))
			t.Fail()
		}

		user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
		user2repo := "repoA"
		user2Ctx := NewAPITestContext(t, user2.Name, user2repo, auth_model.AccessTokenScopeWriteRepository, auth_model.AccessTokenScopeWriteUser)
		t.Run("CreateUser2Repository", doAPICreateRepository(user2Ctx, &api.CreateRepoOption{
			AutoInit:    true,
			Description: "Temporary repo",
			Name:        user2Ctx.Reponame,
		}, objectFormat))
		user2branchName := "user2branch"
		t.Run("CreateUser2RepositoryBranch", newBranchAndFile(user2Ctx, user2, user2branchName, "user2branchfilename.txt"))
		user2branch := doAPIGetBranch(user2Ctx, user2branchName)(t)
		user2master := doAPIGetBranch(user2Ctx, "master")(t)
		user2tag1 := "tag1"
		t.Run("CreateUser2RepositoryTag1", doAPICreateTag(user2Ctx, user2tag1, "master", "user2branchtag1"))
		user2tag2 := "tag2"
		t.Run("CreateUser2RepositoryTag1", doAPICreateTag(user2Ctx, user2tag2, user2branchName, "user2branchtag2"))

		shortCommitLength := 7

		for _, testCase := range []struct {
			name string
			a    string
			b    string
		}{
			{
				name: "Commits",
				a:    user2master.Commit.ID,
				b:    user2branch.Commit.ID,
			},
			{
				name: "ShortCommits",
				a:    user2master.Commit.ID[:shortCommitLength],
				b:    user2branch.Commit.ID[:shortCommitLength],
			},
			{
				name: "Branches",
				a:    "master",
				b:    user2branchName,
			},
			{
				name: "Tags",
				a:    user2tag1,
				b:    user2tag2,
			},
		} {
			t.Run("SameRepo"+testCase.name, func(t *testing.T) {
				// a...b
				req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/%s...%s", user2.Name, user2repo, testCase.a, testCase.b).
					AddTokenAuth(user2Ctx.Token)
				resp := MakeRequest(t, req, http.StatusOK)

				var apiResp *api.Compare
				DecodeJSON(t, resp, &apiResp)

				assert.Equal(t, 1, apiResp.TotalCommits)
				assert.Len(t, apiResp.Commits, 1)
				assert.Len(t, apiResp.Files, 1)

				// b...a
				req = NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/%s...%s", user2.Name, user2repo, testCase.b, testCase.a).
					AddTokenAuth(user2Ctx.Token)
				resp = MakeRequest(t, req, http.StatusOK)

				DecodeJSON(t, resp, &apiResp)

				assert.Equal(t, 0, apiResp.TotalCommits)
				assert.Empty(t, apiResp.Commits)
				assert.Empty(t, apiResp.Files)
			})
		}

		user4 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 4})
		user4Ctx := NewAPITestContext(t, user4.Name, user2repo, auth_model.AccessTokenScopeWriteRepository, auth_model.AccessTokenScopeWriteUser)

		t.Run("ForkNotFound", func(t *testing.T) {
			req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/%s...%s:%s", user2.Name, user2repo, "master", user4.Name, user2branchName).
				AddTokenAuth(user2Ctx.Token)
			resp := MakeRequest(t, req, http.StatusNotFound)
			requireErrorContains(t, resp, "user4 does not have a fork of user2/repoA and user2/repoA is not a fork of a repository from user4")
		})

		t.Run("User4ForksUser2Repository", doAPIForkRepository(user4Ctx, user2.Name))
		user4branchName := "user4branch"
		t.Run("CreateUser4RepositoryBranch", newBranchAndFile(user4Ctx, user4, user4branchName, "user4branchfilename.txt"))
		user4branch := doAPIGetBranch(user4Ctx, user4branchName)(t)
		user4tag4 := "tag4"
		t.Run("CreateUser4RepositoryTag4", doAPICreateTag(user4Ctx, user4tag4, user4branchName, "user4branchtag4"))

		t.Run("FromTheForkedRepo", func(t *testing.T) {
			// user4/repoA is a fork of user2/repoA and when evaluating
			//
			// user4/repoA/compare/master...user2:user2branch
			//
			// user2/repoA is not explicitly specified, it is implicitly the repository
			// from which user4/repoA was forked
			req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/%s...%s:%s", user4.Name, user2repo, "master", user2.Name, user2branchName).
				AddTokenAuth(user4Ctx.Token)
			resp := MakeRequest(t, req, http.StatusOK)

			var apiResp *api.Compare
			DecodeJSON(t, resp, &apiResp)

			assert.Equal(t, 1, apiResp.TotalCommits)
			assert.Len(t, apiResp.Commits, 1)
			assert.Len(t, apiResp.Files, 1)
		})

		for _, testCase := range []struct {
			name string
			a    string
			b    string
		}{
			{
				name: "Commits",
				a:    user2master.Commit.ID,
				b:    fmt.Sprintf("%s:%s", user4.Name, user4branch.Commit.ID),
			},
			{
				name: "ShortCommits",
				a:    user2master.Commit.ID[:shortCommitLength],
				b:    fmt.Sprintf("%s:%s", user4.Name, user4branch.Commit.ID[:shortCommitLength]),
			},
			{
				name: "Branches",
				a:    "master",
				b:    fmt.Sprintf("%s:%s", user4.Name, user4branchName),
			},
			{
				name: "Tags",
				a:    user2tag1,
				b:    fmt.Sprintf("%s:%s", user4.Name, user4tag4),
			},
			{
				name: "SameRepo",
				a:    "master",
				b:    fmt.Sprintf("%s:%s", user2.Name, user2branchName),
			},
		} {
			t.Run("ForkedRepo"+testCase.name, func(t *testing.T) {
				// user2/repoA is forked into user4/repoA and when evaluating
				//
				// user2/repoA/compare/a...user4:b
				//
				// user4/repoA is not explicitly specified, it is implicitly the repository
				// owned by user4 which is a fork of repoA
				req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/%s...%s", user2.Name, user2repo, testCase.a, testCase.b).
					AddTokenAuth(user2Ctx.Token)
				resp := MakeRequest(t, req, http.StatusOK)

				var apiResp *api.Compare
				DecodeJSON(t, resp, &apiResp)

				assert.Equal(t, 1, apiResp.TotalCommits)
				assert.Len(t, apiResp.Commits, 1)
				assert.Len(t, apiResp.Files, 1)
			})
		}

		t.Run("ForkUserDoesNotExist", func(t *testing.T) {
			notUser := "notauser"
			req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/master...%s:branchname", user2.Name, user2repo, notUser).
				AddTokenAuth(user2Ctx.Token)
			resp := MakeRequest(t, req, http.StatusNotFound)
			requireErrorContains(t, resp, fmt.Sprintf("the owner %s does not exist", notUser))
		})

		t.Run("HeadHasTooManyColon", func(t *testing.T) {
			req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/master...one:two:many", user2.Name, user2repo).
				AddTokenAuth(user2Ctx.Token)
			resp := MakeRequest(t, req, http.StatusNotFound)
			requireErrorContains(t, resp, fmt.Sprintf("must contain zero or one colon (:) but contains 2"))
		})

		for _, testCase := range []struct {
			what     string
			baseHead string
		}{
			{
				what:     "base",
				baseHead: "notexists...master",
			},
			{
				what:     "head",
				baseHead: "master...notexists",
			},
		} {
			t.Run("BaseHeadNotExists "+testCase.what, func(t *testing.T) {
				req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/compare/%s", user2.Name, user2repo, testCase.baseHead).
					AddTokenAuth(user2Ctx.Token)
				resp := MakeRequest(t, req, http.StatusNotFound)
				requireErrorContains(t, resp, fmt.Sprintf("could not find 'notexists' to be a commit, branch or tag in the %s", testCase.what))
			})
		}
	})
}

func TestAPICompareCommitsAccessTokenResources(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")

	// Using the compare API, will be testing that the base repo's security checks implement fine-grained access
	// controls (and baselines with all and public-only).
	testCase := func(t *testing.T, repo, token string, expectedStatus int) {
		req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/compare/master...master", repo)).AddTokenAuth(token)
		MakeRequest(t, req, expectedStatus)
	}

	t.Run("all access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		allToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)

		testCase(t, "user2/repo1", allToken, http.StatusOK)  // public user2/repo1
		testCase(t, "org3/repo3", allToken, http.StatusOK)   // private org3/repo3
		testCase(t, "user2/repo20", allToken, http.StatusOK) // private user2/repo20
	})

	t.Run("public-only access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		publicOnlyToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopePublicOnly, auth_model.AccessTokenScopeReadRepository)

		testCase(t, "user2/repo1", publicOnlyToken, http.StatusOK)        // public user2/repo1
		testCase(t, "org3/repo3", publicOnlyToken, http.StatusNotFound)   // private org3/repo3
		testCase(t, "user2/repo20", publicOnlyToken, http.StatusNotFound) // private user2/repo20
	})

	t.Run("specific repo access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		repo2OnlyToken := createFineGrainedRepoAccessToken(t, "user2",
			[]auth_model.AccessTokenScope{auth_model.AccessTokenScopeReadRepository},
			[]int64{3},
		)

		testCase(t, "user2/repo1", repo2OnlyToken, http.StatusOK)        // public user2/repo1
		testCase(t, "org3/repo3", repo2OnlyToken, http.StatusOK)         // private org3/repo3
		testCase(t, "user2/repo20", repo2OnlyToken, http.StatusNotFound) // private user2/repo20, outside of fine-grain
	})
}
