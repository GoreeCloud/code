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

func TestActionsAPISearchActionJobs_OrgRunner(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteOrganization)

	req := NewRequest(t, "GET",
		fmt.Sprintf("/api/v1/orgs/org3/actions/runners/jobs?labels=%s", "fedora")).
		AddTokenAuth(token)
	res := MakeRequest(t, req, http.StatusOK)

	var jobs []*api.ActionRunJob
	DecodeJSON(t, res, &jobs)

	job395 := api.ActionRunJob{
		ID:      395,
		Attempt: 1,
		Handle:  "40317a2f-2f00-4a82-8cc4-57347989a493",
		RepoID:  1,
		OwnerID: 3,
		Name:    "job_2",
		Needs:   nil,
		RunsOn:  []string{"fedora"},
		TaskID:  47,
		Status:  "waiting",
	}

	assert.ElementsMatch(t, []*api.ActionRunJob{&job395}, jobs)
}

func TestActionsAPISearchActionJobs_OrgRunnerAllPendingJobsWithoutLabels(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteOrganization)

	job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 397})

	req := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners/jobs?labels=").
		AddTokenAuth(token)
	res := MakeRequest(t, req, http.StatusOK)

	var jobs []*api.ActionRunJob
	DecodeJSON(t, res, &jobs)

	assert.Len(t, jobs, 1)
	assert.Equal(t, job.ID, jobs[0].ID)
}

func TestActionsAPISearchActionJobs_OrgRunnerAllPendingJobs(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user1")
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteOrganization)

	job395 := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 395})
	job397 := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 397})

	req := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners/jobs").
		AddTokenAuth(token)
	res := MakeRequest(t, req, http.StatusOK)

	var jobs []*api.ActionRunJob
	DecodeJSON(t, res, &jobs)

	assert.Len(t, jobs, 2)
	assert.Equal(t, job397.ID, jobs[0].ID)
	assert.Equal(t, job395.ID, jobs[1].ID)
}

func TestAPIOrgActionsRunnerRegistrationTokenOperations(t *testing.T) {
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestAPIOrgActionsRunnerRegistrationTokenOperations")()
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	session := loginUser(t, user2.Name)
	readToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadOrganization)

	t.Run("GetRegistrationToken", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners/registration-token")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var registrationToken shared.RegistrationToken
		DecodeJSON(t, response, &registrationToken)

		expected := shared.RegistrationToken{Token: "Sk9wHjBHelH4n1ckQy-mo3KVYRdoaPZ_aaH1ATfgI05"}

		assert.Equal(t, expected, registrationToken)
	})
}

func TestAPIOrgActionsRunnerOperations(t *testing.T) {
	defer unittest.OverrideFixtures("tests/integration/fixtures/TestAPIOrgActionsRunnerOperations")()
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	session := loginUser(t, user2.Name)
	readToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadOrganization)
	writeToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteOrganization)

	runnerOne := &api.ActionRunner{
		ID:          655691,
		UUID:        "a3297f3a-ba5c-4a0f-878e-6cc8b8ac79ec",
		Name:        "runner-1-organization",
		Version:     "dev",
		OwnerID:     3,
		RepoID:      0,
		Description: "A superb runner",
		Labels:      []string{"debian", "gpu"},
		Status:      "offline",
	}
	runnerTwo := &api.ActionRunner{
		ID:          655692,
		UUID:        "6d2d13ef-b19f-47a8-85ad-e82e51f606c5",
		Name:        "runner-2-user",
		Version:     "11.3.1",
		OwnerID:     1,
		RepoID:      0,
		Description: "A splendid runner",
		Labels:      []string{"docker"},
		Status:      "offline",
	}
	runnerThree := &api.ActionRunner{
		ID:          655693,
		UUID:        "0a7e5e05-2da4-44d5-a72a-615da120cef6",
		Name:        "runner-3-organization",
		Version:     "11.3.1",
		OwnerID:     3,
		RepoID:      0,
		Description: "Another fine runner",
		Labels:      []string{"fedora"},
		Status:      "offline",
	}
	runnerFour := &api.ActionRunner{
		ID:          655694,
		UUID:        "166c596c-5016-488d-bd55-b84e5a0460ea",
		Name:        "runner-4-global",
		Version:     "11.3.1",
		OwnerID:     0,
		RepoID:      0,
		Description: "",
		Labels:      []string{},
		Status:      "offline",
	}
	runnerFive := &api.ActionRunner{
		ID:          655695,
		UUID:        "0851ed0a-f0af-4a01-9b98-fc9bf9c1d332",
		Name:        "runner-5-ephemeral",
		Version:     "1.0.0",
		OwnerID:     3,
		RepoID:      0,
		Description: "An ephemeral runner",
		Labels:      []string{"ephemeral-label"},
		Status:      "offline",
		Ephemeral:   true,
	}

	t.Run("Get runners", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		assert.Equal(t, "3", response.Header().Get("X-Total-Count"))

		var runners []*api.ActionRunner
		DecodeJSON(t, response, &runners)

		assert.ElementsMatch(t, []*api.ActionRunner{runnerOne, runnerThree, runnerFive}, runners)
	})

	t.Run("Get runners paginated", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners?page=1&limit=1")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var runners []*api.ActionRunner
		DecodeJSON(t, response, &runners)

		assert.NotEmpty(t, response.Header().Get("Link"))
		assert.NotEmpty(t, response.Header().Get("X-Total-Count"))
		assert.Len(t, runners, 1)
	})

	t.Run("Get visible runners", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners?visible=true")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		assert.NotEmpty(t, response.Header().Get("X-Total-Count"))

		var runners []*api.ActionRunner
		DecodeJSON(t, response, &runners)

		// There are more runners in the result that originate from the global fixtures. The test ignores them to limit
		// the impact of unrelated changes.
		assert.Contains(t, runners, runnerOne)
		assert.NotContains(t, runners, runnerTwo)
		assert.Contains(t, runners, runnerThree)
		assert.Contains(t, runners, runnerFour)
		assert.Contains(t, runners, runnerFive)
	})

	t.Run("Get runner", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners/655691")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var runner *api.ActionRunner
		DecodeJSON(t, response, &runner)

		assert.Equal(t, runnerOne, runner)

		// Instance runner is visible to any organization.
		request = NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners/655694")
		request.AddTokenAuth(readToken)
		response = MakeRequest(t, request, http.StatusOK)

		DecodeJSON(t, response, &runner)

		assert.Equal(t, runnerFour, runner)

		// Runner owned by a user is invisible.
		request = NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners/655692")
		request.AddTokenAuth(readToken)
		MakeRequest(t, request, http.StatusNotFound)
	})

	t.Run("Get ephemeral runner", func(t *testing.T) {
		request := NewRequest(t, "GET", "/api/v1/orgs/org3/actions/runners/655695")
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusOK)

		var runner *api.ActionRunner
		DecodeJSON(t, response, &runner)

		assert.Equal(t, runnerFive, runner)
	})

	t.Run("Delete runner", func(t *testing.T) {
		url := "/api/v1/orgs/org3/actions/runners/655691"

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

		request := NewRequestWithJSON(t, "POST", "/api/v1/orgs/org3/actions/runners", options)
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
		assert.Equal(t, int64(3), registeredRunner.OwnerID)
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

		request := NewRequestWithJSON(t, "POST", "/api/v1/orgs/org3/actions/runners", options)
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

		request := NewRequestWithJSON(t, "POST", "/api/v1/orgs/org3/actions/runners", options)
		request.AddTokenAuth(writeToken)
		response := MakeRequest(t, request, http.StatusCreated)

		var registerRunnerResponse *api.RegisterRunnerResponse
		DecodeJSON(t, response, &registerRunnerResponse)

		secondRequest := NewRequestWithJSON(t, "POST", "/api/v1/orgs/org3/actions/runners", options)
		secondRequest.AddTokenAuth(writeToken)
		secondResponse := MakeRequest(t, secondRequest, http.StatusCreated)

		var secondRegisterRunnerResponse *api.RegisterRunnerResponse
		DecodeJSON(t, secondResponse, &secondRegisterRunnerResponse)

		firstRunner := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunner{UUID: registerRunnerResponse.UUID})
		secondRunner := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunner{UUID: secondRegisterRunnerResponse.UUID})

		assert.NotEqual(t, firstRunner.ID, secondRunner.ID)
		assert.NotEqual(t, firstRunner.UUID, secondRunner.UUID)
	})

	t.Run("Runner registration requires write token for organization scope", func(t *testing.T) {
		options := api.RegisterRunnerOptions{Name: "api-runner"}

		request := NewRequestWithJSON(t, "POST", "/api/v1/orgs/org3/actions/runners", options)
		request.AddTokenAuth(readToken)
		response := MakeRequest(t, request, http.StatusForbidden)

		type errorResponse struct {
			Message string `json:"message"`
		}

		var errorMessage *errorResponse
		DecodeJSON(t, response, &errorMessage)

		assert.Equal(t, "token does not have at least one of required scope(s): [write:organization]", errorMessage.Message)
	})
}
