// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package convert

import (
	"fmt"
	"testing"
	"time"

	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToWebAttachment(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	headRepo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	attachment := &repo_model.Attachment{
		ID:                10,
		UUID:              "uuidxxx",
		RepoID:            1,
		IssueID:           1,
		ReleaseID:         0,
		UploaderID:        0,
		CommentID:         0,
		Name:              "test.png",
		DownloadCount:     90,
		Size:              30,
		NoAutoTime:        false,
		CreatedUnix:       9342,
		CustomDownloadURL: "",
		ExternalURL:       "",
	}

	webAttachment := ToWebAttachment(headRepo, attachment)

	assert.NotNil(t, webAttachment)
	assert.Equal(t, &api.WebAttachment{
		Attachment: &api.Attachment{
			ID:            10,
			Name:          "test.png",
			Created:       time.Unix(9342, 0),
			DownloadCount: 90,
			Size:          30,
			UUID:          "uuidxxx",
			DownloadURL:   fmt.Sprintf("%sattachments/uuidxxx", setting.AppURL),
			Type:          "attachment",
		},
		MimeType: "image/png",
	}, webAttachment)
}
