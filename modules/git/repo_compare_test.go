// Copyright 2018 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFormatPatch(t *testing.T) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	clonedPath, err := cloneRepo(t, bareRepo1Path)
	if err != nil {
		require.NoError(t, err)
		return
	}

	repo, err := openRepositoryWithDefaultContext(clonedPath)
	if err != nil {
		require.NoError(t, err)
		return
	}
	defer repo.Close()

	rd := &bytes.Buffer{}
	err = repo.GetPatch("8d92fc95^", "8d92fc95", rd)
	if err != nil {
		require.NoError(t, err)
		return
	}

	patchb, err := io.ReadAll(rd)
	if err != nil {
		require.NoError(t, err)
		return
	}

	patch := string(patchb)
	assert.Regexp(t, "^From 8d92fc95", patch)
	assert.Contains(t, patch, "Subject: [PATCH] Add file2.txt")
}

func TestReadPatch(t *testing.T) {
	// Ensure we can read the patch files
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	repo, err := openRepositoryWithDefaultContext(bareRepo1Path)
	if err != nil {
		require.NoError(t, err)
		return
	}
	defer repo.Close()
	// This patch doesn't exist
	noFile, err := repo.ReadPatchCommit(0)
	require.Error(t, err)

	// This patch is an empty one (sometimes it's a 404)
	noCommit, err := repo.ReadPatchCommit(1)
	require.Error(t, err)

	// This patch is legit and should return a commit
	oldCommit, err := repo.ReadPatchCommit(2)
	if err != nil {
		require.NoError(t, err)
		return
	}

	assert.Empty(t, noFile)
	assert.Empty(t, noCommit)
	assert.Len(t, oldCommit, 40)
	assert.Equal(t, "6e8e2a6f9efd71dbe6917816343ed8415ad696c3", oldCommit)
}

func TestReadWritePullHead(t *testing.T) {
	// Ensure we can write SHA1 head corresponding to PR and open them
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")

	// As we are writing we should clone the repository first
	clonedPath, err := cloneRepo(t, bareRepo1Path)
	if err != nil {
		require.NoError(t, err)
		return
	}

	repo, err := openRepositoryWithDefaultContext(clonedPath)
	if err != nil {
		require.NoError(t, err)
		return
	}
	defer repo.Close()

	// Try to open non-existing Pull
	_, err = repo.GetRefCommitID(PullPrefix + "0/head")
	require.Error(t, err)

	// Write a fake sha1 with only 40 zeros
	newCommit := "feaf4ba6bc635fec442f46ddd4512416ec43c2c2"
	err = repo.SetReference(PullPrefix+"1/head", newCommit)
	if err != nil {
		require.NoError(t, err)
		return
	}

	// Read the file created
	headContents, err := repo.GetRefCommitID(PullPrefix + "1/head")
	if err != nil {
		require.NoError(t, err)
		return
	}

	assert.Len(t, headContents, 40)
	assert.Equal(t, newCommit, headContents)

	// Remove file after the test
	err = repo.RemoveReference(PullPrefix + "1/head")
	require.NoError(t, err)
}

func TestGetCommitFilesChanged(t *testing.T) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	repo, err := openRepositoryWithDefaultContext(bareRepo1Path)
	require.NoError(t, err)
	defer repo.Close()

	objectFormat, err := repo.GetObjectFormat()
	require.NoError(t, err)

	testCases := []struct {
		base, head string
		files      []string
	}{
		{
			objectFormat.EmptyObjectID().String(),
			"95bb4d39648ee7e325106df01a621c530863a653",
			[]string{"file1.txt"},
		},
		{
			objectFormat.EmptyObjectID().String(),
			"8d92fc957a4d7cfd98bc375f0b7bb189a0d6c9f2",
			[]string{"file2.txt"},
		},
		{
			"95bb4d39648ee7e325106df01a621c530863a653",
			"8d92fc957a4d7cfd98bc375f0b7bb189a0d6c9f2",
			[]string{"file2.txt"},
		},
		{
			objectFormat.EmptyTree().String(),
			"8d92fc957a4d7cfd98bc375f0b7bb189a0d6c9f2",
			[]string{"file1.txt", "file2.txt"},
		},
	}

	for _, tc := range testCases {
		changedFiles, err := repo.GetFilesChangedBetween(tc.base, tc.head)
		require.NoError(t, err)
		assert.ElementsMatch(t, tc.files, changedFiles)
	}
}

func TestGetCommitShortStat(t *testing.T) {
	t.Run("repo1_bare", func(t *testing.T) {
		repo, err := openRepositoryWithDefaultContext(filepath.Join(testReposDir, "repo1_bare"))
		if err != nil {
			require.NoError(t, err)
			return
		}
		defer repo.Close()

		numFiles, totalAddition, totalDeletions, err := repo.GetCommitShortStat("ce064814f4a0d337b333e646ece456cd39fab612")
		require.NoError(t, err)
		assert.Equal(t, 0, numFiles)
		assert.Equal(t, 0, totalAddition)
		assert.Equal(t, 0, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("feaf4ba6bc635fec442f46ddd4512416ec43c2c2")
		require.NoError(t, err)
		assert.Equal(t, 0, numFiles)
		assert.Equal(t, 0, totalAddition)
		assert.Equal(t, 0, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("37991dec2c8e592043f47155ce4808d4580f9123")
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.Equal(t, 1, totalAddition)
		assert.Equal(t, 0, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("6fbd69e9823458e6c4a2fc5c0f6bc022b2f2acd1")
		require.NoError(t, err)
		assert.Equal(t, 2, numFiles)
		assert.Equal(t, 2, totalAddition)
		assert.Equal(t, 0, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("8006ff9adbf0cb94da7dad9e537e53817f9fa5c0")
		require.NoError(t, err)
		assert.Equal(t, 2, numFiles)
		assert.Equal(t, 2, totalAddition)
		assert.Equal(t, 0, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("8d92fc957a4d7cfd98bc375f0b7bb189a0d6c9f2")
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.Equal(t, 1, totalAddition)
		assert.Equal(t, 0, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("95bb4d39648ee7e325106df01a621c530863a653")
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.Equal(t, 1, totalAddition)
		assert.Equal(t, 0, totalDeletions)
	})

	t.Run("repo6_blame_sha256", func(t *testing.T) {
		repo, err := openRepositoryWithDefaultContext(filepath.Join(testReposDir, "repo6_blame_sha256"))
		if err != nil {
			require.NoError(t, err)
			return
		}
		defer repo.Close()

		numFiles, totalAddition, totalDeletions, err := repo.GetCommitShortStat("e2f5660e15159082902960af0ed74fc144921d2b0c80e069361853b3ece29ba3")
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.Equal(t, 1, totalAddition)
		assert.Equal(t, 0, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("9347b0198cd1f25017579b79d0938fa89dba34ad2514f0dd92f6bc975ed1a2fe")
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.Equal(t, 1, totalAddition)
		assert.Equal(t, 1, totalDeletions)

		numFiles, totalAddition, totalDeletions, err = repo.GetCommitShortStat("ab2b57a4fa476fb2edb74dafa577caf918561abbaa8fba0c8dc63c412e17a7cc")
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.Equal(t, 6, totalAddition)
		assert.Equal(t, 0, totalDeletions)
	})

	t.Run("Renames", func(t *testing.T) {
		repo, err := OpenRepository(t.Context(), filepath.Join(testReposDir, "renames"))
		require.NoError(t, err)
		defer repo.Close()

		numFiles, totalAddition, totalDeletions, err := repo.GetCommitShortStat("f667f3a24223414e3bfbe01ab6e445c703ab8e25")
		require.NoError(t, err)
		assert.Equal(t, 1, numFiles)
		assert.Zero(t, totalAddition)
		assert.Zero(t, totalDeletions)
	})
}

func TestGetShortStat(t *testing.T) {
	// https://github.com/git/git/blob/60f3f52f17cceefa5299709b189ce6fe2d181e7b/t/t4068-diff-symmetric-merge-base.sh#L10-L23
	repo, err := OpenRepository(t.Context(), filepath.Join(testReposDir, "symmetric_repo"))
	require.NoError(t, err)
	defer repo.Close()

	t.Run("Normal", func(t *testing.T) {
		t.Run("Via merge base", func(t *testing.T) {
			numFiles, totalAdditions, totalDeletions, err := repo.GetShortStat("br2", "main", true)
			require.NoError(t, err)
			assert.Equal(t, 1, numFiles)
			assert.Equal(t, 1, totalAdditions)
			assert.Zero(t, totalDeletions)

			numFiles, totalAdditions, totalDeletions, err = repo.GetShortStat("main", "br2", true)
			require.NoError(t, err)
			assert.Equal(t, 1, numFiles)
			assert.Equal(t, 1, totalAdditions)
			assert.Zero(t, totalDeletions)
		})

		t.Run("Direct compare", func(t *testing.T) {
			numFiles, totalAdditions, totalDeletions, err := repo.GetShortStat("main", "br2", false)
			require.NoError(t, err)
			assert.Equal(t, 2, numFiles)
			assert.Equal(t, 1, totalAdditions)
			assert.Equal(t, 1, totalDeletions)

			numFiles, totalAdditions, totalDeletions, err = repo.GetShortStat("main", "br3", false)
			require.NoError(t, err)
			assert.Equal(t, 1, numFiles)
			assert.Equal(t, 1, totalAdditions)
			assert.Zero(t, totalDeletions)

			numFiles, totalAdditions, totalDeletions, err = repo.GetShortStat("br3", "main", false)
			require.NoError(t, err)
			assert.Equal(t, 1, numFiles)
			assert.Zero(t, totalAdditions)
			assert.Equal(t, 1, totalDeletions)
		})
	})

	t.Run("No merge base", func(t *testing.T) {
		numFiles, totalAdditions, totalDeletions, err := repo.GetShortStat("main", "br3", true)
		require.ErrorIs(t, err, ErrNoMergebaseFound)
		assert.Zero(t, numFiles)
		assert.Zero(t, totalAdditions)
		assert.Zero(t, totalDeletions)
	})

	t.Run("Multiple merge base", func(t *testing.T) {
		numFiles, totalAdditions, totalDeletions, err := repo.GetShortStat("main", "br1", true)
		require.ErrorIs(t, err, ErrMultipleMergebasesFound)
		assert.Zero(t, numFiles)
		assert.Zero(t, totalAdditions)
		assert.Zero(t, totalDeletions)
	})

	t.Run("Renames", func(t *testing.T) {
		repo, err := OpenRepository(t.Context(), filepath.Join(testReposDir, "renames"))
		require.NoError(t, err)
		defer repo.Close()

		t.Run("Only rename", func(t *testing.T) {
			numFiles, totalAdditions, totalDeletions, err := repo.GetShortStat("bc40f00489096a7d4090a609a6572f528e1acb76", "f667f3a24223414e3bfbe01ab6e445c703ab8e25", true)
			require.NoError(t, err)
			assert.Equal(t, 1, numFiles)
			assert.Zero(t, totalAdditions)
			assert.Zero(t, totalDeletions)
		})

		t.Run("Too much diverged", func(t *testing.T) {
			numFiles, totalAdditions, totalDeletions, err := repo.GetShortStat("bc40f00489096a7d4090a609a6572f528e1acb76", "acdee217ada3fea6e503acfb969724cc799fc516", true)
			require.NoError(t, err)
			assert.Equal(t, 2, numFiles)
			assert.Equal(t, 3, totalAdditions)
			assert.Equal(t, 1, totalDeletions)
		})
	})
}

func TestGetMergeBaseSimple(t *testing.T) {
	repo, err := OpenRepository(t.Context(), filepath.Join(testReposDir, "symmetric_repo"))
	require.NoError(t, err)

	defer repo.Close()

	t.Run("Normal", func(t *testing.T) {
		mergebase, err := repo.GetMergeBaseSimple("main", "br2")
		require.NoError(t, err)
		assert.Equal(t, "9d36f18c8ca14ad28c4751afd14f3e3146a785dc", mergebase)
	})

	t.Run("No mergebase", func(t *testing.T) {
		mergebase, err := repo.GetMergeBaseSimple("main", "br3")
		require.ErrorContains(t, err, "exit status 1")
		assert.Empty(t, mergebase)
	})

	t.Run("Multiple mergebase", func(t *testing.T) {
		mergebase, err := repo.GetMergeBaseSimple("main", "br1")
		require.NoError(t, err)
		assert.Equal(t, "9d36f18c8ca14ad28c4751afd14f3e3146a785dc", mergebase)
	})
}
