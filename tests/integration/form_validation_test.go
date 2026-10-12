// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"strings"
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/tests"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestWebsitePattern(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")

	cases := [][2]string{
		// these routes have a "Website" form input whose validation changes based on `[service].VALID_SITE_URL_SCHEMES`:
		{"admin user edit", "/admin/users/2/edit"},
		{"org settings", "/org/org3/settings"},
		{"repo settings", "/user2/repo1/settings"},
		{"user own settings", "/user/settings"},
	}
	for _, testCase := range cases {
		title := testCase[0]
		urlStr := testCase[1]

		t.Run(title+" form under default schemes", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()

			// Website pattern and note should contain only http(s)
			req := NewRequest(t, "GET", urlStr)
			resp := session.MakeRequest(t, req, http.StatusOK)
			doc := NewHTMLParser(t, resp.Body)
			doc.AssertAttrEqual(t, "input[type=url][name=website]", "pattern", `(http|https)://.+`)
			doc.AssertElementPredicate(t, "input[type=url][name=website] + .help", func(element *goquery.Selection) {
				assert.Equal(t, "Allowed URL schemes include: http, https", strings.TrimSpace(element.Text()))
			})
		})

		t.Run(title+" form under custom schemes", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			defer test.MockProtect(&setting.Service.ValidSiteURLSchemes)()
			setting.Service.ValidSiteURLSchemes = append(setting.Service.ValidSiteURLSchemes, "h3")

			// Website pattern and note should contain only http(s) and h3
			req := NewRequest(t, "GET", urlStr)
			resp := session.MakeRequest(t, req, http.StatusOK)
			doc := NewHTMLParser(t, resp.Body)
			doc.AssertAttrEqual(t, "input[type=url][name=website]", "pattern", `(http|https|h3)://.+`)
			doc.AssertElementPredicate(t, "input[type=url][name=website] + .help", func(element *goquery.Selection) {
				assert.Equal(t, "Allowed URL schemes include: http, https, h3", strings.TrimSpace(element.Text()))
			})
		})
	}
}
