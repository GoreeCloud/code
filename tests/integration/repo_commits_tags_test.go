// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"strings"
	"testing"

	"forgejo.org/tests"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

// TestRepoCommitsWithTags tests that tags are displayed inline with commit messages
// in the commits list, and not in a separate column
func TestRepoCommitsWithTags(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/user2/repo1/commits/branch/master")
	resp := session.MakeRequest(t, req, http.StatusOK)

	doc := NewHTMLParser(t, resp.Body)

	// Find the commit with SHA 65f1bf27bc3bf70f64657658635e66094edbcb4d
	// This commit should have tags v1.1
	commitRow := doc.doc.Find(`#commits-table tbody tr`).FilterFunction(func(i int, s *goquery.Selection) bool {
		shaLink := s.Find("td.sha a")
		href, _ := shaLink.Attr("href")
		return strings.Contains(href, "65f1bf27bc3bf70f64657658635e66094edbcb4d")
	})

	// 1. Check for tag labels within the message cell
	messageCell := commitRow.Find("td.message")
	tagLabels := messageCell.Find("a.ui.label.basic")
	assert.GreaterOrEqual(t, tagLabels.Length(), 1, "Should find tag label")

	// 2. tag has proper HTML attr and links to the correct tag
	tagFound := false
	tagLabels.Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "v1.1") {
			tagFound = true
			href, exists := s.Attr("href")
			assert.True(t, exists, "Tag should have href")
			assert.Contains(t, href, "/src/tag/v1.1", "Tag link should point to tag page")
			assert.Equal(t, 1, s.Find("svg.octicon-tag").Length(), "Tag should have octicon-tag icon")
		}
	})
	assert.True(t, tagFound, "Should find v1.1 tag")

	// 3. tags appear after the commit message and status indicators
	messageHTML, _ := messageCell.Html()
	messageWrapperPos := strings.Index(messageHTML, "message-wrapper")
	ellipsisButtonPos := strings.Index(messageHTML, "ellipsis-button")
	commitStatusPos := strings.Index(messageHTML, "commit-status")
	tagLabelPos := strings.Index(messageHTML, "ui label basic")

	// 4. Tags should appear after the message wrapper
	assert.Greater(t, tagLabelPos, messageWrapperPos, "Tags should appear after message wrapper")

	// 5. If ellipsis button exists, tags should appear after that one
	if ellipsisButtonPos > 0 {
		assert.Greater(t, tagLabelPos, ellipsisButtonPos, "Tags should appear after ellipsis button")
	}

	// 6. If commit status exists, tags should appear after that one
	if commitStatusPos > 0 {
		assert.Greater(t, tagLabelPos, commitStatusPos, "Tags should appear after commit status")
	}
}
