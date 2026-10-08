// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	auth_model "forgejo.org/models/auth"
	issues_model "forgejo.org/models/issues"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/gitrepo"
	api "forgejo.org/modules/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullMergeInstruction(t *testing.T) {
	onApplicationRun(t, func(t *testing.T, _ *url.URL) {
		session := loginUser(t, "user1")
		testRepoFork(t, session, "user2", "repo1", "user1", "repo1")
		testEditFileToNewBranch(t, session, "user1", "repo1", "master", "conflict", "README.md", "Hello, World (Edited Once)\n")
		testEditFileToNewBranch(t, session, "user1", "repo1", "master", "base", "README.md", "Hello, World (Edited Twice)\n")

		// Use API to create a conflicting PR, mirroring TestCantMergeConflict
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)
		req := NewRequestWithJSON(t, http.MethodPost, fmt.Sprintf("/api/v1/repos/%s/%s/pulls", "user1", "repo1"), &api.CreatePullRequestOption{
			Head:  "conflict",
			Base:  "base",
			Title: "create a conflicting pr",
		}).AddTokenAuth(token)
		resp := session.MakeRequest(t, req, http.StatusCreated)

		var pr api.PullRequest
		DecodeJSON(t, resp, &pr)

		user1 := unittest.AssertExistsAndLoadBean(t, &user_model.User{
			Name: "user1",
		})
		repo1 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{
			OwnerID: user1.ID,
			Name:    "repo1",
		})

		// Assert that the PR exists, is open, and is not mergeable (conflicted)
		prLoaded := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{
			ID:         pr.ID,
			HeadRepoID: repo1.ID,
			BaseRepoID: repo1.ID,
			HeadBranch: "conflict",
			BaseBranch: "base",
		}, "status = 0")
		assert.False(t, prLoaded.Mergeable(t.Context()), "PR should be marked as conflicted")

		gitRepo, err := gitrepo.OpenRepository(t.Context(), repo1)
		require.NoError(t, err)
		defer gitRepo.Close()

		// Visit the PR page and check for the manual merge helper
		req = NewRequest(t, "GET", fmt.Sprintf("/user1/repo1/pulls/%d", pr.Index))
		resp = session.MakeRequest(t, req, http.StatusOK)

		htmlDoc := NewHTMLParser(t, resp.Body)

		// Check for "View command line instructions"
		summary := htmlDoc.doc.Find("details.collapsible summary").Text()
		assert.Contains(t, summary, "View command line instructions")

		// Check for "Manual merge helper"
		helperTitle := htmlDoc.doc.Find("details.collapsible h3").Text()
		assert.Contains(t, helperTitle, "Manual merge helper")

		// Check for the description
		helperDesc := htmlDoc.doc.Find("details.collapsible p").Text()
		assert.Contains(t, helperDesc, "Use this merge commit message when completing the merge manually.")
	})
}
