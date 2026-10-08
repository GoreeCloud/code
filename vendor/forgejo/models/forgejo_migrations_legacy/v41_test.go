// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations_legacy

import (
	"testing"

	"forgejo.org/models/db"
	migration_tests "forgejo.org/models/gitea_migrations/test"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AddForeignKeysStopwatchTrackedTime(t *testing.T) {
	type Stopwatch struct {
		ID          int64              `xorm:"pk autoincr"`
		IssueID     int64              `xorm:"INDEX"`
		UserID      int64              `xorm:"INDEX"`
		CreatedUnix timeutil.TimeStamp `xorm:"created"`
	}
	type TrackedTime struct {
		ID          int64 `xorm:"pk autoincr"`
		IssueID     int64 `xorm:"INDEX"`
		UserID      int64 `xorm:"INDEX"`
		CreatedUnix int64 `xorm:"created"`
		Time        int64 `xorm:"NOT NULL"`
		Deleted     bool  `xorm:"NOT NULL DEFAULT false"`
	}
	type User struct {
		ID int64 `xorm:"pk autoincr"`
	}
	type Issue struct {
		ID int64 `xorm:"pk autoincr"`
	}
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(User), new(Issue), new(Stopwatch), new(TrackedTime))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	require.NoError(t, AddForeignKeysStopwatchTrackedTime(x))

	var remainingStopwatch []*Stopwatch
	require.NoError(t,
		db.GetEngine(t.Context()).
			Table("stopwatch").
			Select("`id`, `issue_id`, `user_id`").
			OrderBy("`id`").
			Find(&remainingStopwatch))
	assert.Equal(t,
		[]*Stopwatch{
			{1, 1, 1, 0},
		},
		remainingStopwatch,
		"stopwatch")

	var remainingTrackedTime []*TrackedTime
	require.NoError(t,
		db.GetEngine(t.Context()).
			Table("tracked_time").
			Select("`id`, `issue_id`, `user_id`").
			OrderBy("`id`").
			Find(&remainingTrackedTime))
	assert.Equal(t,
		[]*TrackedTime{
			{ID: 1, IssueID: 1, UserID: 1},
			{ID: 4, IssueID: 1, UserID: 0},
			{ID: 5, IssueID: 1, UserID: 0},
		},
		remainingTrackedTime,
		"tracked_time")
}
