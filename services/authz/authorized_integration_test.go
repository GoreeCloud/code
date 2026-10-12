// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	"testing"

	"forgejo.org/models/auth"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthorizationReducerForAuthorizedIntegration(t *testing.T) {
	defer unittest.OverrideFixtures("services/authz/TestGetAuthorizationReducerForAuthorizedIntegration")()
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("all access", func(t *testing.T) {
		token := unittest.AssertExistsAndLoadBean(t, &auth.AuthorizedIntegration{ID: 5})
		reducer, err := GetAuthorizationReducerForAuthorizedIntegration(t.Context(), token)
		require.NoError(t, err)
		assert.IsType(t, &AllAccessAuthorizationReducer{}, reducer)
	})

	t.Run("public resources only", func(t *testing.T) {
		token := unittest.AssertExistsAndLoadBean(t, &auth.AuthorizedIntegration{ID: 6})
		reducer, err := GetAuthorizationReducerForAuthorizedIntegration(t.Context(), token)
		require.NoError(t, err)
		assert.IsType(t, &PublicReposAuthorizationReducer{}, reducer)
	})

	t.Run("specific repos only", func(t *testing.T) {
		token := unittest.AssertExistsAndLoadBean(t, &auth.AuthorizedIntegration{ID: 7})
		reducer, err := GetAuthorizationReducerForAuthorizedIntegration(t.Context(), token)
		require.NoError(t, err)

		specific, ok := reducer.(*SpecificReposAuthorizationReducer)
		require.True(t, ok)
		require.NotNil(t, specific)

		require.Len(t, specific.ResourceRepos, 1)
		assert.EqualValues(t, 1, specific.ResourceRepos[0].GetTargetRepoID())
	})
}
