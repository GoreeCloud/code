// Copyright 2026 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/require"
)

func TestDeleteJobsOfRun(t *testing.T) {
	t.Run("Deletes completed job", func(t *testing.T) {
		defer unittest.OverrideFixtures("services/actions/TestDeleteJobsOfRun")()
		require.NoError(t, unittest.PrepareTestDatabase())

		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 34901})
		job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 47301, RunID: run.ID})
		unittest.AssertCount(t, &actions_model.ActionTask{JobID: job.ID}, 1)

		require.NoError(t, deleteJobsOfRun(t.Context(), run.ID))

		unittest.AssertNotExistsBean(t, &actions_model.ActionRunJob{ID: job.ID})
		unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 47302})
		unittest.AssertCount(t, &actions_model.ActionTask{JobID: job.ID}, 0)
	})

	t.Run("Error if job has not completed", func(t *testing.T) {
		defer unittest.OverrideFixtures("services/actions/TestDeleteJobsOfRun")()
		require.NoError(t, unittest.PrepareTestDatabase())

		run := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRun{ID: 34902})
		job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 47302, RunID: run.ID})
		unittest.AssertCount(t, &actions_model.ActionTask{JobID: job.ID}, 1)

		err := deleteJobsOfRun(t.Context(), run.ID)
		require.ErrorContains(t, err, "unable to delete job 47302 because it has not completed yet")

		unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: job.ID})
		unittest.AssertCount(t, &actions_model.ActionTask{JobID: job.ID}, 1)
	})
}
