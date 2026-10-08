// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package user_test

import (
	"testing"

	"forgejo.org/models/moderation"
	"forgejo.org/models/user"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
)

const (
	tsCreated   timeutil.TimeStamp = timeutil.TimeStamp(1753093200) // 2025-07-21 10:20:00 UTC
	tsUpdated   timeutil.TimeStamp = timeutil.TimeStamp(1753093320) // 2025-07-21 10:22:00 UTC
	tsLastLogin timeutil.TimeStamp = timeutil.TimeStamp(1753093800) // 2025-07-21 10:30:00 UTC
)

func testShadowCopyField(t *testing.T, scField moderation.ShadowCopyField, key, value string) {
	assert.Equal(t, key, scField.Key)
	assert.Equal(t, value, scField.Value)
}

func TestUserDataGetFieldsMap(t *testing.T) {
	ud := user.UserData{
		Name:        "alexsmith",
		FullName:    "Alex Smith",
		Email:       "alexsmith@example.org",
		LoginName:   "",
		Location:    "@master@seo.net",
		Website:     "http://promote-your-business.biz",
		Pronouns:    "SEO",
		Description: "I can help you promote your business online using SEO.",
		CreatedUnix: tsCreated,
		UpdatedUnix: tsUpdated,
		LastLogin:   tsLastLogin,
		Avatar:      "avatar-hash-user-1002",
		AvatarEmail: "alexsmith@example.org",
	}
	scFields := ud.GetFieldsMap()

	if assert.Len(t, scFields, 13) {
		testShadowCopyField(t, scFields[0], "Name", "alexsmith")
		testShadowCopyField(t, scFields[1], "FullName", "Alex Smith")
		testShadowCopyField(t, scFields[2], "Email", "alexsmith@example.org")
		testShadowCopyField(t, scFields[3], "LoginName", "")
		testShadowCopyField(t, scFields[4], "Location", "@master@seo.net")
		testShadowCopyField(t, scFields[5], "Website", "http://promote-your-business.biz")
		testShadowCopyField(t, scFields[6], "Pronouns", "SEO")
		testShadowCopyField(t, scFields[7], "Description", "I can help you promote your business online using SEO.")
		testShadowCopyField(t, scFields[8], "CreatedUnix", tsCreated.AsLocalTime().String())
		testShadowCopyField(t, scFields[9], "UpdatedUnix", tsUpdated.AsLocalTime().String())
		testShadowCopyField(t, scFields[10], "LastLogin", tsLastLogin.AsLocalTime().String())
		testShadowCopyField(t, scFields[11], "Avatar", "avatar-hash-user-1002")
		testShadowCopyField(t, scFields[12], "AvatarEmail", "alexsmith@example.org")
	}
}
