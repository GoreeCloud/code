// Copyright 2026 The Forgejo Authors
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"strings"
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

/* TestFooterContent asserts go tmpl logic of footer */
func TestFooterContent(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// The footer can be tested on any page, but preferably a lightweight one
	regularPage := "/explore/organizations"
	adminPage := "/admin"

	adminUser := loginUser(t, "user1")
	regularUser := loginUser(t, "user2")

	t.Run(`Default configuration`, func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// For admins Version dubs as a link to /admin/config, for regular users
		// it's just text
		testFooterContent(t, regularUser, regularPage, true, true, false)
		testFooterContent(t, adminUser, regularPage, true, true, true)
		testFooterContent(t, adminUser, adminPage, true, true, true)
	})

	t.Run(`ShowFooterPoweredBy is disabled`, func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		defer test.MockVariableValue(&setting.Other.ShowFooterPoweredBy, false)()

		testFooterContent(t, regularUser, regularPage, false, true, false)
	})

	t.Run(`ShowFooterVersion is disabled`, func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		defer test.MockVariableValue(&setting.Other.ShowFooterVersion, false)()

		// ShowFooterVersion is on by default. Disabling it hides the version on
		// all pages but not on /admin
		testFooterContent(t, regularUser, regularPage, true, false, false)
		testFooterContent(t, adminUser, regularPage, true, false, false)
		testFooterContent(t, adminUser, adminPage, true, true, true)
	})
}

// When visiting certain pages, the corresponding entry of user menu is highlighted
func testFooterContent(t *testing.T, session *TestSession, url string, expectPoweredBy, expectVersion, expectVersionLink bool) {
	page := NewHTMLParser(t, session.MakeRequest(t, NewRequest(t, "GET", url), http.StatusOK).Body)
	// AssertElement will only pass if there's just one such element
	footerLeft := page.Find("footer .left-links").Text()
	assert.Equal(t, expectPoweredBy, strings.Contains(footerLeft, "Powered by Forgejo"))
	assert.Equal(t, expectVersion, strings.Contains(footerLeft, "Version"))
	page.AssertElement(t, `footer .left-links a[href="/admin/config"]`, expectVersionLink)
}
