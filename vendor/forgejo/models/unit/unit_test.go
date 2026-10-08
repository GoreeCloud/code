// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package unit_test

import (
	"testing"

	unit_model "forgejo.org/models/unit"
	"forgejo.org/models/unit/tests"
	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadUnitConfig(t *testing.T) {
	t.Run("regular", func(t *testing.T) {
		defer tests.SaveUnits()()

		setting.Repository.DisabledRepoUnits = []string{"repo.issues"}
		setting.Repository.DefaultRepoUnits = []string{"repo.code", "repo.releases", "repo.issues", "repo.pulls"}
		setting.Repository.DefaultForkRepoUnits = []string{"repo.releases"}
		require.NoError(t, unit_model.LoadUnitConfig())
		assert.Equal(t, []unit_model.Type{unit_model.TypeIssues}, unit_model.DisabledRepoUnitsGet())
		assert.Equal(t, []unit_model.Type{unit_model.TypeCode, unit_model.TypeReleases, unit_model.TypePullRequests}, unit_model.DefaultRepoUnits)
		assert.Equal(t, []unit_model.Type{unit_model.TypeReleases}, unit_model.DefaultForkRepoUnits)
	})
	t.Run("invalid", func(t *testing.T) {
		defer tests.SaveUnits()()

		setting.Repository.DisabledRepoUnits = []string{"repo.issues", "invalid.1"}
		setting.Repository.DefaultRepoUnits = []string{"repo.code", "invalid.2", "repo.releases", "repo.issues", "repo.pulls"}
		setting.Repository.DefaultForkRepoUnits = []string{"invalid.3", "repo.releases"}
		require.NoError(t, unit_model.LoadUnitConfig())
		assert.Equal(t, []unit_model.Type{unit_model.TypeIssues}, unit_model.DisabledRepoUnitsGet())
		assert.Equal(t, []unit_model.Type{unit_model.TypeCode, unit_model.TypeReleases, unit_model.TypePullRequests}, unit_model.DefaultRepoUnits)
		assert.Equal(t, []unit_model.Type{unit_model.TypeReleases}, unit_model.DefaultForkRepoUnits)
	})
	t.Run("duplicate", func(t *testing.T) {
		defer tests.SaveUnits()()

		setting.Repository.DisabledRepoUnits = []string{"repo.issues", "repo.issues"}
		setting.Repository.DefaultRepoUnits = []string{"repo.code", "repo.releases", "repo.issues", "repo.pulls", "repo.code"}
		setting.Repository.DefaultForkRepoUnits = []string{"repo.releases", "repo.releases"}
		require.NoError(t, unit_model.LoadUnitConfig())
		assert.Equal(t, []unit_model.Type{unit_model.TypeIssues}, unit_model.DisabledRepoUnitsGet())
		assert.Equal(t, []unit_model.Type{unit_model.TypeCode, unit_model.TypeReleases, unit_model.TypePullRequests}, unit_model.DefaultRepoUnits)
		assert.Equal(t, []unit_model.Type{unit_model.TypeReleases}, unit_model.DefaultForkRepoUnits)
	})
	t.Run("empty_default", func(t *testing.T) {
		defer tests.SaveUnits()()

		setting.Repository.DisabledRepoUnits = []string{"repo.issues", "repo.issues"}
		setting.Repository.DefaultRepoUnits = []string{}
		setting.Repository.DefaultForkRepoUnits = []string{"repo.releases", "repo.releases"}
		require.NoError(t, unit_model.LoadUnitConfig())
		assert.Equal(t, []unit_model.Type{unit_model.TypeIssues}, unit_model.DisabledRepoUnitsGet())
		assert.ElementsMatch(t, []unit_model.Type{unit_model.TypeCode, unit_model.TypePullRequests, unit_model.TypeReleases, unit_model.TypeWiki, unit_model.TypePackages, unit_model.TypeProjects, unit_model.TypeActions}, unit_model.DefaultRepoUnits)
		assert.Equal(t, []unit_model.Type{unit_model.TypeReleases}, unit_model.DefaultForkRepoUnits)
	})
}
