// Copyright 2024 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"os"
	"path"
	"testing"

	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_calcSync(t *testing.T) {
	gitTags := []*git.Tag{
		/*{
			Name: "v0.1.0-beta", //deleted tag
			Object: git.MustIDFromString(""),
		},
		{
			Name: "v0.1.1-beta", //deleted tag but release should not be deleted because it's a release
			Object: git.MustIDFromString(""),
		},
		*/
		{
			Name:   "v1.0.0", // keep as before
			Object: git.MustIDFromString("1006e6e13c73ad3d9e2d5682ad266b5016523485"),
		},
		{
			Name:   "v1.1.0", // retagged with new commit id
			Object: git.MustIDFromString("bbdb7df30248e7d4a26a909c8d2598a152e13868"),
		},
		{
			Name:   "v1.2.0", // new tag
			Object: git.MustIDFromString("a5147145e2f24d89fd6d2a87826384cc1d253267"),
		},
	}

	dbReleases := []*shortRelease{
		{
			ID:      1,
			TagName: "v0.1.0-beta",
			Sha1:    "244758d7da8dd1d9e0727e8cb7704ed4ba9a17c3",
			IsTag:   true,
		},
		{
			ID:      2,
			TagName: "v0.1.1-beta",
			Sha1:    "244758d7da8dd1d9e0727e8cb7704ed4ba9a17c3",
			IsTag:   false,
		},
		{
			ID:      3,
			TagName: "v1.0.0",
			Sha1:    "1006e6e13c73ad3d9e2d5682ad266b5016523485",
		},
		{
			ID:      4,
			TagName: "v1.1.0",
			Sha1:    "53ab18dcecf4152b58328d1f47429510eb414d50",
		},
	}

	inserts, deletes, updates := calcSync(gitTags, dbReleases)
	if assert.Len(t, inserts, 1, "inserts") {
		assert.Equal(t, *gitTags[2], *inserts[0], "inserts equal")
	}

	if assert.Len(t, deletes, 1, "deletes") {
		assert.EqualValues(t, 1, deletes[0], "deletes equal")
	}

	if assert.Len(t, updates, 1, "updates") {
		assert.Equal(t, *gitTags[1], *updates[0], "updates equal")
	}
}

func TestSyncReleasesWithTags(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// Can be any repository that doesn't have the git tag releases.
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})

	t.Run("SHA1", func(t *testing.T) {
		tmpDir := t.TempDir()

		require.NoError(t, git.InitRepository(t.Context(), tmpDir, false, git.Sha1ObjectFormat.Name()))
		gitRepo, err := git.OpenRepository(t.Context(), tmpDir)
		require.NoError(t, err)
		defer gitRepo.Close()

		require.NoError(t, os.WriteFile(path.Join(tmpDir, "README.md"), []byte("testing the testing"), 0o666))
		require.NoError(t, git.AddChanges(tmpDir, true))
		require.NoError(t, git.CommitChanges(tmpDir, git.CommitChangesOptions{Message: "Add README"}))
		require.NoError(t, gitRepo.CreateAnnotatedTag("v1.0.0", "First release \\o/", "HEAD"))

		require.NoError(t, SyncReleasesWithTags(t.Context(), repo, gitRepo))

		release := unittest.AssertExistsAndLoadBean(t, &repo_model.Release{RepoID: repo.ID, TagName: "v1.0.0"})
		assert.Equal(t, "First release \\o/\n", release.Note)
		assert.True(t, release.IsTag)
		assert.EqualValues(t, 1, release.NumCommits)
	})

	t.Run("SHA256", func(t *testing.T) {
		if !git.SupportHashSha256 {
			t.Skip("skipping because installed Git version doesn't support SHA256")
		}

		tmpDir := t.TempDir()

		require.NoError(t, git.InitRepository(t.Context(), tmpDir, false, git.Sha256ObjectFormat.Name()))
		gitRepo, err := git.OpenRepository(t.Context(), tmpDir)
		require.NoError(t, err)
		defer gitRepo.Close()

		require.NoError(t, os.WriteFile(path.Join(tmpDir, "README.md"), []byte("testing the testing"), 0o666))
		require.NoError(t, git.AddChanges(tmpDir, true))
		require.NoError(t, git.CommitChanges(tmpDir, git.CommitChangesOptions{Message: "Add README"}))
		require.NoError(t, gitRepo.CreateAnnotatedTag("v2.0.0", "Second release \\o/", "HEAD"))

		require.NoError(t, SyncReleasesWithTags(t.Context(), repo, gitRepo))

		release := unittest.AssertExistsAndLoadBean(t, &repo_model.Release{RepoID: repo.ID, TagName: "v2.0.0"})
		assert.Equal(t, "Second release \\o/\n", release.Note)
		assert.True(t, release.IsTag)
		assert.EqualValues(t, 1, release.NumCommits)
	})
}
