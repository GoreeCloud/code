// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"strings"
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/translation"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestPullEditable_ShowEditableLabel(t *testing.T) {
	// This fixture loads a PR which is made from a different repository,
	// and opened by the user who owns the fork (which is necessary for
	// them to be allowed to set the PR as editable).
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestPullEditable")()
	defer tests.PrepareTestEnv(t)()

	t.Run("Show editable label if PR is editable", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		editable := true

		setPREditable(t, editable)
		testEditableLabelShown(t, editable)
	})

	t.Run("Don't show editable label if PR is not editable", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		editable := false

		setPREditable(t, editable)
		testEditableLabelShown(t, editable)
	})
}

func setPREditable(t *testing.T, editable bool) {
	t.Helper()
	session := loginUser(t, "user13")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	req := NewRequestWithJSON(t, "PATCH", "/api/v1/repos/user12/repo10/pulls/2", &api.EditPullRequestOption{
		AllowMaintainerEdit: &editable,
	}).AddTokenAuth(token)
	session.MakeRequest(t, req, http.StatusCreated)
}

func testEditableLabelShown(t *testing.T, expectLabel bool) {
	t.Helper()
	session := loginUser(t, "user12")
	req := NewRequest(t, "GET", "/user12/repo10/pulls/2")
	resp := session.MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)
	htmlDoc.AssertElement(t, "#editable-label", expectLabel)
	locale := translation.NewLocale("en-US")
	if expectLabel {
		sidebarText := htmlDoc.Find(".issue-content-right span.maintainers-can-edit-status").First().Text()
		assert.Equal(t, locale.TrString("repo.pulls.maintainers_can_edit"), strings.TrimSpace(sidebarText))
	} else {
		sidebarText := htmlDoc.Find(".issue-content-right span.maintainers-can-edit-status").First().Text()
		assert.Equal(t, locale.TrString("repo.pulls.maintainers_cannot_edit"), strings.TrimSpace(sidebarText))
	}
}
