// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth_test

import (
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRepositoriesAccessibleWithToken(t *testing.T) {
	defer unittest.OverrideFixtures("models/auth/TestGetRepositoriesAccessibleWithToken")()
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("No Resources", func(t *testing.T) {
		resources, err := auth_model.GetRepositoriesAccessibleWithToken(t.Context(), 999)
		require.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("Has Resources", func(t *testing.T) {
		resources, err := auth_model.GetRepositoriesAccessibleWithToken(t.Context(), 3)
		require.NoError(t, err)
		require.Len(t, resources, 3)

		// Verify all expected repo IDs are present
		repoIDs := make([]int64, len(resources))
		for i, res := range resources {
			repoIDs[i] = res.RepoID
		}
		assert.Contains(t, repoIDs, int64(1))
		assert.Contains(t, repoIDs, int64(2))
		assert.Contains(t, repoIDs, int64(3))
	})
}

func TestGetRepositoriesAccessibleWithTokens(t *testing.T) {
	defer unittest.OverrideFixtures("models/auth/TestGetRepositoriesAccessibleWithTokens")()
	require.NoError(t, unittest.PrepareTestDatabase())

	token1 := unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{ID: 1})
	token2 := unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{ID: 2})
	token3 := unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{ID: 3})

	t.Run("No Tokens", func(t *testing.T) {
		resources, err := auth_model.GetRepositoriesAccessibleWithTokens(t.Context(), []*auth_model.AccessToken{})
		require.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("Multiple Access Tokens", func(t *testing.T) {
		resources, err := auth_model.GetRepositoriesAccessibleWithTokens(t.Context(), []*auth_model.AccessToken{token1, token2, token3})
		require.NoError(t, err)

		repos1, ok := resources[token1.ID]
		assert.False(t, ok)
		assert.Empty(t, repos1)

		repos2, ok := resources[token2.ID]
		assert.True(t, ok)
		assert.Len(t, repos2, 2)

		repos3, ok := resources[token3.ID]
		assert.True(t, ok)
		assert.Len(t, repos3, 3)
	})
}

func TestInsertAccessTokenResourceRepos(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	token1 := unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{ID: 1})
	token2 := unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{ID: 2})
	token3 := unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{ID: 3})

	t.Run("blank insert", func(t *testing.T) {
		err := auth_model.InsertAccessTokenResourceRepos(t.Context(), token1.ID, nil)
		require.NoError(t, err)
	})

	t.Run("multiple insert", func(t *testing.T) {
		resRepo1 := &auth_model.AccessTokenResourceRepo{
			TokenID: token2.ID,
			RepoID:  1,
		}
		resRepo3 := &auth_model.AccessTokenResourceRepo{
			TokenID: token2.ID,
			RepoID:  3,
		}
		err := auth_model.InsertAccessTokenResourceRepos(t.Context(), token2.ID,
			[]*auth_model.AccessTokenResourceRepo{resRepo1, resRepo3})
		require.NoError(t, err)

		unittest.AssertCount(t, &auth_model.AccessTokenResourceRepo{TokenID: token2.ID}, 2)
	})

	t.Run("in tx", func(t *testing.T) {
		// Pre-condition: count is 0.
		unittest.AssertCount(t, &auth_model.AccessTokenResourceRepo{TokenID: token3.ID}, 0)

		// Verify that InsertAccessTokenResourceRepos performs inserts in a TX by having a second one with an invalid
		// RepoID, causing a foreign key violation
		resRepo1 := &auth_model.AccessTokenResourceRepo{
			TokenID: token3.ID,
			RepoID:  1,
		}
		resRepo3 := &auth_model.AccessTokenResourceRepo{
			TokenID: token3.ID,
			RepoID:  30000, // invalid
		}
		err := auth_model.InsertAccessTokenResourceRepos(t.Context(), token3.ID,
			[]*auth_model.AccessTokenResourceRepo{resRepo1, resRepo3})
		require.ErrorContains(t, err, "foreign key")

		// Count remains 0; the first record was not inserted.
		unittest.AssertCount(t, &auth_model.AccessTokenResourceRepo{TokenID: token3.ID}, 0)
	})
}
