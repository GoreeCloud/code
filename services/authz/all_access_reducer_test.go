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

func TestAllAccessAuthorizationReducer(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	reducer := &AllAccessAuthorizationReducer{}

	t.Run("ReduceRepoAccess no changes", func(t *testing.T) {
		repo1 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 1})
		for _, am := range []perm.AccessMode{perm.AccessModeOwner, perm.AccessModeAdmin, perm.AccessModeWrite, perm.AccessModeRead} {
			p1, err := reducer.ReduceRepoAccess(t.Context(), repo1, am)
			require.NoError(t, err)
			assert.Equal(t, am, p1)
		}
	})

	t.Run("RepoFilter no restrictions", func(t *testing.T) {
		numRepos, err := db.GetEngine(t.Context()).Table(&repo.Repository{}).Count()
		require.NoError(t, err)

		cond := reducer.RepoReadAccessFilter()

		var rows []*repo.Repository
		err = db.GetEngine(t.Context()).Table(&repo.Repository{}).Where(cond).OrderBy("id").Cols("id", "owner_id", "is_private").Find(&rows)
		require.NoError(t, err)
		assert.Len(t, rows, int(numRepos))
	})

	t.Run("AllowAdminOverride is true", func(t *testing.T) {
		assert.True(t, reducer.AllowAdminOverride())
	})
}
