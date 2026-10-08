// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"net/http"
	"strings"
	"testing"

	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

// TestRepoCommitsTemplateVariables ensures that template variables in commits_list.tmpl are correctly referenced
func TestRepoCommitsTemplateVariables(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")

	// Test the main commits page
	req := NewRequest(t, "GET", "/user2/repo1/commits/branch/master")
	resp := session.MakeRequest(t, req, http.StatusOK)
	assert.Equal(t, http.StatusOK, resp.Code, "Template should render without errors")

	doc := NewHTMLParser(t, resp.Body)

	// 1. Repository.Link is used in tag template
	tagLinks := doc.doc.Find("a.ui.label.basic[href*='/src/tag/']")
	if tagLinks.Length() > 0 {
		href, _ := tagLinks.First().Attr("href")
		assert.Contains(t, href, "/user2/repo1/src/tag/", "Repository link should be correctly rendered in tag URLs")
	}

	// 2. Repository.ObjectFormatName is used in the SHA column header
	shaHeader := doc.doc.Find("#commits-table thead tr th.sha")
	assert.Equal(t, 1, shaHeader.Length(), "SHA column header should exist")
	headerText := strings.TrimSpace(shaHeader.Text())
	assert.NotEmpty(t, headerText, "SHA column header should have text (ObjectFormatName)")
	// Should be uppercase SHA1 or SHA256 depending on the repository format
	assert.True(t, headerText == "SHA1" || headerText == "SHA256", "ObjectFormatName should be rendered correctly, got: %s", headerText)

	// 3. Repository.ComposeMetas is used for rendering commit messages
	commitMessages := doc.doc.Find("#commits-table tbody tr td.message .commit-summary")
	assert.Positive(t, commitMessages.Length(), "Should have commit messages rendered")

	// 4. RepoLink variable is used throughout
	commitLinks := doc.doc.Find("#commits-table tbody tr td.sha a[href*='/commit/']")
	assert.Positive(t, commitLinks.Length(), "Should have commit links")
	firstCommitLink, _ := commitLinks.First().Attr("href")
	assert.Contains(t, firstCommitLink, "/user2/repo1/commit/", "RepoLink should be correctly used in commit URLs")

	// 5. CommitTagsMap is used for tag rendering
	// If $.CommitTagsMap is mistyped, the template would fail with a 500 error
	// (for detailed tag rendering tests see repo_commits_tags_test.go)
	tagLabels := doc.doc.Find("#commits-table tbody tr td.message a.ui.label.basic")
	if tagLabels.Length() > 0 {
		assert.NotContains(t, tagLabels.First().Text(), "{{", "Tags should be properly rendered without template syntax")
	}
}
