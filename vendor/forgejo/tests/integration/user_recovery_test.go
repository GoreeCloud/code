// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/test"
	"forgejo.org/modules/translation"
	"forgejo.org/services/mailer"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestForgotPassword(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	test := func(t *testing.T, user *user_model.User, email *user_model.EmailAddress) {
		t.Helper()

		called := false
		defer test.MockVariableValue(&mailer.SendAsync, func(msgs ...*mailer.Message) {
			assert.Len(t, msgs, 1)
			assert.Equal(t, user.EmailTo(), msgs[0].To)
			assert.EqualValues(t, translation.NewLocale("en-US").Tr("mail.reset_password"), msgs[0].Subject)
			assert.Contains(t, msgs[0].Body, translation.NewLocale("en-US").Tr("mail.reset_password.text", "3 hours"))
			called = true
		})()

		req := NewRequestWithValues(t, "POST", "/user/forgot_password", map[string]string{
			"email": email.Email,
		})
		MakeRequest(t, req, http.StatusOK)

		assert.True(t, called)
	}
	t.Run("Unactivated email address", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		test(t, unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 11}), unittest.AssertExistsAndLoadBean(t, &user_model.EmailAddress{UID: 11}, "is_activated = false"))
	})

	t.Run("Activated email address", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		test(t, unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 12}), unittest.AssertExistsAndLoadBean(t, &user_model.EmailAddress{UID: 12}, "is_activated = true"))
	})
}
