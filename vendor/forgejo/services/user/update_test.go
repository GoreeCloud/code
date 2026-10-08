// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	password_module "forgejo.org/modules/auth/password"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUser(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	admin := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	require.Error(t, UpdateUser(db.DefaultContext, admin, &UpdateOptions{
		IsAdmin: optional.Some(false),
	}))

	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 28})

	opts := &UpdateOptions{
		KeepEmailPrivate:             optional.Some(false),
		FullName:                     optional.Some("Changed Name"),
		Website:                      optional.Some("https://gitea.com/"),
		Location:                     optional.Some("location"),
		Description:                  optional.Some("description"),
		AllowGitHook:                 optional.Some(true),
		AllowImportLocal:             optional.Some(true),
		MaxRepoCreation:              optional.Some(10),
		IsRestricted:                 optional.Some(true),
		IsActive:                     optional.Some(false),
		IsAdmin:                      optional.Some(true),
		Visibility:                   optional.Some(structs.VisibleTypePrivate),
		KeepActivityPrivate:          optional.Some(true),
		Language:                     optional.Some("lang"),
		Theme:                        optional.Some("theme"),
		DiffViewStyle:                optional.Some("split"),
		AllowCreateOrganization:      optional.Some(false),
		EmailNotificationsPreference: optional.Some("disabled"),
		SetLastLogin:                 true,
	}
	require.NoError(t, UpdateUser(db.DefaultContext, user, opts))

	assert.Equal(t, opts.KeepEmailPrivate.ValueOrZeroValue(), user.KeepEmailPrivate)
	assert.Equal(t, opts.FullName.ValueOrZeroValue(), user.FullName)
	assert.Equal(t, opts.Website.ValueOrZeroValue(), user.Website)
	assert.Equal(t, opts.Location.ValueOrZeroValue(), user.Location)
	assert.Equal(t, opts.Description.ValueOrZeroValue(), user.Description)
	assert.Equal(t, opts.AllowGitHook.ValueOrZeroValue(), user.AllowGitHook)
	assert.Equal(t, opts.AllowImportLocal.ValueOrZeroValue(), user.AllowImportLocal)
	assert.Equal(t, opts.MaxRepoCreation.ValueOrZeroValue(), user.MaxRepoCreation)
	assert.Equal(t, opts.IsRestricted.ValueOrZeroValue(), user.IsRestricted)
	assert.Equal(t, opts.IsActive.ValueOrZeroValue(), user.IsActive)
	assert.Equal(t, opts.IsAdmin.ValueOrZeroValue(), user.IsAdmin)
	assert.Equal(t, opts.Visibility.ValueOrZeroValue(), user.Visibility)
	assert.Equal(t, opts.KeepActivityPrivate.ValueOrZeroValue(), user.KeepActivityPrivate)
	assert.Equal(t, opts.Language.ValueOrZeroValue(), user.Language)
	assert.Equal(t, opts.Theme.ValueOrZeroValue(), user.Theme)
	assert.Equal(t, opts.DiffViewStyle.ValueOrZeroValue(), user.DiffViewStyle)
	assert.Equal(t, opts.AllowCreateOrganization.ValueOrZeroValue(), user.AllowCreateOrganization)
	assert.Equal(t, opts.EmailNotificationsPreference.ValueOrZeroValue(), user.EmailNotificationsPreference)

	user = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 28})
	assert.Equal(t, opts.KeepEmailPrivate.ValueOrZeroValue(), user.KeepEmailPrivate)
	assert.Equal(t, opts.FullName.ValueOrZeroValue(), user.FullName)
	assert.Equal(t, opts.Website.ValueOrZeroValue(), user.Website)
	assert.Equal(t, opts.Location.ValueOrZeroValue(), user.Location)
	assert.Equal(t, opts.Description.ValueOrZeroValue(), user.Description)
	assert.Equal(t, opts.AllowGitHook.ValueOrZeroValue(), user.AllowGitHook)
	assert.Equal(t, opts.AllowImportLocal.ValueOrZeroValue(), user.AllowImportLocal)
	assert.Equal(t, opts.MaxRepoCreation.ValueOrZeroValue(), user.MaxRepoCreation)
	assert.Equal(t, opts.IsRestricted.ValueOrZeroValue(), user.IsRestricted)
	assert.Equal(t, opts.IsActive.ValueOrZeroValue(), user.IsActive)
	assert.Equal(t, opts.IsAdmin.ValueOrZeroValue(), user.IsAdmin)
	assert.Equal(t, opts.Visibility.ValueOrZeroValue(), user.Visibility)
	assert.Equal(t, opts.KeepActivityPrivate.ValueOrZeroValue(), user.KeepActivityPrivate)
	assert.Equal(t, opts.Language.ValueOrZeroValue(), user.Language)
	assert.Equal(t, opts.Theme.ValueOrZeroValue(), user.Theme)
	assert.Equal(t, opts.DiffViewStyle.ValueOrZeroValue(), user.DiffViewStyle)
	assert.Equal(t, opts.AllowCreateOrganization.ValueOrZeroValue(), user.AllowCreateOrganization)
	assert.Equal(t, opts.EmailNotificationsPreference.ValueOrZeroValue(), user.EmailNotificationsPreference)
}

func TestUpdateAuth(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 28})
	userCopy := *user

	require.NoError(t, UpdateAuth(db.DefaultContext, user, &UpdateAuthOptions{
		LoginName: optional.Some("new-login"),
	}))
	assert.Equal(t, "new-login", user.LoginName)

	require.NoError(t, UpdateAuth(db.DefaultContext, user, &UpdateAuthOptions{
		Password:           optional.Some("%$DRZUVB576tfzgu"),
		MustChangePassword: optional.Some(true),
	}))
	assert.True(t, user.MustChangePassword)
	assert.NotEqual(t, userCopy.Passwd, user.Passwd)
	assert.NotEqual(t, userCopy.Salt, user.Salt)

	require.NoError(t, UpdateAuth(db.DefaultContext, user, &UpdateAuthOptions{
		ProhibitLogin: optional.Some(true),
	}))
	assert.True(t, user.ProhibitLogin)

	require.ErrorIs(t, UpdateAuth(db.DefaultContext, user, &UpdateAuthOptions{
		Password: optional.Some("aaaa"),
	}), password_module.ErrMinLength)
}
