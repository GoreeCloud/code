// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth_test

import (
	"testing"
	"time"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeAuthorizedIntegration(t *testing.T) *auth_model.AuthorizedIntegration {
	t.Helper()
	ai := &auth_model.AuthorizedIntegration{
		UserID:           2,
		Scope:            auth_model.AccessTokenScopeAll,
		ResourceAllRepos: true,
		Issuer:           "https://example.org/",
		ClaimRules:       &auth_model.ClaimRules{},
	}
	require.NoError(t, auth_model.InsertAuthorizedIntegration(t.Context(), ai))
	return ai
}

func TestGetAuthorizedIntegration(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	ai := makeAuthorizedIntegration(t)

	get, err := auth_model.GetAuthorizedIntegration(t.Context(), "abc", "123")
	require.ErrorIs(t, err, util.ErrNotExist)
	assert.Nil(t, get)

	get, err = auth_model.GetAuthorizedIntegration(t.Context(), ai.Issuer, ai.Audience)
	require.NoError(t, err)
	require.NotNil(t, get)
	assert.Equal(t, ai.ID, get.ID)
}

func TestGetAuthorizedIntegrationByUI(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	ai := makeAuthorizedIntegration(t)

	get, err := auth_model.GetAuthorizedIntegrationByUI(t.Context(), 3, ai.UI, ai.ID)
	require.ErrorIs(t, err, util.ErrNotExist)
	assert.Nil(t, get)

	get, err = auth_model.GetAuthorizedIntegrationByUI(t.Context(), ai.UserID, auth_model.AuthorizedIntegrationUI("incorrect"), ai.ID)
	require.ErrorIs(t, err, util.ErrNotExist)
	assert.Nil(t, get)

	get, err = auth_model.GetAuthorizedIntegrationByUI(t.Context(), ai.UserID, ai.UI, -1)
	require.ErrorIs(t, err, util.ErrNotExist)
	assert.Nil(t, get)

	get, err = auth_model.GetAuthorizedIntegrationByUI(t.Context(), ai.UserID, ai.UI, ai.ID)
	require.NoError(t, err)
	require.NotNil(t, get)
	assert.Equal(t, ai.ID, get.ID)
}

func TestAuthorizedIntegrationUpdateLastUsed(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	ai := makeAuthorizedIntegration(t)
	ai.UpdatedUnix = 0
	cnt, err := db.GetEngine(t.Context()).ID(ai.ID).Cols("updated_unix").NoAutoTime().Update(ai)
	require.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	timeutil.MockSet(time.Unix(1777130023, 0))
	defer timeutil.MockUnset()

	assert.EqualValues(t, 0, ai.UpdatedUnix)
	require.NoError(t, ai.UpdateLastUsed(t.Context()))
	assert.EqualValues(t, 1777130023, ai.UpdatedUnix) // object field updated
	assert.EqualValues(t, 1777130023, unittest.AssertExistsAndLoadBean(t, &auth_model.AuthorizedIntegration{ID: ai.ID}).UpdatedUnix)

	// nearly immediate redo should have same timestamp due to the 1 minute deduplication:
	timeutil.MockSet(time.Unix(1777130025, 0))
	require.NoError(t, ai.UpdateLastUsed(t.Context()))
	assert.EqualValues(t, 1777130023, ai.UpdatedUnix)                                                                                // object field not updated
	assert.EqualValues(t, 1777130023, unittest.AssertExistsAndLoadBean(t, &auth_model.AuthorizedIntegration{ID: ai.ID}).UpdatedUnix) // database field not updated

	// but if it's a little while later..
	timeutil.MockSet(time.Unix(1777131139, 0))
	require.NoError(t, ai.UpdateLastUsed(t.Context()))
	assert.EqualValues(t, 1777131139, ai.UpdatedUnix)                                                                                // object field updated
	assert.EqualValues(t, 1777131139, unittest.AssertExistsAndLoadBean(t, &auth_model.AuthorizedIntegration{ID: ai.ID}).UpdatedUnix) // database field updated
}

func TestNewAuthorizedIntegration(t *testing.T) {
	ai := &auth_model.AuthorizedIntegration{
		UserID:           2,
		Scope:            auth_model.AccessTokenScopeAll,
		ResourceAllRepos: true,
		Issuer:           "https://example.org/",
		ClaimRules:       &auth_model.ClaimRules{},
	}
	require.NoError(t, auth_model.InsertAuthorizedIntegration(t.Context(), ai))
	assert.Contains(t, ai.Audience, "u:2:")

	ai = &auth_model.AuthorizedIntegration{
		UserID:           2,
		Scope:            auth_model.AccessTokenScopeAll,
		ResourceAllRepos: true,
		Issuer:           "https://example.org/",
		Audience:         "I made my own audience",
		ClaimRules:       &auth_model.ClaimRules{},
	}
	require.ErrorContains(t, auth_model.InsertAuthorizedIntegration(t.Context(), ai), "audience cannot be provided")

	ai = &auth_model.AuthorizedIntegration{
		// Forgot to set UserID
		Scope:            auth_model.AccessTokenScopeAll,
		ResourceAllRepos: true,
		Issuer:           "https://example.org/",
		ClaimRules:       &auth_model.ClaimRules{},
	}
	require.ErrorContains(t, auth_model.InsertAuthorizedIntegration(t.Context(), ai), "UserID must be initialized")
}

func TestAuthorizedIntegrationCalculatedValues(t *testing.T) {
	t.Run("HasRecentActivity", func(t *testing.T) {
		timeutil.MockSet(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))
		ai := &auth_model.AuthorizedIntegration{
			UpdatedUnix: 1609459200,
			CreatedUnix: 1609459200,
		}
		assert.False(t, ai.HasRecentActivity())
		ai.UpdatedUnix = 1609459201
		assert.True(t, ai.HasRecentActivity())
		ai.CreatedUnix = 1577836800
		ai.UpdatedUnix = 1577836801
		assert.False(t, ai.HasRecentActivity())
	})

	t.Run("HasBeenUsed", func(t *testing.T) {
		ai := &auth_model.AuthorizedIntegration{
			UpdatedUnix: 1,
			CreatedUnix: 1,
		}
		assert.False(t, ai.HasBeenUsed())
		ai.UpdatedUnix = 2
		assert.True(t, ai.HasBeenUsed())
	})
}

func TestListAuthorizedIntegrationOptions(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	makeAuthorizedIntegration(t)
	makeAuthorizedIntegration(t)

	ais, err := db.Find[auth_model.AuthorizedIntegration](t.Context(),
		auth_model.ListAuthorizedIntegrationOptions{UserID: optional.None[int64]()})
	require.NoError(t, err)
	assert.Len(t, ais, 2)

	ais, err = db.Find[auth_model.AuthorizedIntegration](t.Context(),
		auth_model.ListAuthorizedIntegrationOptions{UserID: optional.Some(int64(2))})
	require.NoError(t, err)
	assert.Len(t, ais, 2)

	ais, err = db.Find[auth_model.AuthorizedIntegration](t.Context(),
		auth_model.ListAuthorizedIntegrationOptions{UserID: optional.Some(int64(22))})
	require.NoError(t, err)
	assert.Empty(t, ais)
}

func TestUpdateAuthorizedIntegration(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	ai := makeAuthorizedIntegration(t)
	ai.Name = "Hello, world!"
	ai.ResourceAllRepos = false
	createdUnix := ai.CreatedUnix
	updatedUnix := ai.UpdatedUnix

	require.NoError(t, auth_model.UpdateAuthorizedIntegration(t.Context(), ai))

	fromDB := unittest.AssertExistsAndLoadBean(t, &auth_model.AuthorizedIntegration{ID: ai.ID})
	assert.Equal(t, "Hello, world!", fromDB.Name)
	assert.False(t, fromDB.ResourceAllRepos)
	assert.Equal(t, createdUnix, fromDB.CreatedUnix)
	assert.Equal(t, updatedUnix, fromDB.UpdatedUnix) // not changed -- used to track usage for authentication
}

func TestDeleteAuthorizedIntegrationByID(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	t.Run("simple delete", func(t *testing.T) {
		ai := makeAuthorizedIntegration(t)
		err := auth_model.DeleteAuthorizedIntegrationByID(t.Context(), ai.ID, ai.UserID)
		require.NoError(t, err)
		unittest.AssertNotExistsBean(t, &auth_model.AuthorizedIntegration{ID: ai.ID})
	})

	t.Run("delete repo-specific", func(t *testing.T) {
		ai := makeAuthorizedIntegration(t)
		resRepo1 := &auth_model.AuthorizedIntegResourceRepo{
			IntegID: ai.ID,
			RepoID:  1,
		}
		err := auth_model.InsertAuthorizedIntegrationResourceRepos(t.Context(), ai.ID,
			[]*auth_model.AuthorizedIntegResourceRepo{resRepo1})
		require.NoError(t, err)
		unittest.AssertCount(t, &auth_model.AuthorizedIntegResourceRepo{IntegID: ai.ID}, 1)

		err = auth_model.DeleteAuthorizedIntegrationByID(t.Context(), ai.ID, ai.UserID)
		require.NoError(t, err)

		unittest.AssertNotExistsBean(t, &auth_model.AuthorizedIntegration{ID: ai.ID})
		unittest.AssertCount(t, &auth_model.AuthorizedIntegResourceRepo{IntegID: ai.ID}, 0)
	})

	t.Run("delete fails on wrong user", func(t *testing.T) {
		ai := makeAuthorizedIntegration(t)
		err := auth_model.DeleteAuthorizedIntegrationByID(t.Context(), ai.ID, 300)
		require.ErrorIs(t, err, util.ErrNotExist)
	})
}
