// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"context"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/util"
	webhook_module "forgejo.org/modules/webhook"

	"xorm.io/builder"
)

// ActionSchedule represents a schedule of a workflow file
type ActionSchedule struct {
	ID                int64
	Title             string
	Specs             []*ActionScheduleSpec  `xorm:"-"`
	RepoID            int64                  `xorm:"index"`
	Repo              *repo_model.Repository `xorm:"-"`
	OwnerID           int64                  `xorm:"index"`
	WorkflowID        string
	WorkflowDirectory string `xorm:"NOT NULL DEFAULT '.forgejo/workflows'"`
	TriggerUserID     int64
	TriggerUser       *user_model.User `xorm:"-"`
	Ref               string
	CommitSHA         string
	Event             webhook_module.HookEventType
	EventPayload      string `xorm:"LONGTEXT"`
	Content           []byte
	Created           timeutil.TimeStamp `xorm:"created"`
	Updated           timeutil.TimeStamp `xorm:"updated"`
}

func init() {
	db.RegisterModel(new(ActionSchedule))
}

// GetSchedulesMapByIDs returns the schedules by given id slice.
func GetSchedulesMapByIDs(ctx context.Context, ids []int64) (map[int64]*ActionSchedule, error) {
	schedules := make(map[int64]*ActionSchedule, len(ids))
	if len(ids) == 0 {
		return schedules, nil
	}
	return schedules, db.GetEngine(ctx).In("id", ids).Find(&schedules)
}

// CreateScheduleTask creates new schedule task.
func CreateScheduleTask(ctx context.Context, rows []*ActionSchedule) error {
	// Return early if there are no rows to insert
	if len(rows) == 0 {
		return nil
	}

	// Begin transaction
	ctx, committer, err := db.TxContext(ctx)
	if err != nil {
		return err
	}
	defer committer.Close()

	// Loop through each schedule row
	for _, row := range rows {
		row.Title, _ = util.SplitStringAtByteN(row.Title, 255)
		// Create new schedule row
		if err = db.Insert(ctx, row); err != nil {
			return err
		}

		for _, spec := range row.Specs {
			spec.ScheduleID = row.ID
			spec.RepoID = row.RepoID

			// Insert the new schedule spec row
			if err = db.Insert(ctx, spec); err != nil {
				return err
			}
		}
	}

	// Commit transaction
	return committer.Commit()
}

func DeleteScheduleTaskByRepo(ctx context.Context, id int64) error {
	ctx, committer, err := db.TxContext(ctx)
	if err != nil {
		return err
	}
	defer committer.Close()

	if _, err := db.GetEngine(ctx).Delete(&ActionSchedule{RepoID: id}); err != nil {
		return err
	}

	if _, err := db.GetEngine(ctx).Delete(&ActionScheduleSpec{RepoID: id}); err != nil {
		return err
	}

	return committer.Commit()
}

type FindScheduleOptions struct {
	db.ListOptions
	RepoID  int64
	OwnerID int64
}

func (opts FindScheduleOptions) ToConds() builder.Cond {
	cond := builder.NewCond()
	if opts.RepoID > 0 {
		cond = cond.And(builder.Eq{"repo_id": opts.RepoID})
	}
	if opts.OwnerID > 0 {
		cond = cond.And(builder.Eq{"owner_id": opts.OwnerID})
	}

	return cond
}

func (opts FindScheduleOptions) ToOrders() string {
	return "`id` DESC"
}
