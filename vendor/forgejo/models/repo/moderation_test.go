// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package repo_test

import (
	"testing"

	"forgejo.org/models/moderation"
	"forgejo.org/models/repo"
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

func TestRepositoryDataGetFieldsMap(t *testing.T) {
	rd := repo.RepositoryData{
		OwnerID:     1002,
		OwnerName:   "alexsmith",
		Name:        "website",
		Description: "My static website.",
		Website:     "http://promote-your-business.biz",
		Topics:      []string{"bulk-email", "email-services"},
		Avatar:      "avatar-hash-repo-2002",
		CreatedUnix: tsCreated,
		UpdatedUnix: tsUpdated,
	}
	scFields := rd.GetFieldsMap()

	if assert.Len(t, scFields, 9) {
		testShadowCopyField(t, scFields[0], "OwnerID", "1002")
		testShadowCopyField(t, scFields[1], "OwnerName", "alexsmith")
		testShadowCopyField(t, scFields[2], "Name", "website")
		testShadowCopyField(t, scFields[3], "Description", "My static website.")
		testShadowCopyField(t, scFields[4], "Website", "http://promote-your-business.biz")
		testShadowCopyField(t, scFields[5], "Topics", "bulk-email, email-services")
		testShadowCopyField(t, scFields[6], "Avatar", "avatar-hash-repo-2002")
		testShadowCopyField(t, scFields[7], "CreatedUnix", tsCreated.AsLocalTime().String())
		testShadowCopyField(t, scFields[8], "UpdatedUnix", tsUpdated.AsLocalTime().String())
	}
}
