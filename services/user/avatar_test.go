// Copyright The Forgejo Authors.
// SPDX-License-Identifier: MIT

package user

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/avatar"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/storage"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type alreadyDeletedStorage struct {
	storage.DiscardStorage
}

func (s alreadyDeletedStorage) Delete(_ string) error {
	return os.ErrNotExist
}

func TestUserDeleteAvatar(t *testing.T) {
	myImage := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buff bytes.Buffer
	png.Encode(&buff, myImage)

	t.Run("AtomicStorageFailure", func(t *testing.T) {
		defer test.MockProtect[storage.ObjectStorage](&storage.Avatars)()

		require.NoError(t, unittest.PrepareTestDatabase())
		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

		err := UploadAvatar(db.DefaultContext, user, buff.Bytes())
		require.NoError(t, err)
		verification := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		assert.NotEmpty(t, verification.Avatar)

		// fail to delete ...
		storage.Avatars = storage.UninitializedStorage
		err = DeleteAvatar(db.DefaultContext, user)
		require.Error(t, err)

		// ... the avatar is not removed from the database
		verification = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		assert.True(t, verification.UseCustomAvatar)

		// already deleted ...
		storage.Avatars = alreadyDeletedStorage{}
		err = DeleteAvatar(db.DefaultContext, user)
		require.NoError(t, err)

		// ... the avatar is removed from the database
		verification = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		assert.Empty(t, verification.Avatar)
	})

	t.Run("Success", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())
		defer test.MockVariableValue(&setting.Avatar.MaxOriginSize, 3)()
		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

		err := UploadAvatar(db.DefaultContext, user, buff.Bytes())
		require.NoError(t, err)
		verification := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		assert.NotEmpty(t, verification.Avatar)
		avatarID := strings.Clone(verification.Avatar)

		// Check that resized versions are stored in the cache
		for _, size := range avatar.AllowedResizedAvatarSizes {
			_, err := storage.Avatars.Stat(fmt.Sprintf("resized/%d/%s", size, avatarID))
			require.NoError(t, err)
		}

		err = DeleteAvatar(db.DefaultContext, user)
		require.NoError(t, err)

		verification = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		assert.Empty(t, verification.Avatar)

		// The avatar is deleted from the storage
		fi, err := storage.Avatars.Stat(avatarID)
		require.Error(t, err)
		assert.Nil(t, fi)
		// All resized versions of the avatar are also deleted
		for _, size := range avatar.AllowedResizedAvatarSizes {
			fi, err := storage.Avatars.Stat(fmt.Sprintf("resized/%d/%s", size, avatarID))
			require.Error(t, err)
			assert.Nil(t, fi)
		}
	})
}

func TestUserReplaceAvatar(t *testing.T) {
	firstImage := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var firstBuff bytes.Buffer
	png.Encode(&firstBuff, firstImage)
	secondImage := image.NewRGBA(image.Rect(0, 0, 2, 2))
	secondImage.Set(0, 0, color.White)
	secondImage.Set(1, 1, color.Black)
	var secondBuff bytes.Buffer
	png.Encode(&secondBuff, secondImage)

	t.Run("Success", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())
		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

		err := UploadAvatar(db.DefaultContext, user, firstBuff.Bytes())
		require.NoError(t, err)
		user = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		firstImageHash := user.Avatar
		assert.NotEmpty(t, firstImageHash)

		err = UploadAvatar(db.DefaultContext, user, secondBuff.Bytes())
		require.NoError(t, err)
		user = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		secondImageHash := user.Avatar
		assert.NotEmpty(t, secondImageHash)
		assert.NotEqual(t, firstImageHash, secondImageHash)

		// The previous avatar is deleted from storage
		fi, err := storage.Avatars.Stat(firstImageHash)
		require.Error(t, err)
		assert.Nil(t, fi)
	})

	t.Run("ReplaceAvatarByItself", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())
		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

		err := UploadAvatar(db.DefaultContext, user, firstBuff.Bytes())
		require.NoError(t, err)
		user = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		firstImageHash := user.Avatar
		assert.NotEmpty(t, firstImageHash)

		err = UploadAvatar(db.DefaultContext, user, firstBuff.Bytes())
		require.NoError(t, err)
		user = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
		secondImageHash := user.Avatar
		assert.Equal(t, firstImageHash, secondImageHash)

		// The avatar is still available in storage
		fi, err := storage.Avatars.Stat(firstImageHash)
		require.NoError(t, err)
		assert.NotNil(t, fi)
	})
}
