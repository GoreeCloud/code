// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestAdminViewRepos(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")
	req := NewRequest(t, "GET", "/admin/repos?q=&sort=reversealphabetically")
	resp := session.MakeRequest(t, req, http.StatusOK)

	htmlDoc := NewHTMLParser(t, resp.Body)

	// Should be 50 rows of repositories rendered; this is the page size, and there are 65 repos in-fixture.
	assert.Equal(t, 50, htmlDoc.Find("table tbody tr").Length())

	// Check for a specific repo link to see if it is rendered correctly
	link := htmlDoc.Find("table tbody tr td a[href='/user27/repo49']")
	assert.Equal(t, 1, link.Length())
	assert.Equal(t, "repo49", link.Text())
}
