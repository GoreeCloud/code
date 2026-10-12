// Copyright 2018 The Gitea Authors. All rights reserved.
// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forms

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/translation"
	"forgejo.org/modules/typesniffer"
)

// ReadAvatar reads and validates an avatar from a multipart file header.
func ReadAvatar(header *multipart.FileHeader, locale translation.Locale) ([]byte, error) {
	if header == nil || header.Filename == "" {
		return nil, nil
	}

	r, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("Avatar.Open: %w", err)
	}
	defer r.Close()

	if header.Size > setting.Avatar.MaxFileSize {
		return nil, errors.New(locale.TrString("settings.uploaded_avatar_is_too_big", header.Size/1024, setting.Avatar.MaxFileSize/1024))
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	st := typesniffer.DetectContentType(data, "")
	if !st.IsImage() || st.IsSvgImage() {
		return nil, errors.New(locale.TrString("settings.uploaded_avatar_not_a_image"))
	}

	return data, nil
}
