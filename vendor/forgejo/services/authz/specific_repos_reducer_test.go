// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	"testing"

	"forgejo.org/models/auth"
	"forgejo.org/models/db"
	"forgejo.org/models/perm"
	"forgejo.org/models/repo"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpecificReposAuthorizationReducer(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	reducer := &SpecificReposAuthorizationReducer{
		resourceRepos: []*auth.AccessTokenResourceRepo{
			{
				RepoID: 1,
			},
			{
				RepoID: 2,
			},
		},
	}

	t.Run("ReduceRepoAccess unrestricted on targeted repos", func(t *testing.T) {
		repo1 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 1})
		repo2 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 2})
		for _, am := range []perm.AccessMode{perm.AccessModeOwner, perm.AccessModeAdmin, perm.AccessModeWrite, perm.AccessModeRead} {
			p1, err := reducer.ReduceRepoAccess(t.Context(), repo1, am)
			require.NoError(t, err)
			assert.Equal(t, am, p1)
			p2, err := reducer.ReduceRepoAccess(t.Context(), repo2, am)
			require.NoError(t, err)
			assert.Equal(t, am, p2)
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

	t.Run("ReduceRepoAccess restricted to Read on public repos", func(t *testing.T) {
		// public repo
		repo4 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 4})

		for _, am := range []perm.AccessMode{perm.AccessModeOwner, perm.AccessModeAdmin, perm.AccessModeWrite, perm.AccessModeRead} {
			p3, err := reducer.ReduceRepoAccess(t.Context(), repo4, am)
			require.NoError(t, err)
			assert.Equal(t, perm.AccessModeRead, p3)
		}

		// don't elevate AccessModeNone to AccessModeRead:
		p3, err := reducer.ReduceRepoAccess(t.Context(), repo4, perm.AccessModeNone)
		require.NoError(t, err)
		assert.Equal(t, perm.AccessModeNone, p3)
	})

	t.Run("RepoFilter read access only permitted to target repos & public repos", func(t *testing.T) {
		cond := reducer.RepoReadAccessFilter()

		var rows []*repo.Repository
		err := db.GetEngine(t.Context()).Table(&repo.Repository{}).Where(cond).OrderBy("id").Cols("id", "owner_id", "is_private").Find(&rows)
		require.NoError(t, err)

		// Both target repos should be returned:
		assert.EqualValues(t, 1, rows[0].ID)
		assert.EqualValues(t, 2, rows[1].ID)

		// And there should be more return values, all of which appear as public repos:
		assert.Greater(t, len(rows), 2)
		for _, repo := range rows[2:] {
			assert.False(t, repo.IsPrivate)
			require.NoError(t, repo.LoadOwner(t.Context()))
			assert.True(t, repo.Owner.Visibility.IsPublic())
		}
	})

	t.Run("AllowAdminOverride is false", func(t *testing.T) {
		assert.False(t, reducer.AllowAdminOverride())
	})
}
