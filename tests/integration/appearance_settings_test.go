// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"strings"
	"testing"

	"forgejo.org/modules/translation"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestThemeChange(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	locale := translation.NewLocale("en-US")
	user := loginUser(t, "user2")

	// Verify default theme
	testSelectedTheme(t, user, "forgejo-auto", "Forgejo (follow system theme)", "")

	// Change theme to forgejo-dark and verify it works fine
	testChangeTheme(t, user, "forgejo-dark")
	testSelectedTheme(t, user, "forgejo-dark", "Forgejo dark", "")

	// Change theme to gitea-dark and also verify that it's name is not translated
	testChangeTheme(t, user, "gitea-dark")
	testSelectedTheme(t, user, "gitea-dark", "gitea-dark", "")

	// Attempt to change theme to a non-existent one and verify that it wasn't
	// applied, and that there's an error message
	testChangeTheme(t, user, "non-existent")
	testSelectedTheme(t, user, "gitea-dark", "gitea-dark", locale.TrString("settings.theme_update_error"))
}

// Test UI fallback in case user's DB entry has a non-existent theme. This can
// happen if a theme was removed from the instance config at some point
func TestUserHasNonExistentTheme(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// user40 has a "frog" theme in the DB
	user := loginUser(t, "user40")

	// Verify that the UI falls back to the default theme
	testSelectedTheme(t, user, "forgejo-auto", "Forgejo (follow system theme)", "")
}

// testSelectedTheme checks that the expected theme is used in html[data-theme]
// and is default on appearance page
func testSelectedTheme(t *testing.T, session *TestSession, expectedTheme, expectedName, expectedErr string) {
	t.Helper()
	response := session.MakeRequest(t, NewRequest(t, "GET", "/user/settings/appearance"), http.StatusOK)
	page := NewHTMLParser(t, response.Body)

	dataTheme, dataThemeExists := page.Find("html").Attr("data-theme")
	assert.True(t, dataThemeExists)
	assert.Equal(t, expectedTheme, dataTheme)

	selectedTheme := page.Find("form[action='/user/settings/appearance/theme'] .menu .item.selected")
	selectorTheme, selectorThemeExists := selectedTheme.Attr("data-value")
	assert.True(t, selectorThemeExists)
	assert.Equal(t, expectedTheme, selectorTheme)
	assert.Equal(t, expectedName, strings.TrimSpace(selectedTheme.Text()))

	if expectedErr != "" {
		errMsg := strings.TrimSpace(page.Find("#flash-message").Text())
		assert.Equal(t, expectedErr, errMsg)
	}
}

// testChangeTheme changes user's theme
func testChangeTheme(t *testing.T, session *TestSession, newTheme string) {
	t.Helper()
	session.MakeRequest(t, NewRequestWithValues(t, "POST", "/user/settings/appearance/theme", map[string]string{
		"theme": newTheme,
	}), http.StatusSeeOther)
}
