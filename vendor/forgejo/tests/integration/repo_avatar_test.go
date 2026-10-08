// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"bytes"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/avatar"
	app_context "forgejo.org/services/context"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoAvatar(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	session := loginUser(t, user2.Name)
	avatarURL := "/" + repo.OwnerName + "/" + repo.Name + "/settings/avatar"

	t.Run("valid avatar", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		img, err := avatar.RandomImage([]byte("seed"))
		require.NoError(t, err)

		imgData := &bytes.Buffer{}
		require.NoError(t, png.Encode(imgData, img))

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		require.NoError(t, writer.WriteField("source", "local"))
		part, err := writer.CreateFormFile("avatar", "avatar.png")
		require.NoError(t, err)
		_, err = io.Copy(part, imgData)
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		req := NewRequestWithBody(t, "POST", avatarURL, body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		session.MakeRequest(t, req, http.StatusSeeOther)

		flashCookie := session.GetCookie(app_context.CookieNameFlash)
		require.NotNil(t, flashCookie)
		assert.Contains(t, flashCookie.Value, "success")

		// Verify avatar was actually set
		repo = unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
		assert.NotEmpty(t, repo.CustomAvatarRelativePath())

		// Verify avatar is accessible
		req = NewRequest(t, "GET", "/repo-avatars/"+repo.Avatar)
		_ = session.MakeRequest(t, req, http.StatusOK)

		// Clean up
		req = NewRequest(t, "POST", avatarURL+"/delete")
		session.MakeRequest(t, req, http.StatusOK)
	})

	t.Run("no file selected", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// Simulate what a browser sends when user did not select a file
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		require.NoError(t, writer.WriteField("source", "local"))
		_, err := writer.CreateFormFile("avatar", "")
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		req := NewRequestWithBody(t, "POST", avatarURL, body)
		req.Header.Add("Content-Type", writer.FormDataContentType())
		session.MakeRequest(t, req, http.StatusSeeOther)

		// Should silently skip
		flashCookie := session.GetCookie(app_context.CookieNameFlash)
		if flashCookie != nil {
			assert.NotContains(t, flashCookie.Value, "error")
		}
	})
}
