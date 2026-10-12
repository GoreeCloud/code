// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"fmt"
	"net/http"
	"testing"

	actions_model "forgejo.org/models/actions"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionRunDeletion(t *testing.T) {
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	repo1 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1, OwnerID: user2.ID})
	user5 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})
	repo62 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 62, OwnerID: user2.ID})

	sessUser2 := loginUser(t, user2.Name)
	sessUser5 := loginUser(t, user5.Name)

	t.Run("Run removed", func(t *testing.T) {
		defer unittest.OverrideFixtures("tests/integration/fixtures/TestActionRunDeletion")()
		defer tests.PrepareTestEnv(t)()

		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 34901})
		job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{RunID: run.ID})
		unittest.AssertCount(t, &actions_model.ActionArtifact{RunID: run.ID}, 2)
		runner := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunner{ID: 41601})

		requestURL := fmt.Sprintf("/%s/actions/runs/%d/delete", repo1.FullName(), run.Index)
		request := NewRequest(t, "POST", requestURL)
		response := sessUser2.MakeRequest(t, request, http.StatusOK)
		assert.JSONEq(t, `{"ok":true}`, response.Body.String())

		unittest.AssertNotExistsBean(t, &actions_model.ActionRun{ID: run.ID})
		unittest.AssertNotExistsBean(t, &actions_model.ActionRunJob{ID: job.ID})
		unittest.AssertCount(t, &actions_model.ActionArtifact{
			RunID:  run.ID,
			Status: int64(actions_model.ArtifactStatusPendingDeletion),
		}, 2)
		unittest.AssertNotExistsBean(t, &actions_model.ActionRunner{ID: runner.ID})
	})

	t.Run("Error if run has not completed", func(t *testing.T) {
		defer unittest.OverrideFixtures("tests/integration/fixtures/TestActionRunDeletion")()
		defer tests.PrepareTestEnv(t)()

		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 34902})

		requestURL := fmt.Sprintf("/%s/actions/runs/%d/delete", repo1.FullName(), run.Index)
		request := NewRequest(t, "POST", requestURL)
		response := sessUser2.MakeRequest(t, request, http.StatusInternalServerError)
		assert.JSONEq(t, `{"message":"Could not delete run."}`, response.Body.String())

		// Verify that the run still exists.
		unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: run.ID})
	})

	t.Run("Nothing happens if run does not belong to repository", func(t *testing.T) {
		defer unittest.OverrideFixtures("tests/integration/fixtures/TestActionRunDeletion")()
		defer tests.PrepareTestEnv(t)()

		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 34901})

		requestURL := fmt.Sprintf("/%s/actions/runs/%d/delete", repo62.FullName(), run.Index)
		request := NewRequest(t, "POST", requestURL)
		response := sessUser2.MakeRequest(t, request, http.StatusOK)
		assert.JSONEq(t, `{"ok":true}`, response.Body.String())

		// Verify that the run still exists.
		unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: run.ID})
	})

	t.Run("Nothing happens if run does not exist", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertNotExistsBean(t, &actions_model.ActionRun{Index: 260871})

		requestURL := fmt.Sprintf("/%s/actions/runs/%d/delete", repo1.FullName(), 260871)
		request := NewRequest(t, "POST", requestURL)
		response := sessUser2.MakeRequest(t, request, http.StatusOK)
		assert.JSONEq(t, `{"ok":true}`, response.Body.String())
	})

	t.Run("Removal requires ownership", func(t *testing.T) {
		defer unittest.OverrideFixtures("tests/integration/fixtures/TestActionRunDeletion")()
		defer tests.PrepareTestEnv(t)()

		isCollaborator, err := repo_model.IsCollaborator(t.Context(), repo1.ID, user5.ID)
		require.NoError(t, err)
		require.True(t, isCollaborator)

		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 34901})

		requestURL := fmt.Sprintf("/%s/actions/runs/%d/delete", repo1.FullName(), run.Index)
		request := NewRequest(t, "POST", requestURL)
		MakeRequest(t, request, http.StatusNotFound)

		// Verify that run still exists
		unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: run.ID})

		request = NewRequest(t, "POST", requestURL)
		sessUser5.MakeRequest(t, request, http.StatusNotFound)

		// Verify that run still exists
		unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: run.ID})

		request = NewRequest(t, "POST", requestURL)
		sessUser2.MakeRequest(t, request, http.StatusOK)

		// Verify that run no longer exists
		unittest.AssertNotExistsBean(t, &actions_model.ActionRun{ID: run.ID})
	})
}

func TestActionRunPrioritization(t *testing.T) {
	fixtures := []*actions_model.ActionRun{
		{ID: 535681, Index: 1, RepoID: 62, OwnerID: 2, Status: actions_model.StatusSuccess, Priority: actions_model.DefaultRunPriority},
		{ID: 535682, Index: 2, RepoID: 62, OwnerID: 2, Status: actions_model.StatusWaiting, Priority: actions_model.DefaultRunPriority},
		{ID: 535683, Index: 3, RepoID: 62, OwnerID: 2, Status: actions_model.StatusBlocked, Priority: actions_model.MaxRunPriority},
		{ID: 535684, Index: 1, RepoID: 1, OwnerID: 2, Status: actions_model.StatusBlocked, Priority: actions_model.DefaultRunPriority, Prioritize: true},
		{ID: 535685, Index: 4, RepoID: 62, OwnerID: 2, Status: actions_model.StatusWaiting},
	}

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	user5 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})
	repo62 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 62, OwnerID: user2.ID})

	user2Sess := loginUser(t, user2.Name)
	user5Sess := loginUser(t, user5.Name)

	t.Run("Prioritize run", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/2/prioritize", repo62.FullName()))
		response := user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		assert.Equal(t, "/user2/test_workflows/actions?actor=0&page=0&status=0&workflow=",
			response.Header().Get("Location"))

		runTwo := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535682})
		assert.True(t, runTwo.Prioritize)
		assert.Equal(t, actions_model.MaxRunPriority, runTwo.Priority)

		// Verify that the rest of the queue has been reprioritized.
		runThree := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535683})
		assert.False(t, runThree.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runThree.Priority)

		runFour := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535684})
		assert.True(t, runFour.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runFour.Priority) // No change because different repository.

		runFive := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535685})
		assert.False(t, runFive.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runFive.Priority)
	})

	t.Run("Prioritize completed run", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/1/prioritize", repo62.FullName()))
		user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		runOne := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535681})
		assert.True(t, runOne.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runOne.Priority) // No change because run has completed.
	})

	t.Run("Requires write permissions", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/2/prioritize", repo62.FullName()))
		user5Sess.MakeRequest(t, request, http.StatusNotFound)

		// The run should not have been changed because user5 has not the necessary permissions.
		runTwo := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535682})
		assert.False(t, runTwo.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runTwo.Priority)

		request = NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/2/prioritize", repo62.FullName()))
		user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		runTwo = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535682})
		assert.True(t, runTwo.Prioritize)
		assert.Equal(t, actions_model.MaxRunPriority, runTwo.Priority)
	})

	t.Run("Redirect URL contains supplied parameters", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequestWithValues(t, "POST", fmt.Sprintf("/%s/actions/runs/2/prioritize", repo62.FullName()),
			map[string]string{"actor": "10", "page": "3", "status": "2", "workflow": "test.yaml"})
		response := user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		assert.Equal(t, "/user2/test_workflows/actions?actor=10&page=3&status=2&workflow=test.yaml",
			response.Header().Get("Location"))
	})
}

func TestActionRunDeprioritization(t *testing.T) {
	fixtures := []*actions_model.ActionRun{
		{ID: 535681, Index: 1, RepoID: 62, OwnerID: 2, Status: actions_model.StatusSuccess, Priority: actions_model.MaxRunPriority, Prioritize: true},
		{ID: 535682, Index: 2, RepoID: 62, OwnerID: 2, Status: actions_model.StatusWaiting, Priority: actions_model.MaxRunPriority, Prioritize: true},
		{ID: 535683, Index: 3, RepoID: 62, OwnerID: 2, Status: actions_model.StatusBlocked, Priority: actions_model.MaxRunPriority},
		{ID: 535684, Index: 1, RepoID: 1, OwnerID: 2, Status: actions_model.StatusBlocked, Priority: actions_model.DefaultRunPriority, Prioritize: true},
		{ID: 535685, Index: 4, RepoID: 62, OwnerID: 2, Status: actions_model.StatusWaiting},
	}

	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	user5 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})
	repo62 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 62, OwnerID: user2.ID})

	user2Sess := loginUser(t, user2.Name)
	user5Sess := loginUser(t, user5.Name)

	t.Run("Deprioritize run", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/2/deprioritize", repo62.FullName()))
		response := user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		assert.Equal(t, "/user2/test_workflows/actions?actor=0&page=0&status=0&workflow=",
			response.Header().Get("Location"))

		runTwo := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535682})
		assert.False(t, runTwo.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runTwo.Priority)

		// Verify that the rest of the queue has been reprioritized.
		runThree := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535683})
		assert.False(t, runThree.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runThree.Priority)

		runFour := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535684})
		assert.True(t, runFour.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runFour.Priority) // No change because different repository.

		runFive := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535685})
		assert.False(t, runFive.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runFive.Priority)
	})

	t.Run("Deprioritize completed run", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/1/deprioritize", repo62.FullName()))
		user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		runOne := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535681})
		assert.False(t, runOne.Prioritize)
		assert.Equal(t, actions_model.MaxRunPriority, runOne.Priority) // No change because run has completed.
	})

	t.Run("Requires write permissions", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/2/deprioritize", repo62.FullName()))
		user5Sess.MakeRequest(t, request, http.StatusNotFound)

		// The run should not have been changed because user5 has not the necessary permissions.
		runTwo := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535682})
		assert.True(t, runTwo.Prioritize)
		assert.Equal(t, actions_model.MaxRunPriority, runTwo.Priority)

		request = NewRequest(t, "POST", fmt.Sprintf("/%s/actions/runs/2/deprioritize", repo62.FullName()))
		user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		runTwo = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 535682})
		assert.False(t, runTwo.Prioritize)
		assert.Equal(t, actions_model.DefaultRunPriority, runTwo.Priority)
	})

	t.Run("Redirect URL contains supplied parameters", func(t *testing.T) {
		defer tests.PrepareTestEnv(t)()

		unittest.AssertSuccessfulInsert(t, fixtures)

		request := NewRequestWithValues(t, "POST", fmt.Sprintf("/%s/actions/runs/2/deprioritize", repo62.FullName()),
			map[string]string{"actor": "25", "page": "8", "status": "3", "workflow": "build.yaml"})
		response := user2Sess.MakeRequest(t, request, http.StatusSeeOther)

		assert.Equal(t, "/user2/test_workflows/actions?actor=25&page=8&status=3&workflow=build.yaml",
			response.Header().Get("Location"))
	})
}
