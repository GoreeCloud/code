// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	actions_model "forgejo.org/models/actions"
	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	api "forgejo.org/modules/structs"
	"forgejo.org/routers/api/v1/shared"
	"forgejo.org/tests"

	gouuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionsAPISearchActionJobs_UserRunner(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	normalUsername := "user2"
	session := loginUser(t, normalUsername)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteUser)

	req := NewRequest(t, "GET",
		fmt.Sprintf("/api/v1/user/actions/runners/jobs?labels=%s", "debian-latest")).
		AddTokenAuth(token)
	res := MakeRequest(t, req, http.StatusOK)

	var jobs []*api.ActionRunJob
	DecodeJSON(t, res, &jobs)

	job394 := api.ActionRunJob{
		ID:      394,
		Attempt: 2,
		Handle:  "a723d3e3-49a1-4e6b-947f-e987e60bfbd6",
		RepoID:  1,
		OwnerID: 2,
		Name:    "job_2",
		Needs:   nil,
		RunsOn:  []string{"debian-latest"},
		TaskID:  47,
		Status:  "waiting",
	}

	assert.ElementsMatch(t, []*api.ActionRunJob{&job394}, jobs)
}

func TestActionsAPISearchActionJobs_UserRunnerAllPendingJobsWithoutLabels(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	normalUsername := "user1"
	session := loginUser(t, normalUsername)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteUser)
	job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 196})

	req := NewRequest(t, "GET", "/api/v1/user/actions/runners/jobs?labels=").
		AddTokenAuth(token)
	res := MakeRequest(t, req, http.StatusOK)

	var jobs []*api.ActionRunJob
	DecodeJSON(t, res, &jobs)

	assert.Len(t, jobs, 1)
	assert.Equal(t, job.ID, jobs[0].ID)
}

func TestActionsAPISearchActionJobs_UserRunnerAllPendingJobs(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	normalUsername := "user2"
	session := loginUser(t, normalUsername)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteUser)
	job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 394})

	req := NewRequest(t, "GET", "/api/v1/user/actions/runners/jobs").
		AddTokenAuth(token)
	res := MakeRequest(t, req, http.StatusOK)

	var jobs []*api.ActionRunJob
	DecodeJSON(t, res, &jobs)

	assert.Len(t, jobs, 1)
	assert.Equal(t, job.ID, jobs[0].ID)
}

func TestAPIUserActionsRunnerRegistrationTokenOperations(t *testing.T) {
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestAPIUserActionsRunnerRegistrationTokenOperations")()
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	session := loginUser(t, user2.Name)
	readToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser)

	t.Run("GetRegistrationToken", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/user/actions/runners/registration-token")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var registrationToken shared.RegistrationToken
		DecodeJSON(t, response, &registrationToken)

		expected := shared.RegistrationToken{Token: "Xb3WmQBum2S0-WwFY399A0DhnPkgRdXzpEOJaMmL5UT"}

		assert.Equal(t, expected, registrationToken)
	})
}

func TestAPIUserActionsRunnerOperations(t *testing.T) {
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestAPIUserActionsRunnerOperations")()
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	session := loginUser(t, user2.Name)
	readToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadUser)
	writeToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteUser)

	runnerOne := &api.ActionRunner{
		ID:          71301,
		UUID:        "99fc4a58-a25e-4dbe-b6ea-3d55dddcd216",
		Name:        "runner-1-user",
		Version:     "dev",
		OwnerID:     2,
		RepoID:      0,
		Description: "A superb runner",
		Labels:      []string{"debian", "gpu"},
		Status:      "offline",
	}
	runnerThree := &api.ActionRunner{
		ID:          71303,
		UUID:        "70bc0da3-35b2-4129-bbc9-4679dfdda4d0",
		Name:        "runner-3-user",
		Version:     "11.3.1",
		OwnerID:     2,
		RepoID:      0,
		Description: "Another fine runner",
		Labels:      []string{"fedora"},
		Status:      "offline",
	}
	runnerFour := &api.ActionRunner{
		ID:          71304,
		UUID:        "3873c473-47b8-4559-9fa5-843277419780",
		Name:        "runner-4-global",
		Version:     "11.3.1",
		OwnerID:     0,
		RepoID:      0,
		Description: "",
		Labels:      []string{},
		Status:      "offline",
		Ephemeral:   false,
	}
	runnerFive := &api.ActionRunner{
		ID:          71305,
		UUID:        "3ca04a95-3e75-4e48-8b7a-63427ebcf3b8",
		Name:        "runner-5-user-ephemeral",
		Version:     "1.0.0",
		OwnerID:     2,
		RepoID:      0,
		Description: "An ephemeral runner",
		Labels:      []string{"ephemeral-label"},
		Status:      "offline",
		Ephemeral:   true,
	}

	t.Run("Get runners", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/user/actions/runners")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		assert.Equal(t, "3", response.Header().Get("X-Total-Count"))

		var runners []*api.ActionRunner
		DecodeJSON(t, response, &runners)

		assert.ElementsMatch(t, []*api.ActionRunner{runnerOne, runnerThree, runnerFive}, runners)
	})

	t.Run("Get runners paginated", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/user/actions/runners?page=1&limit=1")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var runners []*api.ActionRunner
		DecodeJSON(t, response, &runners)

		assert.NotEmpty(t, response.Header().Get("Link"))
		assert.NotEmpty(t, response.Header().Get("X-Total-Count"))
		assert.Len(t, runners, 1)
	})

	t.Run("Get visible runners", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/user/actions/runners?visible=true")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var runners []*api.ActionRunner
		DecodeJSON(t, response, &runners)

		// There are more runners in the result that originate from the global fixtures. The test ignores them to limit
		// the impact of unrelated changes.
		assert.Contains(t, runners, runnerOne)
		assert.Contains(t, runners, runnerThree)
		assert.Contains(t, runners, runnerFour)
		assert.Contains(t, runners, runnerFive)
	})

	t.Run("Get runner", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/user/actions/runners/71303")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var runner *api.ActionRunner
		DecodeJSON(t, response, &runner)

		assert.Equal(t, runnerThree, runner)

		// Runner owned by instance is visible to user.
		request = NewRequest(t, "GET", "/api/v1/user/actions/runners/71304")
		request.AddTokenAuth(readToken)
		response = MakeRequest(t, request, http.StatusOK)

		DecodeJSON(t, response, &runner)

		assert.Equal(t, runnerFour, runner)

		// Runner owned by different user is invisible.
		request = NewRequest(t, "GET", "/api/v1/user/actions/runners/71302")
		request.AddTokenAuth(readToken)
		MakeRequest(t, request, http.StatusNotFound)
	})

	t.Run("Get ephemeral runner", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/user/actions/runners/71305")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var runner *api.ActionRunner
		DecodeJSON(t, response, &runner)

		assert.Equal(t, runnerFive, runner)
	})

	t.Run("Delete runner", func(t *testing.T) {
		url := "/api/v1/user/actions/runners/71303"

		request := NewRequest(t, "GET", url)
		request.AddTokenAuth(readToken)
		MakeRequest(t, request, http.StatusOK)

		deleteRequest := NewRequest(t, "DELETE", url)
		deleteRequest.AddTokenAuth(writeToken)
		MakeRequest(t, deleteRequest, http.StatusNoContent)

		request = NewRequest(t, "GET", url)
		request.AddTokenAuth(readToken)
		MakeRequest(t, request, http.StatusNotFound)
	})

	t.Run("Register runner", func(t *testing.T) {
		options := api.RegisterRunnerOptions{Name: "api-runner", Description: "Some description"}

		request := NewRequestWithJSON(t, "POST", "/api/v1/user/actions/runners", options)
		request.AddTokenAuth(writeToken)
		response := MakeRequest(t, request, http.StatusCreated)

		var registerRunnerResponse *api.RegisterRunnerResponse
		DecodeJSON(t, response, &registerRunnerResponse)

		assert.NotNil(t, registerRunnerResponse)
		assert.Positive(t, registerRunnerResponse.ID)
		assert.Equal(t, gouuid.Version(4), gouuid.MustParse(registerRunnerResponse.UUID).Version())
		assert.Regexp(t, "(?i)^[0-9a-f]{40}$", registerRunnerResponse.Token)

		registeredRunner := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunner{UUID: registerRunnerResponse.UUID})
		assert.Equal(t, registerRunnerResponse.ID, registeredRunner.ID)
		assert.Equal(t, registerRunnerResponse.UUID, registeredRunner.UUID)
		assert.Equal(t, user2.ID, registeredRunner.OwnerID)
		assert.Zero(t, registeredRunner.RepoID)
		assert.Equal(t, "api-runner", registeredRunner.Name)
		assert.Equal(t, "Some description", registeredRunner.Description)
		assert.Empty(t, registeredRunner.AgentLabels)
		assert.Empty(t, registeredRunner.Version)
		assert.NotEmpty(t, registeredRunner.TokenHash)
		assert.NotEmpty(t, registeredRunner.TokenSalt)
		assert.False(t, registeredRunner.Ephemeral)
	})

	t.Run("Register ephemeral runner", func(t *testing.T) {
		options := api.RegisterRunnerOptions{Name: "ephemeral-runner", Description: "Ephemeral runner", Ephemeral: true}

		request := NewRequestWithJSON(t, "POST", "/api/v1/user/actions/runners", options)
		request.AddTokenAuth(writeToken)
		response := MakeRequest(t, request, http.StatusCreated)

		var registerRunnerResponse *api.RegisterRunnerResponse
		DecodeJSON(t, response, &registerRunnerResponse)

		registeredRunner := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunner{UUID: registerRunnerResponse.UUID})
		assert.Equal(t, registerRunnerResponse.UUID, registeredRunner.UUID)
		assert.True(t, registeredRunner.Ephemeral)
	})

	t.Run("Runner registration does not update runner with identical name", func(t *testing.T) {
		options := api.RegisterRunnerOptions{Name: "api-runner"}

		request := NewRequestWithJSON(t, "POST", "/api/v1/user/actions/runners", options)
		request.AddTokenAuth(writeToken)
		response := MakeRequest(t, request, http.StatusCreated)

		var registerRunnerResponse *api.RegisterRunnerResponse
		DecodeJSON(t, response, &registerRunnerResponse)

		secondRequest := NewRequestWithJSON(t, "POST", "/api/v1/user/actions/runners", options)
		secondRequest.AddTokenAuth(writeToken)
		secondResponse := MakeRequest(t, secondRequest, http.StatusCreated)

		var secondRegisterRunnerResponse *api.RegisterRunnerResponse
		DecodeJSON(t, secondResponse, &secondRegisterRunnerResponse)

		firstRunner := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunner{UUID: registerRunnerResponse.UUID})
		secondRunner := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunner{UUID: secondRegisterRunnerResponse.UUID})

		assert.NotEqual(t, firstRunner.ID, secondRunner.ID)
		assert.NotEqual(t, firstRunner.UUID, secondRunner.UUID)
	})

	t.Run("Runner registration requires write token for user scope", func(t *testing.T) {
		options := api.RegisterRunnerOptions{Name: "api-runner"}

		request := NewRequestWithJSON(t, "POST", "/api/v1/user/actions/runners", options)
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusForbidden)

		type errorResponse struct {
			Message string `json:"message"`
		}

		var errorMessage *errorResponse
		DecodeJSON(t, response, &errorMessage)

		assert.Equal(t, "token does not have at least one of required scope(s): [write:user]", errorMessage.Message)
	})
}
