// Copyright 2019 The Gitea Authors.
// All rights reserved.
// SPDX-License-Identifier: MIT

package pull

import (
	"testing"

	"forgejo.org/models/db"
	issues_model "forgejo.org/models/issues"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unit"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/git"
	"forgejo.org/modules/gitrepo"
	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO TestPullRequest_PushToBaseRepo

func TestPullRequest_CommitMessageTrailersPattern(t *testing.T) {
	// Not a valid trailer section
	assert.False(t, commitMessageTrailersPattern.MatchString(""))
	assert.False(t, commitMessageTrailersPattern.MatchString("No trailer."))
	assert.False(t, commitMessageTrailersPattern.MatchString("Signed-off-by: Bob <bob@example.com>\nNot a trailer due to following text."))
	assert.False(t, commitMessageTrailersPattern.MatchString("Message body not correctly separated from trailer section by empty line.\nSigned-off-by: Bob <bob@example.com>"))
	// Valid trailer section
	assert.True(t, commitMessageTrailersPattern.MatchString("Signed-off-by: Bob <bob@example.com>"))
	assert.True(t, commitMessageTrailersPattern.MatchString("Signed-off-by: Bob <bob@example.com>\nOther-Trailer: Value"))
	assert.True(t, commitMessageTrailersPattern.MatchString("Message body correctly separated from trailer section by empty line.\n\nSigned-off-by: Bob <bob@example.com>"))
	assert.True(t, commitMessageTrailersPattern.MatchString("Multiple trailers.\n\nSigned-off-by: Bob <bob@example.com>\nOther-Trailer: Value"))
	assert.True(t, commitMessageTrailersPattern.MatchString("Newline after trailer section.\n\nSigned-off-by: Bob <bob@example.com>\n"))
	assert.True(t, commitMessageTrailersPattern.MatchString("No space after colon is accepted.\n\nSigned-off-by:Bob <bob@example.com>"))
	assert.True(t, commitMessageTrailersPattern.MatchString("Additional whitespace is accepted.\n\nSigned-off-by \t :  \tBob   <bob@example.com>   "))
	assert.True(t, commitMessageTrailersPattern.MatchString("Folded value.\n\nFolded-trailer: This is\n a folded\n   trailer value\nOther-Trailer: Value"))
}

func TestPullRequest_GetDefaultMergeMessage_InternalTracker(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	pr := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 2})

	require.NoError(t, pr.LoadBaseRepo(db.DefaultContext))
	gitRepo, err := gitrepo.OpenRepository(git.DefaultContext, pr.BaseRepo)
	require.NoError(t, err)
	defer gitRepo.Close()

	mergeMessage, body, err := GetDefaultMergeMessage(db.DefaultContext, gitRepo, pr, "")
	require.NoError(t, err)
	assert.Equal(t, "Merge pull request 'issue3' (#3) from branch2 into master", mergeMessage)
	assert.Equal(t, "Reviewed-on: https://try.gitea.io/user2/repo1/pulls/3\n", body)

	pr.BaseRepoID = 1
	pr.HeadRepoID = 2
	mergeMessage, _, err = GetDefaultMergeMessage(db.DefaultContext, gitRepo, pr, "")
	require.NoError(t, err)
	assert.Equal(t, "Merge pull request 'issue3' (#3) from user2/repo1:branch2 into master", mergeMessage)

	setting.AppURL = "https://example.org/suburl/"
	setting.AppSubURL = "/suburl"
	_, body, err = GetDefaultMergeMessage(db.DefaultContext, gitRepo, pr, "")
	require.NoError(t, err)
	assert.Equal(t, "Reviewed-on: https://example.org/suburl/user2/repo1/pulls/3\n", body)
}

func TestPullRequest_GetDefaultMergeMessage_ExternalTracker(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	externalTracker := repo_model.RepoUnit{
		Type: unit.TypeExternalTracker,
		Config: &repo_model.ExternalTrackerConfig{
			ExternalTrackerFormat: "https://someurl.com/{user}/{repo}/{issue}",
		},
	}
	baseRepo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	baseRepo.Units = []*repo_model.RepoUnit{&externalTracker}

	pr := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 2, BaseRepo: baseRepo})

	require.NoError(t, pr.LoadBaseRepo(db.DefaultContext))
	gitRepo, err := gitrepo.OpenRepository(git.DefaultContext, pr.BaseRepo)
	require.NoError(t, err)
	defer gitRepo.Close()

	mergeMessage, _, err := GetDefaultMergeMessage(db.DefaultContext, gitRepo, pr, "")
	require.NoError(t, err)

	assert.Equal(t, "Merge pull request 'issue3' (!3) from branch2 into master", mergeMessage)

	pr.BaseRepoID = 1
	pr.HeadRepoID = 2
	pr.BaseRepo = nil
	pr.HeadRepo = nil
	mergeMessage, _, err = GetDefaultMergeMessage(db.DefaultContext, gitRepo, pr, "")
	require.NoError(t, err)

	assert.Equal(t, "Merge pull request 'issue3' (#3) from user2/repo2:branch2 into master", mergeMessage)
}

func TestPullRequest_GetDefaultMergeMessage_GlobalTemplate(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	pr := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 2})

	require.NoError(t, pr.LoadBaseRepo(t.Context()))
	gitRepo, err := gitrepo.OpenRepository(t.Context(), pr.BaseRepo)
	require.NoError(t, err)
	defer gitRepo.Close()

	templateRepo, err := git.OpenRepository(t.Context(), "./../../modules/git/tests/repos/templates_repo")
	require.NoError(t, err)
	defer templateRepo.Close()

	mergeMessageTemplates[repo_model.MergeStyleMerge] = "${PullRequestTitle} (${PullRequestReference})\n${PullRequestDescription}"

	// Check template is used for Merge...
	mergeMessage, body, err := GetDefaultMergeMessage(t.Context(), gitRepo, pr, repo_model.MergeStyleMerge)
	require.NoError(t, err)

	assert.Equal(t, "issue3 (#3)", mergeMessage)
	assert.Equal(t, "content for the third issue", body)

	// ...but not for RebaseMerge
	mergeMessage, _, err = GetDefaultMergeMessage(t.Context(), gitRepo, pr, repo_model.MergeStyleRebaseMerge)
	require.NoError(t, err)

	assert.Equal(t, "Merge pull request 'issue3' (#3) from branch2 into master", mergeMessage)

	// ...and that custom Merge template takes priority
	mergeMessage, body, err = GetDefaultMergeMessage(t.Context(), templateRepo, pr, repo_model.MergeStyleMerge)
	require.NoError(t, err)

	assert.Equal(t, "Default merge message template", mergeMessage)
	assert.Equal(t, "\nThis line was read from .forgejo/default_merge_message/MERGE_TEMPLATE.md", body)
}
