// Copyright 2026 The Forgejo Authors
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestOrgMembersPage(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	testPage := "/org/org3/members"

	t.Run("Guest PoV", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		doc := NewHTMLParser(t, MakeRequest(t, NewRequest(t, "GET", testPage), http.StatusOK).Body)
		/* No interactive buttons - though such evaluation is easy to break in rename */
		assert.Equal(t, 0, doc.Find(".members .list .link-action").Length())
		assert.Equal(t, 0, doc.Find(".members .list .delete-button").Length())
	})

	t.Run("Member PoV", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		session := loginUser(t, "user4") // user4 is a member of org3
		doc := NewHTMLParser(t, session.MakeRequest(t, NewRequest(t, "GET", testPage), http.StatusOK).Body)
		/* Interactive buttons are only available for own entry in the list */
		assert.Equal(t, 1, doc.Find(".members .list .link-action").Length())
		assert.Equal(t, 1, doc.Find(".members .list .delete-button").Length())
	})

	t.Run("Owner PoV", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		session := loginUser(t, "user2") // user2 owns org3
		doc := NewHTMLParser(t, session.MakeRequest(t, NewRequest(t, "GET", testPage), http.StatusOK).Body)
		/* Interactive buttons are available for all entries in the list (> 2) */
		assert.Less(t, 2, doc.Find(".members .list .link-action").Length())
		assert.Less(t, 2, doc.Find(".members .list .delete-button").Length())
	})
}
