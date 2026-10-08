// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package issues_test

import (
	"testing"

	"forgejo.org/models/issues"
	"forgejo.org/models/moderation"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
)

const (
	tsCreated timeutil.TimeStamp = timeutil.TimeStamp(1753093500) // 2025-07-21 10:25:00 UTC
	tsUpdated timeutil.TimeStamp = timeutil.TimeStamp(1753093525) // 2025-07-21 10:25:25 UTC
)

func testShadowCopyField(t *testing.T, scField moderation.ShadowCopyField, key, value string) {
	assert.Equal(t, key, scField.Key)
	assert.Equal(t, value, scField.Value)
}

func TestIssueDataGetFieldsMap(t *testing.T) {
	id := issues.IssueData{
		RepoID:         2001,
		Index:          2,
		PosterID:       1002,
		Title:          "Professional marketing services",
		Content:        "Visit my website at promote-your-business.biz for a list of available services.",
		ContentVersion: 0,
		CreatedUnix:    tsCreated,
		UpdatedUnix:    tsUpdated,
	}
	scFields := id.GetFieldsMap()

	if assert.Len(t, scFields, 8) {
		testShadowCopyField(t, scFields[0], "RepoID", "2001")
		testShadowCopyField(t, scFields[1], "Index", "2")
		testShadowCopyField(t, scFields[2], "Poster", "1002")
		testShadowCopyField(t, scFields[3], "Title", "Professional marketing services")
		testShadowCopyField(t, scFields[4], "Content", "Visit my website at promote-your-business.biz for a list of available services.")
		testShadowCopyField(t, scFields[5], "ContentVersion", "0")
		testShadowCopyField(t, scFields[6], "CreatedUnix", tsCreated.AsLocalTime().String())
		testShadowCopyField(t, scFields[7], "UpdatedUnix", tsUpdated.AsLocalTime().String())
	}
}

func TestCommentDataGetFieldsMap(t *testing.T) {
	cd := issues.CommentData{
		PosterID:       1002,
		IssueID:        3001,
		Content:        "Check out [alexsmith/website](/alexsmith/website)",
		ContentVersion: 0,
		CreatedUnix:    tsCreated,
		UpdatedUnix:    tsUpdated,
	}
	scFields := cd.GetFieldsMap()

	if assert.Len(t, scFields, 6) {
		testShadowCopyField(t, scFields[0], "Poster", "1002")
		testShadowCopyField(t, scFields[1], "IssueID", "3001")
		testShadowCopyField(t, scFields[2], "Content", "Check out [alexsmith/website](/alexsmith/website)")
		testShadowCopyField(t, scFields[3], "ContentVersion", "0")
		testShadowCopyField(t, scFields[4], "CreatedUnix", tsCreated.AsLocalTime().String())
		testShadowCopyField(t, scFields[5], "UpdatedUnix", tsUpdated.AsLocalTime().String())
	}
}
