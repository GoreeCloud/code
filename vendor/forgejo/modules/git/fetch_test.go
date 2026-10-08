// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
package git

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	t.Run("SHA1", func(t *testing.T) {
		dstDir := t.TempDir()

		require.NoError(t, Clone(t.Context(), filepath.Join(testReposDir, "repo1_bare"), dstDir, CloneRepoOptions{}))

		repo, err := OpenRepository(t.Context(), dstDir)
		require.NoError(t, err)
		defer repo.Close()

		t.Run("Reference", func(t *testing.T) {
			otherRepoPath, err := filepath.Abs(filepath.Join(testReposDir, "language_stats_repo"))
			require.NoError(t, err)

			fetchedCommitID, err := repo.Fetch(otherRepoPath, "refs/heads/master")
			require.NoError(t, err)
			assert.Equal(t, "5684d0c8cfdfb17fcd59101826efc9ff54b80df4", fetchedCommitID)

			c, err := repo.getCommit(MustIDFromString(fetchedCommitID))
			require.NoError(t, err)
			assert.NotNil(t, c)
		})

		t.Run("CommitID", func(t *testing.T) {
			otherRepoPath, err := filepath.Abs(filepath.Join(testReposDir, "repo6_blame"))
			require.NoError(t, err)

			fetchedCommitID, err := repo.Fetch(otherRepoPath, "45fb6cbc12f970b04eacd5cd4165edd11c8d7376")
			require.NoError(t, err)
			assert.Equal(t, "45fb6cbc12f970b04eacd5cd4165edd11c8d7376", fetchedCommitID)

			c, err := repo.getCommit(MustIDFromString(fetchedCommitID))
			require.NoError(t, err)
			assert.NotNil(t, c)
		})

		t.Run("Invalid reference", func(t *testing.T) {
			otherRepoPath, err := filepath.Abs(filepath.Join(testReposDir, "repo6_blame"))
			require.NoError(t, err)

			fetchedCommitID, err := repo.Fetch(otherRepoPath, "refs/heads/does-not-exist")
			require.ErrorIs(t, err, ErrRemoteRefNotFound)
			assert.Empty(t, fetchedCommitID)
		})
	})

	t.Run("SHA256", func(t *testing.T) {
		skipIfSHA256NotSupported(t)

		dstDir := t.TempDir()

		require.NoError(t, Clone(t.Context(), filepath.Join(testReposDir, "repo1_bare_sha256"), dstDir, CloneRepoOptions{}))

		repo, err := OpenRepository(t.Context(), dstDir)
		require.NoError(t, err)
		defer repo.Close()

		t.Run("Reference", func(t *testing.T) {
			otherRepoPath, err := filepath.Abs(filepath.Join(testReposDir, "repo6_blame_sha256"))
			require.NoError(t, err)

			fetchedCommitID, err := repo.Fetch(otherRepoPath, "refs/heads/main")
			require.NoError(t, err)
			assert.Equal(t, "e2f5660e15159082902960af0ed74fc144921d2b0c80e069361853b3ece29ba3", fetchedCommitID)

			c, err := repo.getCommit(MustIDFromString(fetchedCommitID))
			require.NoError(t, err)
			assert.NotNil(t, c)
		})

		t.Run("CommitID", func(t *testing.T) {
			otherRepoPath, err := filepath.Abs(filepath.Join(testReposDir, "repo6_merge_sha256"))
			require.NoError(t, err)

			fetchedCommitID, err := repo.Fetch(otherRepoPath, "d2e5609f630dd8db500f5298d05d16def282412e3e66ed68cc7d0833b29129a1")
			require.NoError(t, err)
			assert.Equal(t, "d2e5609f630dd8db500f5298d05d16def282412e3e66ed68cc7d0833b29129a1", fetchedCommitID)

			c, err := repo.getCommit(MustIDFromString(fetchedCommitID))
			require.NoError(t, err)
			assert.NotNil(t, c)
		})

		t.Run("Invalid reference", func(t *testing.T) {
			otherRepoPath, err := filepath.Abs(filepath.Join(testReposDir, "repo6_blame_sha256"))
			require.NoError(t, err)

			fetchedCommitID, err := repo.Fetch(otherRepoPath, "refs/heads/does-not-exist")
			require.ErrorIs(t, err, ErrRemoteRefNotFound)
			assert.Empty(t, fetchedCommitID)
		})
	})
}
