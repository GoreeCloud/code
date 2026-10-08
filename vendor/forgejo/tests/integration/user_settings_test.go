// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/models/auth"
	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/container"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/modules/translation"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserSettingsAccount tests the contents of a user's account settings
// with(out) disabled user features.
func TestUserSettingsAccount(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	t.Run("all features enabled", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		doc := getHTMLDoc(t, loginUser(t, "user2"), "/user/settings/account", http.StatusOK)
		doc.AssertElement(t, "#password", true)
		doc.AssertElement(t, "#email", true)
		doc.AssertElement(t, "#delete-form", true)
	})

	t.Run("password disabled", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		disabled := container.SetOf(setting.UserFeatureManagePassword)
		defer test.MockVariableValue(&setting.Admin.UserDisabledFeatures, disabled)()
		defer test.MockVariableValue(&setting.Admin.ExternalUserDisableFeatures, disabled)()

		doc := getHTMLDoc(t, loginUser(t, "user2"), "/user/settings/account", http.StatusOK)
		doc.AssertElement(t, "#password", false)
		doc.AssertElement(t, "#email", true)
		doc.AssertElement(t, "#delete-form", true)
	})

	t.Run("deletion disabled", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		disabled := container.SetOf(setting.UserFeatureDeletion)
		defer test.MockVariableValue(&setting.Admin.UserDisabledFeatures, disabled)()
		defer test.MockVariableValue(&setting.Admin.ExternalUserDisableFeatures, disabled)()

		doc := getHTMLDoc(t, loginUser(t, "user2"), "/user/settings/account", http.StatusOK)
		doc.AssertElement(t, "#password", true)
		doc.AssertElement(t, "#email", true)
		doc.AssertElement(t, "#delete-form", false)
	})

	t.Run("deletion, password disabled", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		disabled := container.SetOf(
			setting.UserFeatureDeletion,
			setting.UserFeatureManagePassword,
		)
		defer test.MockVariableValue(&setting.Admin.UserDisabledFeatures, disabled)()
		defer test.MockVariableValue(&setting.Admin.ExternalUserDisableFeatures, disabled)()

		doc := getHTMLDoc(t, loginUser(t, "user2"), "/user/settings/account", http.StatusOK)
		doc.AssertElement(t, "#password", false)
		doc.AssertElement(t, "#email", true)
		doc.AssertElement(t, "#delete-form", false)
	})
}

// TestUserSettingsUpdatePassword tests updating a user's password with(out)
// disabled user features.
func TestUserSettingsUpdatePassword(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	t.Run("password enabled", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// changing password should work
		session := loginUser(t, "user2")
		req := NewRequestWithValues(t, "POST", "/user/settings/account", map[string]string{
			"old_password": "password",
			"password":     "password",
			"retype":       "password",
		})
		session.MakeRequest(t, req, http.StatusSeeOther)
	})

	t.Run("password disabled", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		disabled := container.SetOf(setting.UserFeatureManagePassword)
		defer test.MockVariableValue(&setting.Admin.UserDisabledFeatures, disabled)()
		defer test.MockVariableValue(&setting.Admin.ExternalUserDisableFeatures, disabled)()

		// changing password should not work
		session := loginUser(t, "user2")
		req := NewRequestWithValues(t, "POST", "/user/settings/account", map[string]string{
			"old_password": "password",
			"password":     "password",
			"retype":       "password",
		})
		session.MakeRequest(t, req, http.StatusNotFound)
	})
}

// TestUserSettingsDelete tests deleting a user with(out) disabled user
// features.
func TestUserSettingsDelete(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	t.Run("deletion disabled", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		disabled := container.SetOf(setting.UserFeatureDeletion)
		defer test.MockVariableValue(&setting.Admin.UserDisabledFeatures, disabled)()
		defer test.MockVariableValue(&setting.Admin.ExternalUserDisableFeatures, disabled)()

		// deleting user should not work
		session := loginUser(t, "user2")
		req := NewRequest(t, "POST", "/user/settings/account/delete")
		session.MakeRequest(t, req, http.StatusNotFound)
	})
}

func TestUserRename(t *testing.T) {
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestUserRename")()
	defer tests.PrepareTestEnv(t)()

	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	session := loginUser(t, "user2")
	trMsg := translation.NewLocale("en-US").Tr("settings.password_username_disabled")

	test := func(t *testing.T, session *TestSession, allowed bool) {
		t.Helper()

		resp := session.MakeRequest(t, NewRequest(t, "GET", "/user/settings"), http.StatusOK)
		if allowed {
			assert.NotContains(t, resp.Body.String(), trMsg)
		} else {
			assert.Contains(t, resp.Body.String(), trMsg)
		}
	}

	t.Run("Local", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		test(t, session, true)
	})

	t.Run("OAuth2", func(t *testing.T) {
		user.LoginSource = 1001
		user.LoginType = auth.OAuth2
		_, err := db.GetEngine(t.Context()).Cols("login_source", "login_type").Update(user)
		require.NoError(t, err)

		t.Run("Not allowed", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			test(t, session, false)
		})

		user.LoginSource = 1002
		_, err = db.GetEngine(t.Context()).Cols("login_source", "login_type").Update(user)
		require.NoError(t, err)

		t.Run("Allowed", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			test(t, session, true)
		})
	})
}
