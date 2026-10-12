// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	actions_model "forgejo.org/models/actions"
	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	unit_model "forgejo.org/models/unit"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/gitrepo"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"
	actions_service "forgejo.org/services/actions"
	"forgejo.org/tests/forgery"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionGetTokenMetadata(t *testing.T) {
	if !setting.Database.Type.IsSQLite3() {
		t.Skip()
	}

	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
		repo := forgery.CreateRepository(t, user2, &forgery.CreateRepositoryOptions{
			Files: forgery.MapFS{
				".forgejo/workflows/dispatch.yml": forgery.MapFile(
					"name: test\n" +
						"enable-email-notifications: true\n" +
						"on: [workflow_dispatch]\n" +
						"jobs:\n" +
						"  test:\n" +
						"    runs-on: ubuntu-latest\n" +
						"    steps:\n" +
						"      - run: echo helloworld\n",
				),
			},
		})
		forgery.EnableRepoUnits(t, repo, unit_model.TypeActions)

		gitRepo, err := gitrepo.OpenRepository(db.DefaultContext, repo)
		require.NoError(t, err)
		defer gitRepo.Close()

		workflow, err := actions_service.GetWorkflowFromCommit(gitRepo, "main", "dispatch.yml")
		require.NoError(t, err)
		assert.Equal(t, "refs/heads/main", workflow.Ref)

		inputGetter := func(key string) string {
			return ""
		}

		runner := newMockRunner()
		runner.registerAsRepoRunner(t, user2.Name, repo.Name, "mock-runner", []string{"ubuntu-latest"})

		_, _, err = workflow.Dispatch(db.DefaultContext, inputGetter, repo, user2)
		require.NoError(t, err)
		runnerTask := runner.fetchTask(t)

		task := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: runnerTask.GetId()})
		require.NoError(t, task.LoadJob(context.Background()))
		require.NoError(t, task.Job.LoadRun(context.Background()))

		t.Run("success", func(t *testing.T) {
			req := NewRequest(t, "GET", "/api/v1/actions/run")
			req.AddTokenAuth(runnerTask.Secrets["FORGEJO_TOKEN"])
			resp := MakeRequest(t, req, http.StatusOK)
			var payload api.ActionRun
			DecodeJSON(t, resp, &payload)
			assert.Equal(t, task.Job.Run.ID, payload.ID)
			assert.Equal(t, "workflow_dispatch", payload.TriggerEvent)
		})

		t.Run("failure (invalid token)", func(t *testing.T) {
			req := NewRequest(t, "GET", "/api/v1/actions/run")
			req.AddTokenAuth("0000000000000000000000000000000000000000")
			MakeRequest(t, req, http.StatusUnauthorized)
		})

		t.Run("failure (PAT instead of automatic token)", func(t *testing.T) {
			session := loginUser(t, "user2")
			repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1, OwnerID: user2.ID})
			userToken := getTokenForLoggedInUser(t, session,
				auth_model.AccessTokenScopePublicOnly, auth_model.AccessTokenScopeReadRepository)
			req := NewRequest(t, "GET",
				fmt.Sprintf("/api/v1/repos/%s/actions/runs", repo.FullName()))
			req.AddTokenAuth(userToken)
			MakeRequest(t, req, http.StatusOK) // make sure the token works

			req = NewRequest(t, "GET", "/api/v1/actions/run")
			req.AddTokenAuth(userToken)
			MakeRequest(t, req, http.StatusForbidden)
		})

		t.Run("failure (no token)", func(t *testing.T) {
			req := NewRequest(t, "GET", "/api/v1/actions/run")
			MakeRequest(t, req, http.StatusForbidden)
		})
	})
}
