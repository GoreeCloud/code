// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"context"
	"fmt"

	"forgejo.org/models/db"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/avatar"
	"forgejo.org/modules/avatarstore"
	"forgejo.org/modules/log"
	"forgejo.org/modules/storage"
)

// UploadAvatar saves custom avatar for user.
func UploadAvatar(ctx context.Context, u *user_model.User, data []byte) error {
	avatarData, img, err := avatar.ProcessAvatarImage(data)
	if err != nil {
		return err
	}

	ctx, committer, err := db.TxContext(ctx)
	if err != nil {
		return err
	}
	defer committer.Close()

	u.UseCustomAvatar = true
	previousAvatar := u.Avatar
	u.Avatar = avatar.HashAvatar(u.ID, data)
	if err = user_model.UpdateUserCols(ctx, u, "use_custom_avatar", "avatar"); err != nil {
		return fmt.Errorf("updateUser: %w", err)
	}

	if len(previousAvatar) > 0 {
		err := avatarstore.DeleteAvatar(previousAvatar, storage.Avatars)
		if err != nil {
			return err
		}
	}
	if err := avatarstore.StoreAvatar(u.CustomAvatarRelativePath(), avatarData, img, storage.Avatars); err != nil {
		return fmt.Errorf("Failed to store avatar at %s: %w", u.CustomAvatarRelativePath(), err)
	}

	return committer.Commit()
}

// DeleteAvatar deletes the user's custom avatar.
func DeleteAvatar(ctx context.Context, u *user_model.User) error {
	aPath := u.CustomAvatarRelativePath()
	log.Trace("DeleteAvatar[%d]: %s", u.ID, aPath)

	return db.WithTx(ctx, func(ctx context.Context) error {
		hasAvatar := len(u.Avatar) > 0
		u.UseCustomAvatar = false
		u.Avatar = ""
		if _, err := db.GetEngine(ctx).ID(u.ID).Cols("avatar, use_custom_avatar").Update(u); err != nil {
			return fmt.Errorf("DeleteAvatar: %w", err)
		}

		if hasAvatar {
			return avatarstore.DeleteAvatar(aPath, storage.Avatars)
		}
		return nil
	})
}
