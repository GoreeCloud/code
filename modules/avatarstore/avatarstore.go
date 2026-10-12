// Copyright 2026 the Forgejo authors. All rights reserved.
// SPDX-License-Identifier: MIT

package avatarstore

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"forgejo.org/modules/avatar"
	"forgejo.org/modules/log"
	"forgejo.org/modules/storage"

	"golang.org/x/image/draw"
)

// StoreAvatar stores an avatar in an object store and precomputes resized versions of it
func StoreAvatar(avatarPath string, avatarData []byte, avatarImg image.Image, imgStore storage.ObjectStorage) error {
	err := storage.SaveFrom(imgStore, avatarPath, func(w io.Writer) error {
		_, err := w.Write(avatarData)
		return err
	})
	if err != nil {
		return err
	}
	if avatarImg != nil {
		// pre-compute rescaled versions of the avatar
		return PrecomputeResizedAvatars(imgStore, avatarImg, avatarPath)
	}
	return nil
}

// PrecomputeResizedAvatars computes resized versions of the avatars and stores them in the cache
func PrecomputeResizedAvatars(resizedStore storage.ObjectStorage, avatarImg image.Image, avatarPath string) error {
	for _, size := range avatar.AllowedResizedAvatarSizes {
		rescaled := avatar.Scale(avatarImg, size, size, draw.BiLinear)
		err := storage.SaveFrom(resizedStore, fmt.Sprintf("resized/%d/%s", size, avatarPath), func(w io.Writer) error {
			return png.Encode(w, rescaled)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAvatar removes an avatar file from both the original and resized storage
func DeleteAvatar(avatarPath string, imgStore storage.ObjectStorage) error {
	if err := imgStore.Delete(avatarPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove %s: %w", avatarPath, err)
		}
		log.Warn("Deleting avatar %s but it doesn't exist", avatarPath)
	}
	for _, size := range avatar.AllowedResizedAvatarSizes {
		err := imgStore.Delete(fmt.Sprintf("resized/%d/%s", size, avatarPath))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove resized avatar resized/%d/%s: %w", size, avatarPath, err)
		}
	}
	return nil
}
