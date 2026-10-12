// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"
	"time"

	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionUser_CreateDelete(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 2})
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	require.ErrorContains(t, InsertActionUser(t.Context(), &ActionUser{
		UserID: user.ID,
	}), "FOREIGN KEY")

	require.ErrorContains(t, InsertActionUser(t.Context(), &ActionUser{
		RepoID: repo.ID,
	}), "FOREIGN KEY")

	actionUser := &ActionUser{
		UserID: user.ID,
		RepoID: repo.ID,
	}
	require.NoError(t, InsertActionUser(t.Context(), actionUser))
	assert.NotZero(t, actionUser.ID)
	assert.NotZero(t, actionUser.LastAccess)

	otherUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 3})
	actionUserNotSameUser := &ActionUser{
		UserID: otherUser.ID,
		RepoID: repo.ID,
	}
	require.NoError(t, InsertActionUser(t.Context(), actionUserNotSameUser))

	otherRepo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 3})
	actionUserNotSameRepo := &ActionUser{
		UserID: user.ID,
		RepoID: otherRepo.ID,
	}
	require.NoError(t, InsertActionUser(t.Context(), actionUserNotSameRepo))

	unittest.AssertExistsAndLoadBean(t, &ActionUser{ID: actionUser.ID})
	require.NoError(t, DeleteActionUserByUserIDAndRepoID(t.Context(), user.ID, repo.ID))
	unittest.AssertNotExistsBean(t, &ActionUser{ID: actionUser.ID})
}

func TestActionUser_RevokeInactiveActionUser(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 2})
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	actionUser := &ActionUser{
		UserID: user.ID,
		RepoID: repo.ID,
	}
	require.NoError(t, InsertActionUser(t.Context(), actionUser))

	t.Run("not revoked because it was just created", func(t *testing.T) {
		unittest.AssertExistsAndLoadBean(t, &ActionUser{ID: actionUser.ID})
		require.NoError(t, RevokeInactiveActionUser(t.Context()))
		unittest.AssertExistsAndLoadBean(t, &ActionUser{ID: actionUser.ID})
	})

	// needs to be at least 1 second because unix timestamp resolution is 1 second
	defer test.MockVariableValue(&expire, 1*time.Second)()

	t.Run("used not updated too frequently", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		usedActionUser, err := GetActionUserByUserIDAndRepoIDAndUpdateAccess(t.Context(), user.ID, repo.ID)
		require.NoError(t, err)
		require.Equal(t, actionUser.ID, usedActionUser.ID)
		assert.Equal(t, usedActionUser.LastAccess, actionUser.LastAccess)
	})

	defer test.MockVariableValue(&updateFrequency, 0)()

	t.Run("not revoked because it was recently used", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		usedActionUser, err := GetActionUserByUserIDAndRepoIDAndUpdateAccess(t.Context(), user.ID, repo.ID)
		require.NoError(t, err)
		require.Equal(t, actionUser.ID, usedActionUser.ID)
		assert.Greater(t, usedActionUser.LastAccess, actionUser.LastAccess)
		require.NoError(t, RevokeInactiveActionUser(t.Context()))
		unittest.AssertExistsAndLoadBean(t, &ActionUser{ID: actionUser.ID})
	})

	t.Run("revoked because it was not recently used", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		unittest.AssertExistsAndLoadBean(t, &ActionUser{ID: actionUser.ID})
		require.NoError(t, RevokeInactiveActionUser(t.Context()))
		unittest.AssertNotExistsBean(t, &ActionUser{ID: actionUser.ID})
	})
}
