// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	actions_model "forgejo.org/models/actions"
	secret_model "forgejo.org/models/secret"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSecretsOfJob(t *testing.T) {
	tests := []struct {
		name     string
		runJobID int64
		secrets  map[string]string
	}{
		{
			name:     "push run",
			runJobID: 600,
			secrets: map[string]string{
				"SECRET_1": "the sky is blue",
				"SECRET_2": "the ocean is also blue",
			},
		},
		{
			name:     "on: pull_request_target workflow, local PR (not fork)",
			runJobID: 601,
			secrets: map[string]string{
				"SECRET_1": "the sky is blue",
				"SECRET_2": "the ocean is also blue",
			},
		},
		{
			name:     "on: pull_request_target workflow, fork PR",
			runJobID: 602,
			secrets: map[string]string{
				"SECRET_1": "the sky is blue",
				"SECRET_2": "the ocean is also blue",
			},
		},
		{
			name:     "on: pull_request workflow, local PR (not fork)",
			runJobID: 603,
			secrets: map[string]string{
				"SECRET_1": "the sky is blue",
				"SECRET_2": "the ocean is also blue",
			},
		},
		{
			name:     "on: pull_request workflow, fork PR",
			runJobID: 604,
			secrets:  map[string]string{},
		},
		{
			name:     "workflow call inner job, inherit secrets",
			runJobID: 605,
			secrets: map[string]string{
				"SECRET_1": "the sky is blue",
				"SECRET_2": "the ocean is also blue",
			},
		},
		{
			name:     "workflow call two layer inner job, inherit secrets",
			runJobID: 607,
			secrets: map[string]string{
				// Even though we're 'inherit' in this case, we're inheriting from the parent call which is a subset
				// (and modification) of the secrets -- so shouldn't see SECRET_2.
				"SECRET_1": "the sky is blue -- but are you sure?",
			},
		},
		{
			name:     "workflow call inner job, defined secrets",
			runJobID: 610,
			secrets: map[string]string{
				"FORGEJO":  "context forgejo = refs/heads/main",
				"INPUTS":   "context inputs = some_wd_input_value",
				"MATRIX":   "context matrix = some-dimension-value",
				"NEEDS":    "context needs = abcdefghijklmnopqrstuvwxyz",
				"SECRETS":  "context secrets = the sky is blue",
				"STRATEGY": "context strategy = false",
				"VARS":     "context vars = this is a repo variable",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer unittest.OverrideFixtures("services/actions/TestGetSecretsOfJob")()
			require.NoError(t, unittest.PrepareTestDatabase())

			// Due to encryption, more maintainable to do this rather than create secrets in fixture data
			_, err := secret_model.InsertEncryptedSecret(t.Context(), 2, 0, "secret_1", "the sky is blue")
			require.NoError(t, err)
			_, err = secret_model.InsertEncryptedSecret(t.Context(), 0, 63, "secret_2", "the ocean is also blue")
			require.NoError(t, err)

			runJob := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: tt.runJobID})
			actualSecrets, err := getSecretsOfJob(t.Context(), runJob)
			require.NoError(t, err)
			assert.Equal(t, tt.secrets, actualSecrets)
		})
	}
}
