// Copyright 2026 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"context"
	"fmt"

	"forgejo.org/models/actions"
	"forgejo.org/models/db"
)

// deleteJobsOfRun removes all jobs that belong to the given run, including its associated tasks. Each job has to be
// completed for the operation to succeed.
func deleteJobsOfRun(ctx context.Context, runID int64) error {
	return db.WithTx(ctx, func(ctx context.Context) error {
		jobs, err := actions.GetRunJobsByRunID(ctx, runID)
		if err != nil {
			return fmt.Errorf("unable to load jobs of run %d: %w", runID, err)
		}

		for _, job := range jobs {
			if !job.Status.IsDone() {
				return fmt.Errorf("unable to delete job %d because it has not completed yet", job.ID)
			}

			tasks, err := actions.GetTasksOfJob(ctx, job.ID)
			if err != nil {
				return err
			}
			for _, task := range tasks {
				err = deleteTask(ctx, task.ID)
				if err != nil {
					return err
				}
			}

			err = actions.DeleteJob(ctx, job.ID)
			if err != nil {
				return fmt.Errorf("unable to delete job %d of run %d: %w", job.ID, job.RunID, err)
			}
		}

		return nil
	})
}
