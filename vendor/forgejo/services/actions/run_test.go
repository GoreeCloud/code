// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActions_CancelOrApproveRun(t *testing.T) {
	t.Run("run, job and task Running changes to run, job and task Cancelled", func(t *testing.T) {
		defer unittest.OverrideFixtures("services/actions/TestActions_CancelOrApproveRun")()
		require.NoError(t, unittest.PrepareTestDatabase())

		taskID := int64(711900)
		task := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: taskID})
		require.Equal(t, actions_model.StatusRunning.String(), task.Status.String())
		job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: task.JobID})
		require.Equal(t, actions_model.StatusRunning.String(), job.Status.String())
		require.Zero(t, job.Stopped)
		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		require.Equal(t, actions_model.StatusRunning.String(), run.Status.String())

		require.NoError(t, CancelRun(t.Context(), run))

		run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		assert.Equal(t, actions_model.StatusCancelled.String(), run.Status.String())
		job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: task.JobID})
		assert.Equal(t, actions_model.StatusCancelled.String(), job.Status.String())
		assert.NotZero(t, job.Stopped)
		task = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: taskID})
		require.Equal(t, actions_model.StatusCancelled.String(), task.Status.String())
	})

	t.Run("run Running, job and task Success changes to run Cancelled", func(t *testing.T) {
		defer unittest.OverrideFixtures("services/actions/TestActions_CancelOrApproveRun")()
		require.NoError(t, unittest.PrepareTestDatabase())

		taskID := int64(710900)
		task := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: taskID})
		require.Equal(t, actions_model.StatusSuccess.String(), task.Status.String())
		job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: task.JobID})
		require.Equal(t, actions_model.StatusSuccess.String(), job.Status.String())
		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		require.Equal(t, actions_model.StatusRunning.String(), run.Status.String())

		require.NoError(t, CancelRun(t.Context(), run))

		run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		assert.Equal(t, actions_model.StatusCancelled.String(), run.Status.String())
		job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: task.JobID})
		assert.Equal(t, actions_model.StatusSuccess, job.Status)
		task = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: taskID})
		require.Equal(t, actions_model.StatusSuccess, task.Status)
	})

	t.Run("run Waiting and job Blocked for Approval changes to run and job Cancelled", func(t *testing.T) {
		defer unittest.OverrideFixtures("services/actions/TestActions_CancelOrApproveRun")()
		require.NoError(t, unittest.PrepareTestDatabase())

		jobID := int64(10800)
		job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: jobID})
		require.Equal(t, actions_model.StatusBlocked.String(), job.Status.String())
		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		require.Equal(t, actions_model.StatusWaiting.String(), run.Status.String())
		require.True(t, run.NeedApproval)

		require.NoError(t, CancelRun(t.Context(), run))

		run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		assert.Equal(t, actions_model.StatusCancelled.String(), run.Status.String())
		assert.False(t, run.NeedApproval)
		job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: jobID})
		assert.Equal(t, actions_model.StatusCancelled, job.Status)
	})

	t.Run("run Waiting and job Blocked for Approval changes to job Waiting", func(t *testing.T) {
		defer unittest.OverrideFixtures("services/actions/TestActions_CancelOrApproveRun")()
		require.NoError(t, unittest.PrepareTestDatabase())

		jobID := int64(10800)
		job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: jobID})
		require.Equal(t, actions_model.StatusBlocked.String(), job.Status.String())
		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		require.Equal(t, actions_model.StatusWaiting.String(), run.Status.String())
		require.True(t, run.NeedApproval)

		doerID := int64(30)
		require.NoError(t, ApproveRun(t.Context(), run, doerID))

		run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: job.RunID})
		assert.Equal(t, actions_model.StatusWaiting.String(), run.Status.String())
		assert.False(t, run.NeedApproval)
		assert.Equal(t, doerID, run.ApprovedBy)
		job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: jobID})
		assert.Equal(t, actions_model.StatusWaiting, job.Status)
	})
}

func TestActions_consistencyCheckRun(t *testing.T) {
	tests := []struct {
		name                     string
		runID                    int64
		errContains              string
		consumed                 bool
		runJobNames              []string
		preExecutionError        actions_model.PreExecutionError
		preExecutionErrorDetails []any
	}{
		{
			name:  "consistent: not incomplete_matrix",
			runID: 900,
		},
		{
			name:  "consistent: incomplete_matrix all needs exist",
			runID: 901,
		},
		{
			name:                     "inconsistent: incomplete_matrix all needs exist",
			runID:                    902,
			preExecutionError:        actions_model.ErrorCodeIncompleteMatrixMissingJob,
			preExecutionErrorDetails: []any{"job_1", "oops-something-wrong-here", "define-matrix"},
		},
		{
			name:                     "inconsistent: static matrix missing dimension",
			runID:                    903,
			preExecutionError:        actions_model.ErrorCodeIncompleteRunsOnMissingMatrixDimension,
			preExecutionErrorDetails: []any{"job_1", "platform-oops-wrong-dimension"},
		},
		{
			name:  "consistent: matrix missing dimension but matrix is dynamic",
			runID: 904,
		},
		{
			name:                     "unknown job in needs",
			runID:                    905,
			preExecutionError:        actions_model.ErrorCodeUnknownJobInNeeds,
			preExecutionErrorDetails: []any{"job_2", "unknown, Job_1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer unittest.OverrideFixtures("services/actions/TestActions_consistencyCheckRun")()
			require.NoError(t, unittest.PrepareTestDatabase())

			run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: tt.runID})

			err := consistencyCheckRun(t.Context(), run)
			require.NoError(t, err)

			run = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: tt.runID})
			assert.Equal(t, tt.preExecutionError, run.PreExecutionErrorCode)
			assert.Equal(t, tt.preExecutionErrorDetails, run.PreExecutionErrorDetails)
		})
	}
}
