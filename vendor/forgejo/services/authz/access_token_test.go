// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	"strings"
	"testing"

	"forgejo.org/models/auth"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthorizationReducerForAccessToken(t *testing.T) {
	defer unittest.OverrideFixtures("services/authz/TestGetAuthorizationReducerForAccessToken")()
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("all access", func(t *testing.T) {
		token := unittest.AssertExistsAndLoadBean(t, &auth.AccessToken{ID: 5})
		reducer, err := GetAuthorizationReducerForAccessToken(t.Context(), token)
		require.NoError(t, err)
		assert.IsType(t, &AllAccessAuthorizationReducer{}, reducer)
	})

	t.Run("public resources only", func(t *testing.T) {
		token := unittest.AssertExistsAndLoadBean(t, &auth.AccessToken{ID: 6})
		reducer, err := GetAuthorizationReducerForAccessToken(t.Context(), token)
		require.NoError(t, err)
		assert.IsType(t, &PublicReposAuthorizationReducer{}, reducer)
	})

	t.Run("specific repos only", func(t *testing.T) {
		token := unittest.AssertExistsAndLoadBean(t, &auth.AccessToken{ID: 7})
		reducer, err := GetAuthorizationReducerForAccessToken(t.Context(), token)
		require.NoError(t, err)

		specific, ok := reducer.(*SpecificReposAuthorizationReducer)
		require.True(t, ok)
		require.NotNil(t, specific)

		require.Len(t, specific.resourceRepos, 1)
		assert.EqualValues(t, 1, specific.resourceRepos[0].RepoID)
	})
}

func TestValidateAccessToken(t *testing.T) {
	t.Run("valid - all access", func(t *testing.T) {
		token := &auth.AccessToken{
			ResourceAllRepos: true,
			Scope:            auth.AccessTokenScopeReadRepository,
		}
		err := ValidateAccessToken(token, nil)
		require.NoError(t, err)
	})

	t.Run("valid - specified repos", func(t *testing.T) {
		token := &auth.AccessToken{
			ResourceAllRepos: false,
			Scope:            auth.AccessTokenScopeReadRepository,
		}
		resources := []*auth.AccessTokenResourceRepo{{RepoID: 12}}
		err := ValidateAccessToken(token, resources)
		require.NoError(t, err)
	})

	t.Run("invalid - no specified repos", func(t *testing.T) {
		token := &auth.AccessToken{
			ResourceAllRepos: false,
			Scope:            auth.AccessTokenScopeReadRepository,
		}
		resources := []*auth.AccessTokenResourceRepo{}
		err := ValidateAccessToken(token, resources)
		require.ErrorIs(t, err, ErrSpecifiedReposNone)
	})

	t.Run("invalid - specified repos & public-only", func(t *testing.T) {
		token := &auth.AccessToken{
			ResourceAllRepos: false,
			Scope:            auth.AccessTokenScope(strings.Join([]string{string(auth.AccessTokenScopePublicOnly), string(auth.AccessTokenScopeReadRepository)}, ",")),
		}
		resources := []*auth.AccessTokenResourceRepo{{RepoID: 12}}
		err := ValidateAccessToken(token, resources)
		require.ErrorIs(t, err, ErrSpecifiedReposNoPublicOnly)
	})

	t.Run("invalid - specified repos unsupported scopes", func(t *testing.T) {
		token := &auth.AccessToken{
			ResourceAllRepos: false,
			Scope:            auth.AccessTokenScopeReadAdmin,
		}
		resources := []*auth.AccessTokenResourceRepo{{RepoID: 12}}
		err := ValidateAccessToken(token, resources)
		require.ErrorIs(t, err, ErrSpecifiedReposInvalidScope)
		require.ErrorContains(t, err, string(auth.AccessTokenScopeReadAdmin))
	})
}
