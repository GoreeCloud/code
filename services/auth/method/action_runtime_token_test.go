// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package method

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	actions_model "forgejo.org/models/actions"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/services/actions"
	auth_service "forgejo.org/services/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionRuntimeTokenVerify(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("Actions JWT", func(t *testing.T) {
		const RunningTaskID = 47
		task := &actions_model.ActionTask{
			ID: RunningTaskID,
			Job: &actions_model.ActionRunJob{
				ID:    2,
				RunID: 1,
			},
		}
		token, err := actions.CreateAuthorizationToken(task, map[string]any{}, false, &repo_model.ActionsConfig{})
		require.NoError(t, err)

		req := http.Request{
			URL: &url.URL{Path: "/api/v1/"},
			Header: map[string][]string{
				"Authorization": {fmt.Sprintf("Bearer %s", token)},
			},
		}

		o := ActionRuntimeToken{}
		output := o.Verify(&req, nil, nil)
		ar, authSuccess := output.(*auth_service.AuthenticationSuccess)
		require.True(t, authSuccess, "expected type AuthenticationSuccess, but was: %#v", output)
		authResult := ar.Result
		assert.Equal(t, int64(user_model.ActionsUserID), authResult.User().ID)
		isActionsToken, authTaskID := authResult.ActionsTaskID().Get()
		assert.True(t, isActionsToken)
		assert.Equal(t, int64(RunningTaskID), authTaskID)
	})
}

func TestCheckTaskIsRunning(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	cases := map[string]struct {
		TaskID   int64
		Expected bool
	}{
		"Running":   {TaskID: 47, Expected: true},
		"Missing":   {TaskID: 1, Expected: false},
		"Cancelled": {TaskID: 46, Expected: false},
	}

	for name := range cases {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			actual := checkTaskIsRunning(t.Context(), c.TaskID)
			assert.Equal(t, c.Expected, actual)
		})
	}
}
