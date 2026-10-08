// Copyright 2017 Gitea. All rights reserved.
// SPDX-License-Identifier: MIT

package git_test

import (
	"testing"
	"time"

	"forgejo.org/models/db"
	git_model "forgejo.org/models/git"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/git"
	"forgejo.org/modules/gitrepo"
	"forgejo.org/modules/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommitStatuses(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo1 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})

	sha1 := "1234123412341234123412341234123412341234"

	statuses, maxResults, err := db.FindAndCount[git_model.CommitStatus](db.DefaultContext, &git_model.CommitStatusOptions{
		ListOptions: db.ListOptions{Page: 1, PageSize: 50},
		RepoID:      repo1.ID,
		SHA:         sha1,
	})
	require.NoError(t, err)
	assert.EqualValues(t, 6, maxResults)
	assert.Len(t, statuses, 6)

	assert.Equal(t, "ci/awesomeness", statuses[0].Context)
	assert.Equal(t, structs.CommitStatusPending, statuses[0].State)
	assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/statuses/1234123412341234123412341234123412341234", statuses[0].APIURL(db.DefaultContext))

	assert.Equal(t, "cov/awesomeness", statuses[1].Context)
	assert.Equal(t, structs.CommitStatusWarning, statuses[1].State)
	assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/statuses/1234123412341234123412341234123412341234", statuses[1].APIURL(db.DefaultContext))

	assert.Equal(t, "cov/awesomeness", statuses[2].Context)
	assert.Equal(t, structs.CommitStatusSuccess, statuses[2].State)
	assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/statuses/1234123412341234123412341234123412341234", statuses[2].APIURL(db.DefaultContext))

	assert.Equal(t, "ci/awesomeness", statuses[3].Context)
	assert.Equal(t, structs.CommitStatusFailure, statuses[3].State)
	assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/statuses/1234123412341234123412341234123412341234", statuses[3].APIURL(db.DefaultContext))

	assert.Equal(t, "deploy/awesomeness", statuses[4].Context)
	assert.Equal(t, structs.CommitStatusError, statuses[4].State)
	assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/statuses/1234123412341234123412341234123412341234", statuses[4].APIURL(db.DefaultContext))

	assert.Equal(t, "deploy/awesomeness", statuses[5].Context)
	assert.Equal(t, structs.CommitStatusPending, statuses[5].State)
	assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/statuses/1234123412341234123412341234123412341234", statuses[5].APIURL(db.DefaultContext))

	statuses, maxResults, err = db.FindAndCount[git_model.CommitStatus](db.DefaultContext, &git_model.CommitStatusOptions{
		ListOptions: db.ListOptions{Page: 2, PageSize: 50},
		RepoID:      repo1.ID,
		SHA:         sha1,
	})
	require.NoError(t, err)
	assert.EqualValues(t, 6, maxResults)
	assert.Empty(t, statuses)
}

func Test_CalcCommitStatus(t *testing.T) {
	kases := []struct {
		statuses []*git_model.CommitStatus
		expected *git_model.CommitStatus
	}{
		{
			statuses: []*git_model.CommitStatus{
				{
					State: structs.CommitStatusPending,
				},
			},
			expected: &git_model.CommitStatus{
				State: structs.CommitStatusPending,
			},
		},
		{
			statuses: []*git_model.CommitStatus{
				{
					State: structs.CommitStatusSuccess,
				},
				{
					State: structs.CommitStatusPending,
				},
			},
			expected: &git_model.CommitStatus{
				State: structs.CommitStatusPending,
			},
		},
		{
			statuses: []*git_model.CommitStatus{
				{
					State: structs.CommitStatusSuccess,
				},
				{
					State: structs.CommitStatusPending,
				},
				{
					State: structs.CommitStatusSuccess,
				},
			},
			expected: &git_model.CommitStatus{
				State: structs.CommitStatusPending,
			},
		},
		{
			statuses: []*git_model.CommitStatus{
				{
					State: structs.CommitStatusError,
				},
				{
					State: structs.CommitStatusPending,
				},
				{
					State: structs.CommitStatusSuccess,
				},
			},
			expected: &git_model.CommitStatus{
				State: structs.CommitStatusError,
			},
		},
		{
			statuses: []*git_model.CommitStatus{
				{
					State: structs.CommitStatusWarning,
				},
				{
					State: structs.CommitStatusPending,
				},
				{
					State: structs.CommitStatusSuccess,
				},
			},
			expected: &git_model.CommitStatus{
				State: structs.CommitStatusWarning,
			},
		},
		{
			statuses: []*git_model.CommitStatus{
				{
					State: structs.CommitStatusSuccess,
					ID:    1,
				},
				{
					State: structs.CommitStatusSuccess,
					ID:    2,
				},
				{
					State: structs.CommitStatusSuccess,
					ID:    3,
				},
			},
			expected: &git_model.CommitStatus{
				State: structs.CommitStatusSuccess,
				ID:    3,
			},
		},
		{
			statuses: []*git_model.CommitStatus{
				{
					State: structs.CommitStatusFailure,
				},
				{
					State: structs.CommitStatusError,
				},
				{
					State: structs.CommitStatusWarning,
				},
			},
			expected: &git_model.CommitStatus{
				State: structs.CommitStatusError,
			},
		},
		{
			statuses: []*git_model.CommitStatus{},
			expected: nil,
		},
	}

	for _, kase := range kases {
		assert.Equal(t, kase.expected, git_model.CalcCommitStatus(kase.statuses))
	}
}

func TestFindRepoRecentCommitStatusContexts(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo2 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 2})
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	gitRepo, err := gitrepo.OpenRepository(git.DefaultContext, repo2)
	require.NoError(t, err)
	defer gitRepo.Close()

	commit, err := gitRepo.GetBranchCommit(repo2.DefaultBranch)
	require.NoError(t, err)

	defer func() {
		_, err := db.DeleteByBean(db.DefaultContext, &git_model.CommitStatus{
			RepoID:    repo2.ID,
			CreatorID: user2.ID,
			SHA:       commit.ID.String(),
		})
		require.NoError(t, err)
	}()

	err = git_model.NewCommitStatus(db.DefaultContext, git_model.NewCommitStatusOptions{
		Repo:    repo2,
		Creator: user2,
		SHA:     commit.ID,
		CommitStatus: &git_model.CommitStatus{
			State:     structs.CommitStatusFailure,
			TargetURL: "https://example.com/tests/",
			Context:   "compliance/lint-backend",
		},
	})
	require.NoError(t, err)

	err = git_model.NewCommitStatus(db.DefaultContext, git_model.NewCommitStatusOptions{
		Repo:    repo2,
		Creator: user2,
		SHA:     commit.ID,
		CommitStatus: &git_model.CommitStatus{
			State:     structs.CommitStatusSuccess,
			TargetURL: "https://example.com/tests/",
			Context:   "compliance/lint-backend",
		},
	})
	require.NoError(t, err)

	contexts, err := git_model.FindRepoRecentCommitStatusContexts(db.DefaultContext, repo2.ID, time.Hour)
	require.NoError(t, err)
	if assert.Len(t, contexts, 1) {
		assert.Equal(t, "compliance/lint-backend", contexts[0])
	}
}

func TestCleanupCommitStatus(t *testing.T) {
	defer unittest.OverrideFixtures("models/git/TestCleanupCommitStatus")()
	require.NoError(t, unittest.PrepareTestDatabase())

	// No changes after a dry run:
	originalCount := unittest.GetCount(t, &git_model.CommitStatus{})
	err := git_model.CleanupCommitStatus(t.Context(), 100, 100, true)
	require.NoError(t, err)
	countAfterDryRun := unittest.GetCount(t, &git_model.CommitStatus{})
	assert.Equal(t, originalCount, countAfterDryRun)

	// Perform actual cleanup
	err = git_model.CleanupCommitStatus(t.Context(), 100, 100, false)
	require.NoError(t, err)

	// Varying descriptions
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 10})
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 11})

	// Varying state
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 12})
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 13})

	// Varying context
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 14})
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 15})

	// Varying sha
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 16})
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 17})

	// Varying repo ID
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 18})
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 19})

	// Expected to remain or be removed from cleanup of fixture data:
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 20})
	unittest.AssertNotExistsBean(t, &git_model.CommitStatus{ID: 21})
	unittest.AssertNotExistsBean(t, &git_model.CommitStatus{ID: 22})
	unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{ID: 23})
	unittest.AssertNotExistsBean(t, &git_model.CommitStatus{ID: 24})
}
