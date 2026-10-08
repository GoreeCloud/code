// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"context"
	"fmt"
	"time"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/db"
	"forgejo.org/modules/actions"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"
)

var (
	transferLingeringLogsMax   = 3000
	transferLingeringLogsSleep = 1 * time.Second
	transferLingeringLogsOld   = 24 * time.Hour
)

func TransferLingeringLogs(ctx context.Context) error {
	return transferLingeringLogs(ctx, transferLingeringLogsOpts(time.Now()))
}

func transferLingeringLogsOpts(now time.Time) actions_model.FindTaskOptions {
	// performance considerations: the search is linear because
	// LogInStorage has no index. But it is bounded by
	// LogExpired which is always true for older records and has an index.
	return actions_model.FindTaskOptions{
		Status:       actions_model.DoneStatuses(),
		LogInStorage: optional.Some(false),
		LogExpired:   optional.Some(false),
		// do it after a long delay to avoid any possibility of race with an ongoing operation
		// as it is not protected by a transaction
		UpdatedBefore: timeutil.TimeStamp(now.Add(-transferLingeringLogsOld).Unix()),
	}
}

func transferLingeringLogs(ctx context.Context, opts actions_model.FindTaskOptions) error {
	count := 0
	err := db.Iterate(ctx, opts.ToConds(), func(ctx context.Context, task *actions_model.ActionTask) error {
		if err := TransferLogsAndUpdateLogInStorage(ctx, task); err != nil {
			return err
		}
		log.Debug("processed task %d", task.ID)
		count++
		if count < transferLingeringLogsMax {
			log.Debug("sleeping %v to not stress the storage", transferLingeringLogsSleep)
			time.Sleep(transferLingeringLogsSleep)
		}
		if count >= transferLingeringLogsMax {
			return fmt.Errorf("stopped after processing %v tasks and will resume later", transferLingeringLogsMax)
		}
		return nil
	})
	if count >= transferLingeringLogsMax {
		log.Info("%v", err)
		return nil
	}
	if count > 0 {
		log.Info("processed %d tasks", count)
	}
	return err
}

func TransferLogsAndUpdateLogInStorage(ctx context.Context, task *actions_model.ActionTask) error {
	if task.LogInStorage {
		return nil
	}
	remove, err := TransferLogs(ctx, task.LogFilename)
	if err != nil {
		return err
	}
	task.LogInStorage = true
	if err := actions_model.UpdateTask(ctx, task, "log_in_storage"); err != nil {
		return err
	}
	remove()

	return nil
}

func TransferLogs(ctx context.Context, logFilename string) (func(), error) {
	exists, err := actions.ExistsLogs(ctx, logFilename)
	if err != nil {
		return nil, err
	}
	if !exists {
		return func() {}, nil
	}
	return actions.TransferLogs(ctx, logFilename)
}
