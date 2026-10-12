// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/models/user"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/storage"
	"forgejo.org/modules/test"
	doctor "forgejo.org/services/doctor"
	repo_service "forgejo.org/services/repository"
	user_service "forgejo.org/services/user"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateUserAvatar(ctx context.Context, t *testing.T, user *user.User) string {
	err := user_service.UploadAvatar(ctx, user, generateAvatar(user.ID))
	require.NoError(t, err)
	assert.NotEmpty(t, user.Avatar)
	return user.CustomAvatarRelativePath()
}

func generateRepoAvatar(ctx context.Context, t *testing.T, repo *repo.Repository) string {
	err := repo_service.UploadAvatar(ctx, repo, generateAvatar(repo.ID))
	require.NoError(t, err)
	assert.NotEmpty(t, repo.Avatar)
	return repo.CustomAvatarRelativePath()
}

func generateAvatar(objectID int64) []byte {
	avatar := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	// make the avatar distinctive for the user by setting a pixel based on the object id
	avatar.SetRGBA(3, 3, color.RGBA{R: uint8(objectID % 255), G: uint8(objectID / 255 % 255), B: 55, A: 66})
	var buff bytes.Buffer
	png.Encode(&buff, avatar)
	return buff.Bytes()
}

func TestRemoveUnusedUserAvatars(t *testing.T) {
	defer tests.PrepareTestEnv(t, 1)()
	// make the maximum uncached image size small, so that our test image is bigger than that
	defer test.MockVariableValue(&setting.Avatar.MaxOriginSize, 3)()
	var err error

	ctx := db.DefaultContext

	user2 := unittest.AssertExistsAndLoadBean(t, &user.User{ID: 2})
	user3 := unittest.AssertExistsAndLoadBean(t, &user.User{ID: 3})
	avatarPathUser2 := generateUserAvatar(ctx, t, user2)
	avatarPathUser3 := generateUserAvatar(ctx, t, user3)

	// disconnect the avatar from user2, keeping the avatar for user3
	user2.Avatar = ""
	err = user.UpdateUserCols(ctx, user2, "avatar")
	require.NoError(t, err)

	// make sure both avatars are still stored as a file
	_, err = storage.Avatars.Stat(avatarPathUser2)
	require.NoError(t, err)
	_, err = storage.Avatars.Stat(avatarPathUser3)
	require.NoError(t, err)
	// the downscaled versions are also stored
	_, err = storage.Avatars.Stat(fmt.Sprintf("resized/64/%s", avatarPathUser2))
	require.NoError(t, err)
	_, err = storage.Avatars.Stat(fmt.Sprintf("resized/64/%s", avatarPathUser3))
	require.NoError(t, err)

	doctor.CheckStorage(&doctor.CheckStorageOptions{Avatars: true})(ctx, log.GetLogger("doctor"), true)

	// the avatar for user2 is no longer stored
	_, err = storage.Avatars.Stat(avatarPathUser2)
	require.Error(t, err)
	// its downscaled versions are not stored either
	_, err = storage.Avatars.Stat(fmt.Sprintf("resized/64/%s", avatarPathUser2))
	require.Error(t, err)
	// the avatar for user3 is still stored stored
	_, err = storage.Avatars.Stat(avatarPathUser3)
	require.NoError(t, err)
	// the downscaled versions are also stored
	_, err = storage.Avatars.Stat(fmt.Sprintf("resized/64/%s", avatarPathUser3))
	require.NoError(t, err)
}

func TestRemoveUnusedRepoAvatars(t *testing.T) {
	defer tests.PrepareTestEnv(t, 1)()
	// make the maximum uncached image size small, so that our test image is bigger than that
	defer test.MockVariableValue(&setting.Avatar.MaxOriginSize, 3)()
	var err error

	ctx := db.DefaultContext

	repo2 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 2})
	repo3 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 3})
	avatarPathRepo2 := generateRepoAvatar(ctx, t, repo2)
	avatarPathRepo3 := generateRepoAvatar(ctx, t, repo3)

	// disconnect the avatar from repo2, but keep it for repo3
	repo2.Avatar = ""
	err = repo.UpdateRepositoryCols(ctx, repo2, "avatar")
	require.NoError(t, err)

	// make sure the avatar is still stored as a file
	_, err = storage.RepoAvatars.Stat(avatarPathRepo2)
	require.NoError(t, err)
	// the downscaled versions are also stored
	_, err = storage.RepoAvatars.Stat(fmt.Sprintf("resized/64/%s", avatarPathRepo2))
	require.NoError(t, err)

	doctor.CheckStorage(&doctor.CheckStorageOptions{RepoAvatars: true})(ctx, log.GetLogger("doctor"), true)

	// the avatar for repo2 is no longer stored
	_, err = storage.RepoAvatars.Stat(avatarPathRepo2)
	require.Error(t, err)
	// its downscaled versions are not stored either
	_, err = storage.RepoAvatars.Stat(fmt.Sprintf("resized/64/%s", avatarPathRepo2))
	require.Error(t, err)
	// the avatar for repo3 is still stored
	_, err = storage.RepoAvatars.Stat(avatarPathRepo3)
	require.NoError(t, err)
	// its downscaled versions are also still stored
	_, err = storage.RepoAvatars.Stat(fmt.Sprintf("resized/64/%s", avatarPathRepo3))
	require.NoError(t, err)
}
