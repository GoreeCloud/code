// Copyright 2020 The Gitea Authors. All rights reserved.
// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package project

import (
	"testing"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsProjectTypeValid(t *testing.T) {
	const UnknownType Type = 15

	cases := []struct {
		typ   Type
		valid bool
	}{
		{TypeIndividual, true},
		{TypeRepository, true},
		{TypeOrganization, true},
		{UnknownType, false},
	}

	for _, v := range cases {
		assert.Equal(t, v.valid, IsTypeValid(v.typ))
	}
}

func TestGetProjects(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	projects, err := db.Find[Project](db.DefaultContext, SearchOptions{RepoID: 1})
	require.NoError(t, err)

	// 1 value for this repo exists in the fixtures
	assert.Len(t, projects, 1)

	projects, err = db.Find[Project](db.DefaultContext, SearchOptions{RepoID: 3})
	require.NoError(t, err)

	// 1 value for this repo exists in the fixtures
	assert.Len(t, projects, 1)
}

func TestProjectsSort(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	tests := []struct {
		sortType string
		wants    []int64
	}{
		{
			sortType: "default",
			wants:    []int64{1, 3, 2, 7, 6, 5, 4},
		},
		{
			sortType: "oldest",
			wants:    []int64{4, 5, 6, 7, 2, 3, 1},
		},
		{
			sortType: "recentupdate",
			wants:    []int64{1, 3, 2, 7, 6, 5, 4},
		},
		{
			sortType: "leastupdate",
			wants:    []int64{4, 5, 6, 7, 2, 3, 1},
		},
	}

	for _, tt := range tests {
		projects, count, err := db.FindAndCount[Project](db.DefaultContext, SearchOptions{
			OrderBy: GetSearchOrderByBySortType(tt.sortType),
		})
		require.NoError(t, err)
		assert.Equal(t, int64(7), count)
		if assert.Len(t, projects, 7) {
			for i := range projects {
				assert.Equal(t, tt.wants[i], projects[i].ID)
			}
		}
	}
}

func TestGetProjectForUserByID(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	found := func(t *testing.T, uid, id int64) {
		t.Helper()

		p, err := GetProjectForUserByID(t.Context(), uid, id)
		require.NoError(t, err)
		if assert.NotNil(t, p) {
			assert.Equal(t, id, p.ID)
		}
	}

	notFound := func(t *testing.T, uid, id int64) {
		t.Helper()

		p, err := GetProjectForUserByID(t.Context(), uid, id)
		require.ErrorIs(t, err, ErrProjectNotExist{ID: id})
		assert.Nil(t, p)
	}

	found(t, 2, 4)
	found(t, 2, 5)
	found(t, 2, 6)
	found(t, 3, 7)
	notFound(t, 1, 4)
	notFound(t, 1, 5)
	notFound(t, 1, 6)
	notFound(t, 1, 7)
}

func TestChangeProjectStatus(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("Unchanged", func(t *testing.T) {
		project := unittest.AssertExistsAndLoadBean(t, &Project{ID: 1})

		require.NoError(t, ChangeProjectStatus(t.Context(), project, project.IsClosed))

		projectAfter := unittest.AssertExistsAndLoadBean(t, &Project{ID: 1})
		assert.Equal(t, project.IsClosed, projectAfter.IsClosed)
	})

	t.Run("Normal", func(t *testing.T) {
		project := unittest.AssertExistsAndLoadBean(t, &Project{ID: 1})
		isClosed := !project.IsClosed
		repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})

		require.NoError(t, ChangeProjectStatus(t.Context(), project, isClosed))

		projectAfter := unittest.AssertExistsAndLoadBean(t, &Project{ID: 1})
		repoAfter := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})
		assert.Equal(t, isClosed, projectAfter.IsClosed)
		assert.Equal(t, repo.NumProjects, repoAfter.NumProjects)
		assert.Equal(t, repo.NumOpenProjects-1, repoAfter.NumOpenProjects)
		assert.Equal(t, repo.NumClosedProjects+1, repoAfter.NumClosedProjects)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		project := &Project{ID: 1001, RepoID: 1}
		repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})

		require.NoError(t, ChangeProjectStatus(t.Context(), project, true))

		repoAfter := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})
		assert.Equal(t, repo.NumProjects, repoAfter.NumProjects)
		assert.Equal(t, repo.NumOpenProjects, repoAfter.NumOpenProjects)
		assert.Equal(t, repo.NumClosedProjects, repoAfter.NumClosedProjects)
	})
}
