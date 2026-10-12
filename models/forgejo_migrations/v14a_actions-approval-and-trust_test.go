// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"testing"
	"time"

	"forgejo.org/models/db"
	migration_tests "forgejo.org/models/gitea_migrations/test"
	"forgejo.org/modules/timeutil"
	webhook_module "forgejo.org/modules/webhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_v14ActionsApprovalAndTrustPopulateTableActionUser(t *testing.T) {
	type ConcurrencyMode int
	type Status int

	type ActionUser struct {
		ID     int64 `xorm:"pk autoincr"`
		UserID int64 `xorm:"INDEX UNIQUE(action_user_index) REFERENCES(user, id)"`
		RepoID int64 `xorm:"INDEX UNIQUE(action_user_index) REFERENCES(repository, id)"`

		TrustedWithPullRequests bool

		LastAccess timeutil.TimeStamp `xorm:"INDEX"`
	}
	type ActionRun struct {
		ID            int64
		Title         string
		RepoID        int64  `xorm:"index unique(repo_index) index(concurrency)"`
		OwnerID       int64  `xorm:"index"`
		WorkflowID    string `xorm:"index"`                    // the name of workflow file
		Index         int64  `xorm:"index unique(repo_index)"` // a unique number for each run of a repository
		TriggerUserID int64  `xorm:"index"`
		ScheduleID    int64
		Ref           string `xorm:"index"` // the commit/tag/… that caused the run
		CommitSHA     string
		Event         webhook_module.HookEventType // the webhook event that causes the workflow to run
		EventPayload  string                       `xorm:"LONGTEXT"`
		TriggerEvent  string                       // the trigger event defined in the `on` configuration of the triggered workflow
		Status        Status                       `xorm:"index"`
		Version       int                          `xorm:"version default 0"` // Status could be updated concomitantly, so an optimistic lock is needed
		// Started and Stopped is used for recording last run time, if rerun happened, they will be reset to 0
		Started timeutil.TimeStamp
		Stopped timeutil.TimeStamp
		// PreviousDuration is used for recording previous duration
		PreviousDuration time.Duration
		Created          timeutil.TimeStamp `xorm:"created"`
		Updated          timeutil.TimeStamp `xorm:"updated"`
		NotifyEmail      bool

		// pull request trust
		IsForkPullRequest   bool
		PullRequestPosterID int64
		PullRequestID       int64 `xorm:"index"`
		NeedApproval        bool
		ApprovedBy          int64 `xorm:"index"`

		ConcurrencyGroup string `xorm:"'concurrency_group' index(concurrency)"`
		ConcurrencyType  ConcurrencyMode

		PreExecutionError string `xorm:"LONGTEXT"` // used to report errors that blocked execution of a workflow
	}
	type Repository struct {
		ID int64 `xorm:"pk autoincr"`
	}
	type User struct {
		ID int64 `xorm:"pk autoincr"`
	}
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(User), new(Repository), new(ActionUser), new(ActionRun))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	require.NoError(t, v14ActionsApprovalAndTrustPopulateTableActionUser(x))

	var users []*ActionUser
	require.NoError(t, db.GetEngine(t.Context()).Select("`repo_id`, `user_id`").OrderBy("`id`").Find(&users))
	// See models/gitea_migrations/fixtures/Test_v14ActionsApprovalAndTrustPopulateTableActionUser/action_run.yml
	assert.Equal(t, []*ActionUser{
		{
			UserID: 3,
			RepoID: 15,
		},
		{
			UserID: 3,
			RepoID: 63,
		},
		{
			UserID: 4,
			RepoID: 63,
		},
	}, users)
}
