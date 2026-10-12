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

func TestGetRepositoriesAccessibleWithIntegration(t *testing.T) {
	defer unittest.OverrideFixtures("models/auth/TestGetRepositoriesAccessibleWithIntegration")()
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("No Resources", func(t *testing.T) {
		resources, err := auth_model.GetRepositoriesAccessibleWithIntegration(t.Context(), 999)
		require.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("Has Resources", func(t *testing.T) {
		resources, err := auth_model.GetRepositoriesAccessibleWithIntegration(t.Context(), 1)
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

func TestInsertAuthorizedIntegration(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	ai1 := makeAuthorizedIntegration(t)
	ai2 := makeAuthorizedIntegration(t)
	ai3 := makeAuthorizedIntegration(t)

	t.Run("blank insert", func(t *testing.T) {
		err := auth_model.InsertAuthorizedIntegrationResourceRepos(t.Context(), ai1.ID, nil)
		require.NoError(t, err)
	})

	t.Run("multiple insert", func(t *testing.T) {
		resRepo1 := &auth_model.AuthorizedIntegResourceRepo{
			IntegID: ai2.ID,
			RepoID:  1,
		}
		resRepo3 := &auth_model.AuthorizedIntegResourceRepo{
			IntegID: ai2.ID,
			RepoID:  3,
		}
		err := auth_model.InsertAuthorizedIntegrationResourceRepos(t.Context(), ai2.ID,
			[]*auth_model.AuthorizedIntegResourceRepo{resRepo1, resRepo3})
		require.NoError(t, err)

		unittest.AssertCount(t, &auth_model.AuthorizedIntegResourceRepo{IntegID: ai2.ID}, 2)
	})

	t.Run("in tx", func(t *testing.T) {
		// Pre-condition: count is 0.
		unittest.AssertCount(t, &auth_model.AuthorizedIntegResourceRepo{IntegID: ai3.ID}, 0)

		// Verify that InsertAuthorizedIntegrationResourceRepos performs inserts in a TX by having a second one with an invalid
		// RepoID, causing a foreign key violation
		resRepo1 := &auth_model.AuthorizedIntegResourceRepo{
			IntegID: ai3.ID,
			RepoID:  1,
		}
		resRepo3 := &auth_model.AuthorizedIntegResourceRepo{
			IntegID: ai3.ID,
			RepoID:  30000, // invalid
		}
		err := auth_model.InsertAuthorizedIntegrationResourceRepos(t.Context(), ai3.ID,
			[]*auth_model.AuthorizedIntegResourceRepo{resRepo1, resRepo3})
		require.ErrorContains(t, err, "foreign key")

		// Count remains 0; the first record was not inserted.
		unittest.AssertCount(t, &auth_model.AuthorizedIntegResourceRepo{IntegID: ai3.ID}, 0)
	})
}
