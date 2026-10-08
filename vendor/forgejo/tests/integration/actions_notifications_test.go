// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"fmt"
	"net/url"
	"testing"

	actions_model "forgejo.org/models/actions"
	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/setting"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestActionsNotifications(t *testing.T) {
	if !setting.Database.Type.IsSQLite3() {
		t.Skip()
	}

	testCases := []struct {
		name        string
		treePath    string
		fileContent string
		notifyEmail bool
	}{
		{
			name:     "enabled",
			treePath: ".forgejo/workflows/enabled.yml",
			fileContent: `name: enabled
on:
  push:
enable-email-notifications: true
jobs:
  job1:
    runs-on: ubuntu-latest
    steps:
      - run: echo job1
`,
			notifyEmail: true,
		},
		{
			name:     "disabled",
			treePath: ".forgejo/workflows/disabled.yml",
			fileContent: `name: disabled
on:
  push:
jobs:
  job1:
    runs-on: ubuntu-latest
    steps:
      - run: echo job1
`,
			notifyEmail: false,
		},
	}
	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
		session := loginUser(t, user2.Name)
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository, auth_model.AccessTokenScopeWriteUser)

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				defer tests.PrintCurrentTest(t)()

				apiRepo := createActionsTestRepo(t, token, testCase.name, false)
				runner := newMockRunner()
				runner.registerAsRepoRunner(t, user2.Name, apiRepo.Name, "mock-runner", []string{"ubuntu-latest"})
				opts := getWorkflowCreateFileOptions(user2, apiRepo.DefaultBranch, fmt.Sprintf("create %s", testCase.treePath), testCase.fileContent)
				createWorkflowFile(t, token, user2.Name, apiRepo.Name, testCase.treePath, opts)

				task := runner.fetchTask(t)
				actionTask := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: task.Id})
				actionRunJob := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: actionTask.JobID})
				actionRun := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: actionRunJob.RunID})
				assert.Equal(t, testCase.notifyEmail, actionRun.NotifyEmail)

				httpContext := NewAPITestContext(t, user2.Name, apiRepo.Name, auth_model.AccessTokenScopeWriteUser)
				doAPIDeleteRepository(httpContext)(t)
			})
		}
	})
}
