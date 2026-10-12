// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git_test

import (
	"path/filepath"
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/git"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testReposDir = "tests/repos/"
)

// openRepositoryWithDefaultContext opens the repository at the given path with DefaultContext.
func openRepositoryWithDefaultContext(repoPath string) (*git.Repository, error) {
	return git.OpenRepository(git.DefaultContext, repoPath)
}

func TestGetNotes(t *testing.T) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	bareRepo1, err := openRepositoryWithDefaultContext(bareRepo1Path)
	require.NoError(t, err)
	defer bareRepo1.Close()

	note, err := git.GetNote(t.Context(), bareRepo1, "95bb4d39648ee7e325106df01a621c530863a653")
	require.NoError(t, err)
	assert.Equal(t, []byte("Note contents\n"), note.Message)
	assert.Equal(t, "Vladimir Panteleev", note.Commit.Author.Name)
	assert.Equal(t, "ca6b5ddf303169a72d2a2971acde4f6eea194e5c", note.Commit.ID.String())
}

func TestGetNestedNotes(t *testing.T) {
	repoPath := filepath.Join(testReposDir, "repo3_notes")
	repo, err := openRepositoryWithDefaultContext(repoPath)
	require.NoError(t, err)
	defer repo.Close()

	note, err := git.GetNote(t.Context(), repo, "3e668dbfac39cbc80a9ff9c61eb565d944453ba4")
	require.NoError(t, err)
	assert.Equal(t, []byte("Note 2"), note.Message)
	assert.Equal(t, "654c8b6b63c08bf37f638d3f521626b7fbbd4d37", note.Commit.ID.String())
	note, err = git.GetNote(t.Context(), repo, "ba0a96fa63532d6c5087ecef070b0250ed72fa47")
	require.NoError(t, err)
	assert.Equal(t, []byte("Note 1"), note.Message)
	assert.Equal(t, "654c8b6b63c08bf37f638d3f521626b7fbbd4d37", note.Commit.ID.String())
}

func TestGetNonExistentNotes(t *testing.T) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	bareRepo1, err := openRepositoryWithDefaultContext(bareRepo1Path)
	require.NoError(t, err)
	defer bareRepo1.Close()

	note, err := git.GetNote(t.Context(), bareRepo1, "non_existent_sha")
	require.Error(t, err)
	assert.True(t, git.IsErrNotExist(err))
	assert.Nil(t, note)
}

func TestSetNote(t *testing.T) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")

	tempDir := t.TempDir()
	require.NoError(t, unittest.CopyDir(bareRepo1Path, filepath.Join(tempDir, "repo1")))

	bareRepo1, err := openRepositoryWithDefaultContext(filepath.Join(tempDir, "repo1"))
	require.NoError(t, err)
	defer bareRepo1.Close()

	require.NoError(t, git.SetNote(t.Context(), bareRepo1, "95bb4d39648ee7e325106df01a621c530863a653", "This is a new note", "Test", "test@test.com"))

	note, err := git.GetNote(t.Context(), bareRepo1, "95bb4d39648ee7e325106df01a621c530863a653")
	require.NoError(t, err)
	assert.Equal(t, []byte("This is a new note\n"), note.Message)
	assert.Equal(t, "Test", note.Commit.Author.Name)
}

func TestRemoveNote(t *testing.T) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")

	tempDir := t.TempDir()

	require.NoError(t, unittest.CopyDir(bareRepo1Path, filepath.Join(tempDir, "repo1")))

	bareRepo1, err := openRepositoryWithDefaultContext(filepath.Join(tempDir, "repo1"))
	require.NoError(t, err)
	defer bareRepo1.Close()

	require.NoError(t, git.RemoveNote(t.Context(), bareRepo1, "95bb4d39648ee7e325106df01a621c530863a653"))

	note, err := git.GetNote(t.Context(), bareRepo1, "95bb4d39648ee7e325106df01a621c530863a653")
	require.Error(t, err)
	assert.True(t, git.IsErrNotExist(err))
	assert.Nil(t, note)
}
