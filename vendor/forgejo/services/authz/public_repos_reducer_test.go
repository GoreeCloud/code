// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/perm"
	"forgejo.org/models/repo"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicReposAuthorizationReducer(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	reducer := &PublicReposAuthorizationReducer{}

	t.Run("ReduceRepoAccess unrestricted on public repos", func(t *testing.T) {
		repo1 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 1})
		repo4 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 4})
		for _, am := range []perm.AccessMode{perm.AccessModeOwner, perm.AccessModeAdmin, perm.AccessModeWrite, perm.AccessModeRead} {
			p1, err := reducer.ReduceRepoAccess(t.Context(), repo1, am)
			require.NoError(t, err)
			assert.Equal(t, am, p1)
			p4, err := reducer.ReduceRepoAccess(t.Context(), repo4, am)
			require.NoError(t, err)
			assert.Equal(t, am, p4)
		}
	})

	t.Run("ReduceRepoAccess restricted to None on private repos", func(t *testing.T) {
		// private repo
		repo3 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 3})

		// public repo on a limited-visibility org
		repo38 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 38})

		for _, am := range []perm.AccessMode{perm.AccessModeOwner, perm.AccessModeAdmin, perm.AccessModeWrite, perm.AccessModeRead} {
			p3, err := reducer.ReduceRepoAccess(t.Context(), repo3, am)
			require.NoError(t, err)
			assert.Equal(t, perm.AccessModeNone, p3)

			p38, err := reducer.ReduceRepoAccess(t.Context(), repo38, am)
			require.NoError(t, err)
			assert.Equal(t, perm.AccessModeNone, p38)
		}
	})

	t.Run("RepoFilter unrestricted access only permitted to public repos", func(t *testing.T) {
		cond := reducer.RepoReadAccessFilter()

		var rows []*repo.Repository
		err := db.GetEngine(t.Context()).Table(&repo.Repository{}).Where(cond).OrderBy("id").Cols("id", "owner_id", "is_private").Find(&rows)
		require.NoError(t, err)
		assert.NotEmpty(t, rows)
		for _, repo := range rows {
			assert.False(t, repo.IsPrivate)
			require.NoError(t, repo.LoadOwner(t.Context()))
			assert.True(t, repo.Owner.Visibility.IsPublic())
		}
	})

	t.Run("AllowAdminOverride is false", func(t *testing.T) {
		assert.False(t, reducer.AllowAdminOverride())
	})
}
