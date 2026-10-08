// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unit"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkedRepository(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	testCases := []struct {
		name             string
		attachID         int64
		expectedRepo     *repo_model.Repository
		expectedUnitType unit.Type
	}{
		{"LinkedIssue", 1, &repo_model.Repository{ID: 1}, unit.TypeIssues},
		{"LinkedComment", 3, &repo_model.Repository{ID: 1}, unit.TypePullRequests},
		{"LinkedRelease", 9, &repo_model.Repository{ID: 1}, unit.TypeReleases},
		{"Notlinked", 10, nil, -1},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attach, err := repo_model.GetAttachmentByID(db.DefaultContext, tc.attachID)
			require.NoError(t, err)
			repo, unitType, err := LinkedRepository(db.DefaultContext, attach)
			require.NoError(t, err)
			if tc.expectedRepo != nil {
				assert.Equal(t, tc.expectedRepo.ID, repo.ID)
			}
			assert.Equal(t, tc.expectedUnitType, unitType)
		})
	}
}

func TestConvertMirrorToNormalRepo(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	repo.IsMirror = true
	err := repo_model.UpdateRepositoryCols(db.DefaultContext, repo, "is_mirror")

	require.NoError(t, err)

	err = ConvertMirrorToNormalRepo(db.DefaultContext, repo)
	require.NoError(t, err)
	assert.False(t, repo.IsMirror)
}

func TestDeleteRepository(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	doer := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
	require.NoError(t, DeleteRepository(t.Context(), doer, repo, false))
}

func TestDeleteRepositoryWithReferences(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})

	token1 := unittest.AssertExistsAndLoadBean(t, &auth_model.AccessToken{ID: 1})
	err := db.Insert(t.Context(), &auth_model.AccessTokenResourceRepo{
		TokenID: token1.ID,
		RepoID:  repo.ID,
	})
	require.NoError(t, err)

	doer := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
	require.NoError(t, DeleteRepository(t.Context(), doer, repo, false))
}
