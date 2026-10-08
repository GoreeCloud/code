// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	issues_model "forgejo.org/models/issues"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIGetTrackedTimes(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	issue2 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 2})
	require.NoError(t, issue2.LoadRepo(db.DefaultContext))

	session := loginUser(t, user2.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadIssue)

	req := NewRequestf(t, "GET", "/api/v1/repos/%s/%s/issues/%d/times", user2.Name, issue2.Repo.Name, issue2.Index).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)
	var apiTimes api.TrackedTimeList
	DecodeJSON(t, resp, &apiTimes)
	expect, err := issues_model.GetTrackedTimes(db.DefaultContext, &issues_model.FindTrackedTimesOptions{IssueID: issue2.ID})
	require.NoError(t, err)
	assert.Len(t, apiTimes, 3)

	for i, time := range expect {
		assert.Equal(t, time.ID, apiTimes[i].ID)
		assert.Equal(t, issue2.Title, apiTimes[i].Issue.Title)
		assert.Equal(t, issue2.ID, apiTimes[i].IssueID)
		assert.Equal(t, time.Created.Unix(), apiTimes[i].Created.Unix())
		assert.Equal(t, time.Time, apiTimes[i].Time)
		user, err := user_model.GetUserByID(db.DefaultContext, time.UserID)
		require.NoError(t, err)
		assert.Equal(t, user.Name, apiTimes[i].UserName)
	}

	// test filter
	since := "2000-01-01T00%3A00%3A02%2B00%3A00"  // 946684802
	before := "2000-01-01T00%3A00%3A12%2B00%3A00" // 946684812

	req = NewRequestf(t, "GET", "/api/v1/repos/%s/%s/issues/%d/times?since=%s&before=%s", user2.Name, issue2.Repo.Name, issue2.Index, since, before).
		AddTokenAuth(token)
	resp = MakeRequest(t, req, http.StatusOK)
	var filterAPITimes api.TrackedTimeList
	DecodeJSON(t, resp, &filterAPITimes)
	assert.Len(t, filterAPITimes, 2)
	assert.EqualValues(t, 3, filterAPITimes[0].ID)
	assert.EqualValues(t, 6, filterAPITimes[1].ID)

	// test pagination
	allIDs := []int64{}
	for _, page := range []int{1, 2, 3} {
		req = NewRequestf(t, "GET", "/api/v1/repos/%s/%s/issues/%d/times?page=%d&limit=1", user2.Name, issue2.Repo.Name, issue2.Index, page).
			AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)
		var pageAPITimes api.TrackedTimeList
		DecodeJSON(t, resp, &pageAPITimes)
		require.Len(t, pageAPITimes, 1)
		allIDs = append(allIDs, pageAPITimes[0].ID)
	}
	assert.Equal(t, []int64{2, 3, 6}, allIDs)
}

func TestAPIDeleteTrackedTime(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	time6 := unittest.AssertExistsAndLoadBean(t, &issues_model.TrackedTime{ID: 6})
	issue2 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 2})
	require.NoError(t, issue2.LoadRepo(db.DefaultContext))
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	session := loginUser(t, user2.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteIssue)

	// Deletion not allowed
	req := NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/issues/%d/times/%d", user2.Name, issue2.Repo.Name, issue2.Index, time6.ID).
		AddTokenAuth(token)
	MakeRequest(t, req, http.StatusForbidden)

	time3 := unittest.AssertExistsAndLoadBean(t, &issues_model.TrackedTime{ID: 3})
	req = NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/issues/%d/times/%d", user2.Name, issue2.Repo.Name, issue2.Index, time3.ID).
		AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)
	// Delete non existing time
	MakeRequest(t, req, http.StatusNotFound)

	// Reset time of user 2 on issue 2
	trackedSeconds, err := issues_model.GetTrackedSeconds(db.DefaultContext, issues_model.FindTrackedTimesOptions{IssueID: 2, UserID: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(3661), trackedSeconds)

	req = NewRequestf(t, "DELETE", "/api/v1/repos/%s/%s/issues/%d/times", user2.Name, issue2.Repo.Name, issue2.Index).
		AddTokenAuth(token)
	MakeRequest(t, req, http.StatusNoContent)
	MakeRequest(t, req, http.StatusNotFound)

	trackedSeconds, err = issues_model.GetTrackedSeconds(db.DefaultContext, issues_model.FindTrackedTimesOptions{IssueID: 2, UserID: 2})
	require.NoError(t, err)
	assert.Equal(t, int64(0), trackedSeconds)
}

func TestAPIAddTrackedTimes(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	issue2 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 2})
	require.NoError(t, issue2.LoadRepo(db.DefaultContext))
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	admin := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	session := loginUser(t, admin.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteIssue)

	urlStr := fmt.Sprintf("/api/v1/repos/%s/%s/issues/%d/times", user2.Name, issue2.Repo.Name, issue2.Index)

	req := NewRequestWithJSON(t, "POST", urlStr, &api.AddTimeOption{
		Time:    33,
		User:    user2.Name,
		Created: time.Unix(947688818, 0),
	}).AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)
	var apiNewTime api.TrackedTime
	DecodeJSON(t, resp, &apiNewTime)

	assert.EqualValues(t, 33, apiNewTime.Time)
	assert.Equal(t, user2.ID, apiNewTime.UserID)
	assert.EqualValues(t, 947688818, apiNewTime.Created.Unix())
}

// Listing tracked times w/ `/repos/{owner}/{repo}/times/{user}` requires repository admin or site admin permissions (or
// to just list yourself).  This test is a variation of [TestAPIGetTrackedTimes] which uses the `/{user}` endpoint with
// various access token restrictions, validating this API's implementation, but also validating that public-only and
// repo-scoped access tokens don't have admin access.
func TestAPIGetTrackedTimesAuthorizationReducer(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	adminUsername := "user1"
	normalUsername := "user2"
	session := loginUser(t, adminUsername)

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	issue2 := unittest.AssertExistsAndLoadBean(t, &issues_model.Issue{ID: 2})
	require.NoError(t, issue2.LoadRepo(db.DefaultContext))

	test := func(t *testing.T, token string, expectedStatus int) {
		req := NewRequest(t, "GET",
			fmt.Sprintf("/api/v1/repos/%s/%s/times/%s", user2.Name, issue2.Repo.Name, normalUsername)).
			AddTokenAuth(token)
		resp := MakeRequest(t, req, expectedStatus)
		if expectedStatus == http.StatusOK {
			var apiTimes api.TrackedTimeList
			DecodeJSON(t, resp, &apiTimes)
			assert.Len(t, apiTimes, 3)
		}
	}

	t.Run("all access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		allToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)
		test(t, allToken, http.StatusOK)
	})

	t.Run("public-only access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		publicOnlyToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopePublicOnly, auth_model.AccessTokenScopeReadRepository)
		test(t, publicOnlyToken, http.StatusForbidden)
	})

	t.Run("specific repo access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		repo2OnlyToken := createFineGrainedRepoAccessToken(t, adminUsername,
			[]auth_model.AccessTokenScope{auth_model.AccessTokenScopeReadRepository},
			[]int64{issue2.RepoID},
		)
		test(t, repo2OnlyToken, http.StatusForbidden)
	})
}
