// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"testing"

	cmd "forgejo.org/cmd"
	"forgejo.org/models/db"
	"forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/models/user"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/storage"
	"forgejo.org/modules/test"
	"forgejo.org/tests"

	"github.com/stretchr/testify/require"
)

func TestPrecomputeUserAvatars(t *testing.T) {
	defer tests.PrepareTestEnv(t, 1)()
	var err error

	// make the maximum uncached image size small, so that our test image is bigger than that
	defer test.MockVariableValue(&setting.Avatar.MaxOriginSize, 3)()

	ctx := db.DefaultContext

	u := unittest.AssertExistsAndLoadBean(t, &user.User{ID: 2})
	// generate an avatar for this user
	myImage := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	var buff bytes.Buffer
	png.Encode(&buff, myImage)
	// store the avatar, but only the original (not the resized versions)
	avatarPath := "some_id"
	storage.Avatars.Save(avatarPath, bytes.NewReader(buff.Bytes()), -1)
	u.UseCustomAvatar = true
	u.Avatar = avatarPath
	err = user.UpdateUserCols(ctx, u, "use_custom_avatar", "avatar")
	require.NoError(t, err)

	// run the doctor command
	err = cmd.RunAvatarResize(ctx, true, false)
	require.NoError(t, err)

	// the resized version of the avatar is now stored in the cache
	_, err = storage.Avatars.Stat(fmt.Sprintf("resized/64/%s", avatarPath))
	require.NoError(t, err)
}

func TestPrecomputeRepoAvatars(t *testing.T) {
	defer tests.PrepareTestEnv(t, 1)()
	var err error
	// make the maximum uncached image size small, so that our test image is bigger than that
	defer test.MockVariableValue(&setting.Avatar.MaxOriginSize, 3)()

	ctx := db.DefaultContext

	u := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 2})
	// generate an avatar for this repo
	myImage := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	var buff bytes.Buffer
	png.Encode(&buff, myImage)
	// store the avatar, but only the original (not the resized versions)
	avatarPath := "some_id"
	storage.RepoAvatars.Save(avatarPath, bytes.NewReader(buff.Bytes()), -1)
	u.Avatar = avatarPath
	err = repo.UpdateRepositoryCols(ctx, u, "avatar")
	require.NoError(t, err)
	// make sure there is no resized version of the avatar in the storage yet
	storage.RepoAvatars.Delete(fmt.Sprintf("resized/64/%s", avatarPath))
	_, err = storage.RepoAvatars.Stat(fmt.Sprintf("resized/64/%s", avatarPath))
	require.Error(t, err)

	// run the doctor command
	err = cmd.RunAvatarResize(ctx, false, true)
	require.NoError(t, err)

	// the resized version of the avatar is now stored in the cache
	_, err = storage.RepoAvatars.Stat(fmt.Sprintf("resized/64/%s", avatarPath))
	require.NoError(t, err)
}

func TestPrecomputeAvatarsWithoutArgument(t *testing.T) {
	defer tests.PrepareTestEnv(t, 1)()
	var err error

	ctx := db.DefaultContext

	// run the doctor command
	err = cmd.RunAvatarResize(ctx, false, false)

	// there is an error because we didn't specify which avatars to resize
	require.Error(t, err)
}
