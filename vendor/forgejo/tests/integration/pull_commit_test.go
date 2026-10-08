// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/git"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/test"
	pull_service "forgejo.org/services/pull"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPullCommits(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	session := loginUser(t, "user5")
	req := NewRequest(t, "GET", "/user2/repo1/pulls/3/commits/list")
	resp := session.MakeRequest(t, req, http.StatusOK)

	var pullCommitList struct {
		Commits             []pull_service.CommitInfo `json:"commits"`
		LastReviewCommitSha string                    `json:"last_review_commit_sha"`
	}
	DecodeJSON(t, resp, &pullCommitList)

	if assert.Len(t, pullCommitList.Commits, 2) {
		assert.Equal(t, "5f22f7d0d95d614d25a5b68592adb345a4b5c7fd", pullCommitList.Commits[0].ID)
		assert.Equal(t, "4a357436d925b5c974181ff12a994538ddc5a269", pullCommitList.Commits[1].ID)
	}
	assert.Equal(t, "4a357436d925b5c974181ff12a994538ddc5a269", pullCommitList.LastReviewCommitSha)
}

func TestPullCommitLinks(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/pulls/3/commits")
	resp := MakeRequest(t, req, http.StatusOK)

	htmlDoc := NewHTMLParser(t, resp.Body)

	commitSha := htmlDoc.Find(".commit-list td.sha a.sha.label").First()
	commitShaHref, _ := commitSha.Attr("href")
	assert.Equal(t, "/user2/repo1/pulls/3/commits/5f22f7d0d95d614d25a5b68592adb345a4b5c7fd", commitShaHref)

	commitLink := htmlDoc.Find(".commit-list td.message a").First()
	commitLinkHref, _ := commitLink.Attr("href")
	assert.Equal(t, "/user2/repo1/pulls/3/commits/5f22f7d0d95d614d25a5b68592adb345a4b5c7fd", commitLinkHref)
}

func TestPullCommitLinksSHA256(t *testing.T) {
	if !git.SupportHashSha256 {
		t.Skip("skipping because installed Git version doesn't support SHA256")
		return
	}

	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo256/pulls/1/commits")
	resp := MakeRequest(t, req, http.StatusOK)

	htmlDoc := NewHTMLParser(t, resp.Body)

	commitSha := htmlDoc.Find(".commit-list td.sha a.sha.label").First()
	commitShaHref, commitShaOk := commitSha.Attr("href")
	assert.True(t, commitShaOk)
	assert.Equal(t, "/user2/repo256/pulls/1/commits/004581b3bb63754502364664021404490ee747ce58e98d27c046f2e46f5f2f55", commitShaHref)

	commitLink := htmlDoc.Find(".commit-list td.message a").First()
	commitLinkHref, commitLinkOk := commitLink.Attr("href")
	assert.True(t, commitLinkOk)
	assert.Equal(t, "/user2/repo256/pulls/1/commits/004581b3bb63754502364664021404490ee747ce58e98d27c046f2e46f5f2f55", commitLinkHref)

	commitReq := NewRequest(t, "GET", commitShaHref)
	MakeRequest(t, commitReq, http.StatusOK)
}

func TestPullCommitSignature(t *testing.T) {
	t.Cleanup(func() {
		// Cannot use t.Context(), it is in the done state.
		require.NoError(t, git.InitFull(context.Background()))
	})

	defer test.MockVariableValue(&setting.Repository.Signing.SigningName, "UwU")()
	defer test.MockVariableValue(&setting.Repository.Signing.SigningEmail, "fox@example.com")()
	defer test.MockVariableValue(&setting.Repository.Signing.CRUDActions, []string{"always"})()
	defer test.MockVariableValue(&setting.Repository.Signing.InitialCommit, []string{"always"})()

	filePath := "signed.txt"
	fromBranch := "master"
	toBranch := "branch-signed"

	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		// Use a new GNUPGPHOME to avoid messing with the existing GPG keyring.
		tmpDir := t.TempDir()
		require.NoError(t, os.Chmod(tmpDir, 0o700))
		t.Setenv("GNUPGHOME", tmpDir)

		rootKeyPair, err := importTestingKey()
		require.NoError(t, err)
		defer test.MockVariableValue(&setting.Repository.Signing.SigningKey, rootKeyPair.PrimaryKey.KeyIdShortString())()
		defer test.MockVariableValue(&setting.Repository.Signing.Format, "openpgp")()

		// Ensure the git config is updated with the new signing format.
		require.NoError(t, git.InitFull(t.Context()))

		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
		testCtx := NewAPITestContext(t, user.Name, "pull-request-commit-header-signed", auth_model.AccessTokenScopeWriteRepository, auth_model.AccessTokenScopeWriteUser)
		u.Path = testCtx.GitPath()

		t.Run("Create repository", doAPICreateRepository(testCtx, nil, git.Sha1ObjectFormat))

		t.Run("Create commit", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			options := &api.CreateFileOptions{
				FileOptions: api.FileOptions{
					BranchName:    fromBranch,
					NewBranchName: toBranch,
					Message:       fmt.Sprintf("from:%s to:%s path:%s", fromBranch, toBranch, filePath),
					Author: api.Identity{
						Name:  user.FullName,
						Email: user.Email,
					},
					Committer: api.Identity{
						Name:  user.FullName,
						Email: user.Email,
					},
				},
				ContentBase64: base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "This is new text for %s", filePath)),
			}

			req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/contents/%s", testCtx.Username, testCtx.Reponame, filePath), &options).
				AddTokenAuth(testCtx.Token)
			resp := testCtx.Session.MakeRequest(t, req, http.StatusCreated)

			var contents api.FileResponse
			DecodeJSON(t, resp, &contents)

			assert.True(t, contents.Verification.Verified)
		})

		t.Run("Create pull request", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			pr, err := doAPICreatePullRequest(testCtx, testCtx.Username, testCtx.Reponame, fromBranch, toBranch)(t)
			require.NoError(t, err)

			req := NewRequest(t, "GET", fmt.Sprintf("/%s/%s/pulls/%d/commits/%s", testCtx.Username, testCtx.Reponame, pr.Index, pr.Head.Sha))
			resp := testCtx.Session.MakeRequest(t, req, http.StatusOK)

			htmlDoc := NewHTMLParser(t, resp.Body)
			htmlDoc.AssertElement(t, "#diff-commit-header .signature-row.message.isSigned.isVerified", true)
		})
	})
}
