// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"forgejo.org/models/db"
	issues_model "forgejo.org/models/issues"
	org_model "forgejo.org/models/organization"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/tests"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testIssueCommentChangeEvent(t *testing.T, htmlDoc *HTMLDoc, commentID, badgeOcticon, avatarTitle, avatarLink string, texts, links []string) {
	// Check badge octicon
	badge := htmlDoc.Find("#issuecomment-" + commentID + " .badge svg." + badgeOcticon)
	assert.Equal(t, 1, badge.Length())

	// Check avatar title
	avatarImg := htmlDoc.Find("#issuecomment-" + commentID + " img.avatar")
	if len(avatarTitle) == 0 {
		assert.Zero(t, avatarImg.Length())
	} else {
		assert.Equal(t, 1, avatarImg.Length())
		title, exists := avatarImg.Attr("title")
		assert.True(t, exists)
		assert.Equal(t, avatarTitle, title)
	}

	// Check avatar link
	avatarA := htmlDoc.Find("#issuecomment-" + commentID + " a.avatar")
	if len(avatarLink) == 0 {
		assert.Zero(t, avatarA.Length())
	} else {
		assert.Equal(t, 1, avatarA.Length())
		href, exists := avatarA.Attr("href")
		assert.True(t, exists)
		assert.Equal(t, avatarLink, href)
	}

	event := htmlDoc.Find("#issuecomment-" + commentID + " .text")

	// Check text content
	for _, text := range texts {
		assert.Contains(t, strings.Join(strings.Fields(event.Text()), " "), text)
	}

	var ids []string
	var hrefs []string
	event.Find("a").Each(func(i int, s *goquery.Selection) {
		if id, exists := s.Attr("id"); exists {
			ids = append(ids, id)
		}
		if href, exists := s.Attr("href"); exists {
			hrefs = append(hrefs, href)
		}
	})

	// Check anchors (id)
	assert.Equal(t, []string{"event-" + commentID}, ids)

	// Check links (href)
	issueCommentLink := "#issuecomment-" + commentID
	found := slices.Contains(links, issueCommentLink)
	if !found {
		links = append(links, issueCommentLink)
	}
	assert.Equal(t, links, hrefs)
}

func TestIssueCommentChangeMilestone(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Add milestone
	testIssueCommentChangeEvent(t, htmlDoc, "2000",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 added this to the milestone1 milestone"},
		[]string{"/user1", "/user2/repo1/milestone/1"})

	// Modify milestone
	testIssueCommentChangeEvent(t, htmlDoc, "2001",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 modified the milestone from milestone1 to milestone2"},
		[]string{"/user1", "/user2/repo1/milestone/1", "/user2/repo1/milestone/2"})

	// Remove milestone
	testIssueCommentChangeEvent(t, htmlDoc, "2002",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 removed this from the milestone2 milestone"},
		[]string{"/user1", "/user2/repo1/milestone/2"})

	// Added milestone that in the meantime was deleted
	testIssueCommentChangeEvent(t, htmlDoc, "2003",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 added this to the (deleted) milestone"},
		[]string{"/user1"})

	// Modified milestone - from a meantime deleted one to a valid one
	testIssueCommentChangeEvent(t, htmlDoc, "2004",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 modified the milestone from (deleted) to milestone1"},
		[]string{"/user1", "/user2/repo1/milestone/1"})

	// Modified milestone - from a valid one to a meantime deleted one
	testIssueCommentChangeEvent(t, htmlDoc, "2005",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 modified the milestone from milestone1 to (deleted)"},
		[]string{"/user1", "/user2/repo1/milestone/1"})

	// Modified milestone - from a meantime deleted one to a meantime deleted one
	testIssueCommentChangeEvent(t, htmlDoc, "2006",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 modified the milestone from (deleted) to (deleted)"},
		[]string{"/user1"})

	// Removed milestone that in the meantime was deleted
	testIssueCommentChangeEvent(t, htmlDoc, "2007",
		"octicon-milestone", "User One", "/user1",
		[]string{"user1 removed this from the (deleted) milestone"},
		[]string{"/user1"})
}

func TestIssueCommentChangeProject(t *testing.T) {
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestIssueCommentChangeProject")()
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Add project
	testIssueCommentChangeEvent(t, htmlDoc, "2010",
		"octicon-project", "User One", "/user1",
		[]string{"user1 added this to the First project project"},
		[]string{"/user1", "/user2/repo1/projects/1"})

	// Modify project
	testIssueCommentChangeEvent(t, htmlDoc, "2011",
		"octicon-project", "User One", "/user1",
		[]string{"user1 modified the project from First project to second project"},
		[]string{"/user1", "/user2/repo1/projects/1", "/org3/repo3/projects/2"})

	// Remove project
	testIssueCommentChangeEvent(t, htmlDoc, "2012",
		"octicon-project", "User One", "/user1",
		[]string{"user1 removed this from the second project project"},
		[]string{"/user1", "/org3/repo3/projects/2"})

	// Deleted project
	testIssueCommentChangeEvent(t, htmlDoc, "2013",
		"octicon-project", "User One", "/user1",
		[]string{"user1 added this to the (deleted) project"},
		[]string{"/user1"})

	// Add to user project.
	testIssueCommentChangeEvent(t, htmlDoc, "10001",
		"octicon-project", "User One", "/user1",
		[]string{"user1 added this to the project on user2 project"},
		[]string{"/user1", "/user2/-/projects/4"})

	// Change from user project to repo project.
	testIssueCommentChangeEvent(t, htmlDoc, "10002",
		"octicon-project", "User One", "/user1",
		[]string{"user1 modified the project from project on user2 to second project"},
		[]string{"/user1", "/user2/-/projects/4", "/org3/repo3/projects/2"})

	// Change from repo project to user project.
	testIssueCommentChangeEvent(t, htmlDoc, "10003",
		"octicon-project", "User One", "/user1",
		[]string{"user1 modified the project from second project to project on user2"},
		[]string{"/user1", "/org3/repo3/projects/2", "/user2/-/projects/4"})

	// Remove repo project.
	testIssueCommentChangeEvent(t, htmlDoc, "10004",
		"octicon-project", "User One", "/user1",
		[]string{"user1 removed this from the project on user2 project"},
		[]string{"/user1", "/user2/-/projects/4"})
}

func TestIssueCommentChangeLabel(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Add multiple labels
	testIssueCommentChangeEvent(t, htmlDoc, "2020",
		"octicon-tag", "User One", "/user1",
		[]string{"user1 added the label1 label2 labels "},
		[]string{"/user1", "/user2/repo1/issues?labels=1", "/user2/repo1/issues?labels=2"})
	assert.Empty(t, htmlDoc.Find("#issuecomment-2021 .text").Text())

	// Remove single label
	testIssueCommentChangeEvent(t, htmlDoc, "2022",
		"octicon-tag", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 removed the label1 label "},
		[]string{"/user2", "/user2/repo1/issues?labels=1"})

	// Modify labels (add and remove)
	testIssueCommentChangeEvent(t, htmlDoc, "2023",
		"octicon-tag", "User One", "/user1",
		[]string{"user1 added label1 and removed label2 labels "},
		[]string{"/user1", "/user2/repo1/issues?labels=1", "/user2/repo1/issues?labels=2"})
	assert.Empty(t, htmlDoc.Find("#issuecomment-2024 .text").Text())

	// Add single label
	testIssueCommentChangeEvent(t, htmlDoc, "2025",
		"octicon-tag", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 added the label2 label "},
		[]string{"/user2", "/user2/repo1/issues?labels=2"})

	// Remove multiple labels
	testIssueCommentChangeEvent(t, htmlDoc, "2026",
		"octicon-tag", "User One", "/user1",
		[]string{"user1 removed the label1 label2 labels "},
		[]string{"/user1", "/user2/repo1/issues?labels=1", "/user2/repo1/issues?labels=2"})
	assert.Empty(t, htmlDoc.Find("#issuecomment-2027 .text").Text())
}

func TestIssueCommentChangeAssignee(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Self-assign
	testIssueCommentChangeEvent(t, htmlDoc, "2040",
		"octicon-person", "User One", "/user1",
		[]string{"user1 self-assigned this"},
		[]string{"/user1"})

	// Remove other
	testIssueCommentChangeEvent(t, htmlDoc, "2041",
		"octicon-person", "User One", "/user1",
		[]string{"user1 was unassigned by user2"},
		[]string{"/user1", "/user2"})

	// Add other
	testIssueCommentChangeEvent(t, htmlDoc, "2042",
		"octicon-person", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 was assigned by user1"},
		[]string{"/user2", "/user1"})

	// Self-remove
	testIssueCommentChangeEvent(t, htmlDoc, "2043",
		"octicon-person", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 removed their assignment"},
		[]string{"/user2"})
}

func TestIssueCommentChangeReviewRequest(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	pull := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 6})
	require.NoError(t, pull.LoadIssue(db.DefaultContext))
	issue := pull.Issue
	require.NoError(t, issue.LoadRepo(db.DefaultContext))

	user1, err := user_model.GetUserByID(db.DefaultContext, 1)
	require.NoError(t, err)
	user2, err := user_model.GetUserByID(db.DefaultContext, 2)
	require.NoError(t, err)
	team1, err := org_model.GetTeamByID(db.DefaultContext, 2)
	require.NoError(t, err)
	assert.NotNil(t, team1)

	// Request from other
	comment1, err := issues_model.AddReviewRequest(db.DefaultContext, issue, user2, user1)
	require.NoError(t, err)

	// Refuse review
	comment2, err := issues_model.RemoveReviewRequest(db.DefaultContext, issue, user2, user2)
	require.NoError(t, err)

	// Request from other
	comment3, err := issues_model.AddReviewRequest(db.DefaultContext, issue, user2, user1)
	require.NoError(t, err)
	// Request from team
	comment4, err := issues_model.AddTeamReviewRequest(db.DefaultContext, issue, team1, user1)
	require.NoError(t, err)

	// Remove request from team
	comment5, err := issues_model.RemoveTeamReviewRequest(db.DefaultContext, issue, team1, user2)
	require.NoError(t, err)
	// Request from other
	comment6, err := issues_model.AddReviewRequest(db.DefaultContext, issue, user1, user2)
	require.NoError(t, err)

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/org3/repo3/pulls/2")
	resp := session.MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Request from other
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment1.ID, 10),
		"octicon-eye", "User One", "/user1",
		[]string{"user1 requested review from user2"},
		[]string{"/user1", "/user2"})

	// Refuse review
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment2.ID, 10),
		"octicon-eye", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 refused to review"},
		[]string{"/user2"})

	// Request review from other and from team
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment3.ID, 10),
		"octicon-eye", "User One", "/user1",
		[]string{"user1 requested reviews from user2, team1"},
		[]string{"/user1", "/user2", "/org/org3/teams/team1"})
	assert.Empty(t, htmlDoc.Find("#issuecomment-"+strconv.FormatInt(comment4.ID, 10)+" .text").Text())

	// Remove and add request
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment5.ID, 10),
		"octicon-eye", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 requested reviews from user1 and removed review requests for team1"},
		[]string{"/user2", "/user1", "/org/org3/teams/team1"})
	assert.Empty(t, htmlDoc.Find("#issuecomment-"+strconv.FormatInt(comment6.ID, 10)+" .text").Text())
}

func TestIssueCommentChangeReviewRequestAggregated(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	pull := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 6})
	require.NoError(t, pull.LoadIssue(db.DefaultContext))
	issue := pull.Issue
	require.NoError(t, issue.LoadRepo(db.DefaultContext))

	user1, err := user_model.GetUserByID(db.DefaultContext, 1)
	require.NoError(t, err)
	user2, err := user_model.GetUserByID(db.DefaultContext, 2)
	require.NoError(t, err)
	team1, err := org_model.GetTeamByID(db.DefaultContext, 2)
	require.NoError(t, err)
	assert.NotNil(t, team1)
	label, err := issues_model.GetLabelByID(db.DefaultContext, 1)
	require.NoError(t, err)
	assert.NotNil(t, label)

	// Request from other
	comment1, err := issues_model.AddReviewRequest(db.DefaultContext, issue, user2, user1)
	require.NoError(t, err)
	// Add label
	issues_model.CreateComment(db.DefaultContext, &issues_model.CreateCommentOptions{
		Type:  issues_model.CommentTypeLabel,
		Doer:  user1,
		Issue: issue,
		Label: label,
		Repo:  issue.Repo,
	})

	// Refuse review
	comment2, err := issues_model.RemoveReviewRequest(db.DefaultContext, issue, user2, user2)
	require.NoError(t, err)
	// Remove label
	issues_model.CreateComment(db.DefaultContext, &issues_model.CreateCommentOptions{
		Type:  issues_model.CommentTypeLabel,
		Doer:  user2,
		Issue: issue,
		Label: label,
		Repo:  issue.Repo,
	})

	// Request from other
	comment3, err := issues_model.AddReviewRequest(db.DefaultContext, issue, user2, user1)
	require.NoError(t, err)
	// Request from team
	comment4, err := issues_model.AddTeamReviewRequest(db.DefaultContext, issue, team1, user1)
	require.NoError(t, err)
	// Add label
	issues_model.CreateComment(db.DefaultContext, &issues_model.CreateCommentOptions{
		Type:  issues_model.CommentTypeLabel,
		Doer:  user1,
		Issue: issue,
		Label: label,
		Repo:  issue.Repo,
	})

	// Remove request from team
	comment5, err := issues_model.RemoveTeamReviewRequest(db.DefaultContext, issue, team1, user2)
	require.NoError(t, err)
	// Request from other
	comment6, err := issues_model.AddReviewRequest(db.DefaultContext, issue, user1, user2)
	require.NoError(t, err)
	// Remove label
	issues_model.CreateComment(db.DefaultContext, &issues_model.CreateCommentOptions{
		Type:  issues_model.CommentTypeLabel,
		Doer:  user2,
		Issue: issue,
		Label: label,
		Repo:  issue.Repo,
	})

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/org3/repo3/pulls/2")
	resp := session.MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Request from other
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment1.ID, 10),
		"octicon-eye", "User One", "/user1",
		[]string{"user1 requested review from user2"},
		[]string{"/user1", "#issuecomment-" + strconv.FormatInt(comment1.ID, 10), "/org3/repo3/pulls?labels=1", "/user2"})

	// Refuse review
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment2.ID, 10),
		"octicon-eye", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 refused to review"},
		[]string{"/user2", "#issuecomment-" + strconv.FormatInt(comment2.ID, 10), "/org3/repo3/pulls?labels=1"})

	// Request review from other and from team
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment3.ID, 10),
		"octicon-eye", "User One", "/user1",
		[]string{"user1 requested reviews from user2, team1"},
		[]string{"/user1", "#issuecomment-" + strconv.FormatInt(comment3.ID, 10), "/org3/repo3/pulls?labels=1", "/user2", "/org/org3/teams/team1"})
	assert.Empty(t, htmlDoc.Find("#issuecomment-"+strconv.FormatInt(comment4.ID, 10)+" .text").Text())

	// Remove and add request
	testIssueCommentChangeEvent(t, htmlDoc, strconv.FormatInt(comment5.ID, 10),
		"octicon-eye", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 requested reviews from user1 and removed review requests for team1"},
		[]string{"/user2", "#issuecomment-" + strconv.FormatInt(comment5.ID, 10), "/org3/repo3/pulls?labels=1", "/user1", "/org/org3/teams/team1"})
	assert.Empty(t, htmlDoc.Find("#issuecomment-"+strconv.FormatInt(comment6.ID, 10)+" .text").Text())
}

func TestIssueCommentChangeLock(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Lock without reason
	testIssueCommentChangeEvent(t, htmlDoc, "2050",
		"octicon-lock", "User One", "/user1",
		[]string{"user1 locked and limited conversation to collaborators"},
		[]string{"/user1"})

	// Unlock
	testIssueCommentChangeEvent(t, htmlDoc, "2051",
		"octicon-key", "User One", "/user1",
		[]string{"user1 unlocked this conversation"},
		[]string{"/user1"})

	// Lock with reason
	testIssueCommentChangeEvent(t, htmlDoc, "2052",
		"octicon-lock", "User One", "/user1",
		[]string{"user1 locked as Too heated and limited conversation to collaborators"},
		[]string{"/user1"})

	// Unlock
	testIssueCommentChangeEvent(t, htmlDoc, "2053",
		"octicon-key", "User One", "/user1",
		[]string{"user1 unlocked this conversation"},
		[]string{"/user1"})
}

func TestIssueCommentChangePin(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Pin
	testIssueCommentChangeEvent(t, htmlDoc, "2060",
		"octicon-pin", "User One", "/user1",
		[]string{"user1 pinned this"},
		[]string{"/user1"})

	// Unpin
	testIssueCommentChangeEvent(t, htmlDoc, "2061",
		"octicon-pin", "User One", "/user1",
		[]string{"user1 unpinned this"},
		[]string{"/user1"})
}

func TestIssueCommentChangeOpen(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Close issue
	testIssueCommentChangeEvent(t, htmlDoc, "2070",
		"octicon-circle-slash", "User One", "/user1",
		[]string{"user1 closed this issue"},
		[]string{"/user1"})

	// Reopen issue
	testIssueCommentChangeEvent(t, htmlDoc, "2071",
		"octicon-dot-fill", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 reopened this issue"},
		[]string{"/user2"})

	req = NewRequest(t, "GET", "/user2/repo1/pulls/2")
	resp = MakeRequest(t, req, http.StatusOK)
	htmlDoc = NewHTMLParser(t, resp.Body)

	// Close pull request
	testIssueCommentChangeEvent(t, htmlDoc, "2072",
		"octicon-circle-slash", "User One", "/user1",
		[]string{"user1 closed this pull request"},
		[]string{"/user1"})

	// Reopen pull request
	testIssueCommentChangeEvent(t, htmlDoc, "2073",
		"octicon-dot-fill", "< U<se>r Tw<o > ><", "/user2",
		[]string{"user2 reopened this pull request"},
		[]string{"/user2"})
}

func TestIssueCommentChangeIssueReference(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/issues/1")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Issue reference from issue
	testIssueCommentChangeEvent(t, htmlDoc, "2080",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this issue ", "issue5 #4"},
		[]string{"/user1", "/user2/repo1/issues/4", "#issuecomment-2080", "/user2/repo1/issues/4"})

	// Issue reference from pull
	testIssueCommentChangeEvent(t, htmlDoc, "2081",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this issue ", "issue2 #2"},
		[]string{"/user1", "/user2/repo1/pulls/2", "#issuecomment-2081", "/user2/repo1/pulls/2"})

	// Issue reference from issue in different repo
	testIssueCommentChangeEvent(t, htmlDoc, "2082",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this issue from org3/repo21", "just a normal issue #1"},
		[]string{"/user1", "/org3/repo21/issues/1", "#issuecomment-2082", "/org3/repo21/issues/1"})

	// Issue reference from pull in different repo
	testIssueCommentChangeEvent(t, htmlDoc, "2083",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this issue from user12/repo10 ", "pr2 #1"},
		[]string{"/user1", "/user12/repo10/pulls/1", "#issuecomment-2083", "/user12/repo10/pulls/1"})
}

func TestIssueCommentChangePullReference(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/user2/repo1/pulls/2")
	resp := MakeRequest(t, req, http.StatusOK)
	htmlDoc := NewHTMLParser(t, resp.Body)

	// Pull reference from issue
	testIssueCommentChangeEvent(t, htmlDoc, "2090",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this pull request ", "issue1 #1"},
		[]string{"/user1", "/user2/repo1/issues/1", "#issuecomment-2090", "/user2/repo1/issues/1"})

	// Pull reference from pull
	testIssueCommentChangeEvent(t, htmlDoc, "2091",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this pull request ", "issue2 #2"},
		[]string{"/user1", "/user2/repo1/pulls/2", "#issuecomment-2091", "/user2/repo1/pulls/2"})

	// Pull reference from issue in different repo
	testIssueCommentChangeEvent(t, htmlDoc, "2092",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this pull request from org3/repo21", "just a normal issue #1"},
		[]string{"/user1", "/org3/repo21/issues/1", "#issuecomment-2092", "/org3/repo21/issues/1"})

	// Pull reference from pull in different repo
	testIssueCommentChangeEvent(t, htmlDoc, "2093",
		"octicon-bookmark", "User One", "/user1",
		[]string{"user1 referenced this pull request from user12/repo10 ", "pr2 #1"},
		[]string{"/user1", "/user12/repo10/pulls/1", "#issuecomment-2093", "/user12/repo10/pulls/1"})
}
