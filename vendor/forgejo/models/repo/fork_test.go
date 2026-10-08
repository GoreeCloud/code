// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo_test

import (
	"testing"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserFork(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// User13 has repo 11 forked from repo10
	repo, err := repo_model.GetRepositoryByID(db.DefaultContext, 10)
	require.NoError(t, err)
	assert.NotNil(t, repo)
	repo, err = repo_model.GetUserFork(db.DefaultContext, repo.ID, 13)
	require.NoError(t, err)
	assert.NotNil(t, repo)

	repo, err = repo_model.GetRepositoryByID(db.DefaultContext, 9)
	require.NoError(t, err)
	assert.NotNil(t, repo)
	repo, err = repo_model.GetUserFork(db.DefaultContext, repo.ID, 13)
	require.NoError(t, err)
	assert.Nil(t, repo)
}

func TestGetUserForkLax(t *testing.T) {
	defer unittest.OverrideFixtures("models/repo/TestGetUserForkLax")()
	require.NoError(t, unittest.PrepareTestDatabase())

	// User13 has repo 11 forked from repo10
	repo10, err := repo_model.GetRepositoryByID(db.DefaultContext, 10)
	require.NoError(t, err)
	assert.NotNil(t, repo10)
	require.True(t, repo_model.HasForkedRepoLax(db.DefaultContext, 13, repo10))
	repo11, err := repo_model.GetUserForkLax(db.DefaultContext, repo10, 13)
	require.NoError(t, err)
	assert.NotNil(t, repo11)
	assert.Equal(t, int64(11), repo11.ID)
	assert.Equal(t, int64(10), repo11.ForkID)

	// user13 does not have a fork of repo9
	repo9, err := repo_model.GetRepositoryByID(db.DefaultContext, 9)
	require.NoError(t, err)
	assert.NotNil(t, repo9)
	require.False(t, repo_model.HasForkedRepoLax(db.DefaultContext, 13, repo9))
	fork, err := repo_model.GetUserForkLax(db.DefaultContext, repo9, 13)
	require.NoError(t, err)
	assert.Nil(t, fork)

	// User15 has repo id 64 forked from repo10, which counts as a fork of repo11 since they have a common base
	require.False(t, repo_model.HasForkedRepo(db.DefaultContext, 15, repo11.ID))
	require.True(t, repo_model.HasForkedRepoLax(db.DefaultContext, 15, repo11))
	fork, err = repo_model.GetUserForkLax(db.DefaultContext, repo11, 15)
	require.NoError(t, err)
	assert.NotNil(t, fork)
	assert.Equal(t, int64(64), fork.ID)
	assert.Equal(t, int64(10), fork.ForkID)
}

func TestGetUserForkLaxWithTwoChoices(t *testing.T) {
	defer unittest.OverrideFixtures("models/repo/TestGetUserForkLaxWithTwoChoices")()
	require.NoError(t, unittest.PrepareTestDatabase())

	// Test scenario:
	//
	// - repo10
	//     - forked by user15 as repo64
	//     - forked by user13 as repo11
	//         - forked by user15 as repo65
	//
	// In this scenario, both repo64 and repo65 can be used as forks of repo11 for user15,
	// but we prefer to use repo65 because it's specifically marked as a fork of repo11

	repo10, err := repo_model.GetRepositoryByID(db.DefaultContext, 10)
	require.NoError(t, err)
	assert.NotNil(t, repo10)

	// User15 has repo64 forked from repo10
	require.True(t, repo_model.HasForkedRepoLax(db.DefaultContext, 15, repo10))
	repo64, err := repo_model.GetUserForkLax(db.DefaultContext, repo10, 15)
	require.NoError(t, err)
	assert.NotNil(t, repo64)
	assert.Equal(t, int64(64), repo64.ID)
	assert.Equal(t, int64(10), repo64.ForkID)

	// User13 has repo11 forked from repo10
	require.True(t, repo_model.HasForkedRepoLax(db.DefaultContext, 13, repo10))
	repo11, err := repo_model.GetUserForkLax(db.DefaultContext, repo10, 13)
	require.NoError(t, err)
	assert.NotNil(t, repo11)
	assert.Equal(t, int64(11), repo11.ID)
	assert.Equal(t, int64(10), repo11.ForkID)

	// User15 has repo65 forked from repo11
	require.True(t, repo_model.HasForkedRepoLax(db.DefaultContext, 15, repo11))
	repo65, err := repo_model.GetUserForkLax(db.DefaultContext, repo11, 15)
	require.NoError(t, err)
	assert.NotNil(t, repo65)
	assert.Equal(t, int64(65), repo65.ID)
	assert.Equal(t, int64(11), repo65.ForkID)
}
