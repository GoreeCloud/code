// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/test"
	"forgejo.org/modules/timeutil"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SearchResults struct {
	OK   bool        `json:"ok"`
	Data []*api.User `json:"data"`
}

func TestAPIUserSearchLoggedIn(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	adminUsername := "user1"
	session := loginUser(t, adminUsername)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser)
	query := "user2"
	req := NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.NotEmpty(t, results.Data)
	for _, user := range results.Data {
		assert.Contains(t, user.UserName, query)
		assert.NotEmpty(t, user.Email)
	}

	publicToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser, auth_model.AccessTokenScopePublicOnly)
	req = NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query).
		AddTokenAuth(publicToken)
	resp = MakeRequest(t, req, http.StatusOK)
	results = SearchResults{}
	DecodeJSON(t, resp, &results)
	assert.NotEmpty(t, results.Data)
	for _, user := range results.Data {
		assert.Contains(t, user.UserName, query)
		assert.NotEmpty(t, user.Email)
		assert.Equal(t, "public", user.Visibility)
	}
}

func TestAPIUserSearchNotLoggedIn(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	query := "user2"
	req := NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query)
	resp := MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.NotEmpty(t, results.Data)
	var modelUser *user_model.User
	for _, user := range results.Data {
		assert.Contains(t, user.UserName, query)
		modelUser = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: user.ID})
		assert.Equal(t, modelUser.GetPlaceholderEmail(), user.Email)
	}
}

func TestAPIUserSearchPaged(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	defer test.MockVariableValue(&setting.API.DefaultPagingNum, 5)()

	req := NewRequest(t, "GET", "/api/v1/users/search?limit=1")
	resp := MakeRequest(t, req, http.StatusOK)

	var limitedResults SearchResults
	DecodeJSON(t, resp, &limitedResults)
	assert.Len(t, limitedResults.Data, 1)

	req = NewRequest(t, "GET", "/api/v1/users/search")
	resp = MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.Len(t, results.Data, 5)
}

func TestAPIUserSearchSystemUsers(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	for _, systemUser := range []*user_model.User{
		user_model.NewGhostUser(),
		user_model.NewActionsUser(),
	} {
		t.Run(systemUser.Name, func(t *testing.T) {
			req := NewRequestf(t, "GET", "/api/v1/users/search?uid=%d", systemUser.ID)
			resp := MakeRequest(t, req, http.StatusOK)

			var results SearchResults
			DecodeJSON(t, resp, &results)
			assert.NotEmpty(t, results.Data)
			if assert.Len(t, results.Data, 1) {
				user := results.Data[0]
				assert.Equal(t, user.UserName, systemUser.Name)
				assert.Equal(t, user.ID, systemUser.ID)
			}
		})
	}
}

func TestAPIUserSearchAdminLoggedInUserHidden(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	adminUsername := "user1"
	session := loginUser(t, adminUsername)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser)
	query := "user31"
	req := NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.NotEmpty(t, results.Data)
	for _, user := range results.Data {
		assert.Contains(t, user.UserName, query)
		assert.NotEmpty(t, user.Email)
		assert.Equal(t, "private", user.Visibility)
	}
}

func TestAPIUserSearchNotLoggedInUserHidden(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	query := "user31"
	req := NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query)
	resp := MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.Empty(t, results.Data)
}

func TestAPIUserSearchByEmail(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// admin can search user with private email
	adminUsername := "user1"
	session := loginUser(t, adminUsername)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser)
	query := "user2@example.com"
	req := NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query).
		AddTokenAuth(token)
	resp := MakeRequest(t, req, http.StatusOK)

	var results SearchResults
	DecodeJSON(t, resp, &results)
	assert.Len(t, results.Data, 1)
	assert.Equal(t, query, results.Data[0].Email)

	// no login user can not search user with private email
	req = NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query)
	resp = MakeRequest(t, req, http.StatusOK)
	DecodeJSON(t, resp, &results)
	assert.Empty(t, results.Data)

	// user can search self with private email
	user2 := "user2"
	session = loginUser(t, user2)
	token = getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser)
	req = NewRequestf(t, "GET", "/api/v1/users/search?q=%s", query).
		AddTokenAuth(token)
	resp = MakeRequest(t, req, http.StatusOK)

	DecodeJSON(t, resp, &results)
	assert.Len(t, results.Data, 1)
	assert.Equal(t, query, results.Data[0].Email)
}

func TestUsersSearchSorted(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	createTimestamp := time.Now().Unix() - 1000
	updateTimestamp := time.Now().Unix() - 500
	sess := db.GetEngine(context.Background())

	for i := int64(1); i <= 10; i++ {
		name := "sorttest" + strconv.Itoa(int(i))
		user := &user_model.User{
			Name:        name,
			LowerName:   name,
			LoginName:   name,
			Email:       name + "@example.com",
			Passwd:      name + ".password",
			Avatar:      "xyz",
			Type:        user_model.UserTypeIndividual,
			LoginType:   auth_model.OAuth2,
			CreatedUnix: timeutil.TimeStamp(createTimestamp - i),
			UpdatedUnix: timeutil.TimeStamp(updateTimestamp - i),
		}
		_, err := sess.NoAutoTime().Insert(user)
		require.NoError(t, err)
	}

	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser)

	testCases := []struct {
		sortType      string
		expectedUsers []string
	}{
		{"alphabetically", []string{"sorttest1", "sorttest10", "sorttest2", "sorttest3"}},
		{"reversealphabetically", []string{"sorttest9", "sorttest8", "sorttest7", "sorttest6"}},
		{"newest", []string{"sorttest1", "sorttest2", "sorttest3", "sorttest4"}},
		{"oldest", []string{"sorttest10", "sorttest9", "sorttest8", "sorttest7"}},
		{"recentupdate", []string{"sorttest1", "sorttest2", "sorttest3", "sorttest4"}},
		{"leastupdate", []string{"sorttest10", "sorttest9", "sorttest8", "sorttest7"}},
	}

	for _, testCase := range testCases {
		req := NewRequest(
			t,
			"GET",
			fmt.Sprintf("/api/v1/users/search?q=sorttest&sort=%s&limit=4",
				testCase.sortType,
			),
		).AddTokenAuth(token)
		resp := session.MakeRequest(t, req, http.StatusOK)

		var results SearchResults
		DecodeJSON(t, resp, &results)
		assert.Len(t, results.Data, 4)
		for i, searchData := range results.Data {
			assert.Equalf(t, testCase.expectedUsers[i], searchData.UserName, "Sort type: %s, index %d", testCase.sortType, i)
		}
	}
}
