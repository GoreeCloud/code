// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/test"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func getOrgSettingsFormData(t *testing.T, session *TestSession, orgName string) map[string]string {
	return map[string]string{
		"name":                          orgName,
		"full_name":                     "",
		"email":                         "",
		"description":                   "",
		"website":                       "",
		"location":                      "",
		"visibility":                    "0",
		"repo_admin_change_team_access": "on",
		"max_repo_creation":             "-1",
	}
}

func getOrgSettings(t *testing.T, token, orgName string) *api.Organization {
	t.Helper()

	req := NewRequestf(t, "GET", "/api/v1/orgs/%s", orgName).AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)

	var org *api.Organization
	DecodeJSON(t, resp, &org)

	return org
}

func TestOrgSettingsChangeEmail(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	const orgName = "org3"
	settingsURL := fmt.Sprintf("/org/%s/settings", orgName)

	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadOrganization)

	t.Run("Invalid", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		settings := getOrgSettingsFormData(t, session, orgName)

		settings["email"] = "invalid"
		doc := NewHTMLParser(t, session.MakeRequest(t, NewRequestWithValues(t, "POST", settingsURL, settings), http.StatusOK).Body)
		doc.AssertElement(t, ".status-page-500", false)
		doc.AssertElement(t, ".flash-error", true)

		org := getOrgSettings(t, token, orgName)
		assert.Equal(t, "org3@example.com", org.Email)
	})

	t.Run("Valid", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		settings := getOrgSettingsFormData(t, session, orgName)

		settings["email"] = "example@example.com"
		doc := NewHTMLParser(t, session.MakeRequest(t, NewRequestWithValues(t, "POST", settingsURL, settings), http.StatusSeeOther).Body)
		doc.AssertElement(t, "body", true)
		doc.AssertElement(t, ".status-page-500", false)
		doc.AssertElement(t, ".flash-error", false)

		org := getOrgSettings(t, token, orgName)
		assert.Equal(t, "example@example.com", org.Email)
	})

	t.Run("Empty", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		settings := getOrgSettingsFormData(t, session, orgName)

		settings["email"] = ""
		doc := NewHTMLParser(t, session.MakeRequest(t, NewRequestWithValues(t, "POST", settingsURL, settings), http.StatusSeeOther).Body)
		doc.AssertElement(t, "body", true)
		doc.AssertElement(t, ".status-page-500", false)
		doc.AssertElement(t, ".flash-error", false)

		org := getOrgSettings(t, token, orgName)
		assert.Empty(t, org.Email)
	})
}

func TestOrgSettingsUpdateWebsite(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	const orgName = "org3"
	urlStr := fmt.Sprintf("/org/%s/settings", orgName)
	session := loginUser(t, "user1")

	t.Run("an HTTPS website under default schemes", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// changing website should work
		req := NewRequestWithValues(t, "POST", urlStr, map[string]string{
			"name":    orgName,
			"website": "https://codeberg.org",
		})
		resp := session.MakeRequest(t, req, http.StatusSeeOther)
		assertHasFlashMessages(t, resp, "success")
	})

	t.Run("an H3 website under default schemes", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		// changing website should not work
		req := NewRequestWithValues(t, "POST", urlStr, map[string]string{
			"name":    orgName,
			"website": "h3://codeberg.org",
		})
		resp := session.MakeRequest(t, req, http.StatusOK)
		doc := NewHTMLParser(t, resp.Body)
		flash := doc.Find("#flash-message").Text()
		assert.Equal(t, `Website"Url" is not a valid URL.`, strings.TrimSpace(flash))
	})

	t.Run("an H3 website under custom schemes", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		defer test.MockProtect(&setting.Service.ValidSiteURLSchemes)()
		setting.Service.ValidSiteURLSchemes = append(setting.Service.ValidSiteURLSchemes, "h3")

		// changing website should work
		req := NewRequestWithValues(t, "POST", urlStr, map[string]string{
			"name":    orgName,
			"website": "h3://codeberg.org",
		})
		resp := session.MakeRequest(t, req, http.StatusSeeOther)
		assertHasFlashMessages(t, resp, "success")
	})
}
