// Copyright 2017 The Gitea Authors. All rights reserved.
// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func testExploreStarForkCounters(t *testing.T, repoQuery, expectedStars, expectedForks string) {
	resp := MakeRequest(t, NewRequest(t, "GET", fmt.Sprintf("/explore/repos?search=%s", repoQuery)), http.StatusOK)

	repoListEntry := NewHTMLParser(t, resp.Body).Find(fmt.Sprintf(".flex-list > .flex-item:has(a[href='/%s'])", repoQuery))
	starsAriaLabel, _ := repoListEntry.Find("a[href$='/stars']").Attr("aria-label")
	forksAriaLabel, _ := repoListEntry.Find("a[href$='/forks']").Attr("aria-label")

	assert.Equal(t, expectedStars, starsAriaLabel)
	assert.Equal(t, expectedForks, forksAriaLabel)

	// Verify that correct icons are used
	assert.True(t, repoListEntry.Find("a[href$='/stars'] > svg").HasClass("octicon-star"))
	assert.True(t, repoListEntry.Find("a[href$='/forks'] > svg").HasClass("octicon-repo-forked"))
}

func TestExploreRepos(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/explore/repos")
	MakeRequest(t, req, http.StatusOK)

	t.Run("Counters", func(t *testing.T) {
		if setting.Database.Type.IsPostgreSQL() {
			t.Skip("PGSQL is unable to find user2/repo1")
		}

		repo := "user2/repo1"
		// Initial state: zeroes in the counters
		testExploreStarForkCounters(t, repo, "0 stars", "0 forks")

		// Star the repo
		session := loginUser(t, "user5")
		session.MakeRequest(t, NewRequest(t, "POST", fmt.Sprintf("/%s/action/star", repo)), http.StatusOK)

		// Stars counter should have incremented
		testExploreStarForkCounters(t, repo, "1 star", "0 forks")
	})

	t.Run("Persistent parameters", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		req := NewRequest(t, "GET", "/explore/repos?topic=1&language=Go")
		resp := MakeRequest(t, req, http.StatusOK)
		htmlDoc := NewHTMLParser(t, resp.Body).Find("#repo-search-form")

		assert.Equal(t, "Go", htmlDoc.Find("input[name='language']").AttrOr("value", "not found"))
		assert.Equal(t, "true", htmlDoc.Find("input[name='topic']").AttrOr("value", "not found"))
	})
}
