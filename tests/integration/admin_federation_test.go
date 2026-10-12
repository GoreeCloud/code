// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/routers"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestAdminFederationViewHostsAndUsers(t *testing.T) {
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestAdminFederationViewHostsAndUsers")()
	defer tests.PrepareTestEnv(t)()

	t.Run("Federation enabled", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Federation.Enabled, true)()
		defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

		t.Run("Anonymous user", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			req := NewRequest(t, "GET", "/admin/federation/hosts")
			MakeRequest(t, req, http.StatusSeeOther)
			req = NewRequest(t, "GET", "/admin/federation/hosts/1")
			MakeRequest(t, req, http.StatusSeeOther)
			req = NewRequest(t, "GET", "/admin/federation/users")
			MakeRequest(t, req, http.StatusSeeOther)
		})

		t.Run("Normal user", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			session := loginUser(t, "user2")

			req := NewRequest(t, "GET", "/admin/federation/hosts")
			session.MakeRequest(t, req, http.StatusForbidden)
			req = NewRequest(t, "GET", "/admin/federation/hosts/1")
			session.MakeRequest(t, req, http.StatusForbidden)
			req = NewRequest(t, "GET", "/admin/federation/users")
			session.MakeRequest(t, req, http.StatusForbidden)
		})

		t.Run("Admin user", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			session := loginUser(t, "user1")

			req := NewRequest(t, "GET", "/admin/federation/hosts")
			resp := session.MakeRequest(t, req, http.StatusOK)
			htmlDoc := NewHTMLParser(t, resp.Body)
			hostRows := htmlDoc.Find(".admin-setting-content table tbody tr")
			assert.Equal(t, 2, hostRows.Length())
			assert.Contains(t, hostRows.Text(), "bob.example.com")
			assert.Contains(t, hostRows.Text(), "alice.example.com")

			req = NewRequest(t, "GET", "/admin/federation/users")
			resp = session.MakeRequest(t, req, http.StatusOK)
			htmlDoc = NewHTMLParser(t, resp.Body)
			userRows := htmlDoc.Find(".admin-setting-content table tbody tr")
			assert.Equal(t, 3, userRows.Length())
			assert.Contains(t, userRows.Text(), "/api/v1/activitypub/user-id/1/inbox#bob")
			assert.Contains(t, userRows.Text(), "/api/v1/activitypub/user-id/1/inbox#alice")
			assert.Contains(t, userRows.Text(), "/api/v1/activitypub/user-id/2/inbox#eve")

			req = NewRequest(t, "GET", "/admin/federation/hosts/1")
			resp = session.MakeRequest(t, req, http.StatusOK)
			htmlDoc = NewHTMLParser(t, resp.Body)
			userRows = htmlDoc.Find(".admin-setting-content table tbody tr")
			assert.Equal(t, 1, userRows.Length())
			assert.Contains(t, userRows.Text(), "/api/v1/activitypub/user-id/1/inbox#bob")

			req = NewRequest(t, "GET", "/admin/federation/hosts/2")
			resp = session.MakeRequest(t, req, http.StatusOK)
			htmlDoc = NewHTMLParser(t, resp.Body)
			userRows = htmlDoc.Find(".admin-setting-content table tbody tr")
			assert.Equal(t, 2, userRows.Length())
			assert.Contains(t, userRows.Text(), "/api/v1/activitypub/user-id/1/inbox#alice")
			assert.Contains(t, userRows.Text(), "/api/v1/activitypub/user-id/2/inbox#eve")
		})
	})

	t.Run("Federation disabled", func(t *testing.T) {
		t.Run("Anonymous user", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			req := NewRequest(t, "GET", "/admin/federation/hosts")
			MakeRequest(t, req, http.StatusNotFound)
			req = NewRequest(t, "GET", "/admin/federation/hosts/1")
			MakeRequest(t, req, http.StatusNotFound)
			req = NewRequest(t, "GET", "/admin/federation/users")
			MakeRequest(t, req, http.StatusNotFound)
		})

		t.Run("Normal user", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			session := loginUser(t, "user2")

			req := NewRequest(t, "GET", "/admin/federation/hosts")
			session.MakeRequest(t, req, http.StatusNotFound)
			req = NewRequest(t, "GET", "/admin/federation/hosts/1")
			session.MakeRequest(t, req, http.StatusNotFound)
			req = NewRequest(t, "GET", "/admin/federation/users")
			session.MakeRequest(t, req, http.StatusNotFound)
		})

		t.Run("Admin user", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			session := loginUser(t, "user1")

			req := NewRequest(t, "GET", "/admin/federation/hosts")
			session.MakeRequest(t, req, http.StatusNotFound)
			req = NewRequest(t, "GET", "/admin/federation/hosts/1")
			session.MakeRequest(t, req, http.StatusNotFound)
			req = NewRequest(t, "GET", "/admin/federation/users")
			session.MakeRequest(t, req, http.StatusNotFound)
		})
	})
}
