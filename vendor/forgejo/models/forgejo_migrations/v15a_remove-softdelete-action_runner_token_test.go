// Copyright 2026 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"testing"

	"forgejo.org/models/db"
	migration_tests "forgejo.org/models/gitea_migrations/test"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_removeSoftDeleteActionRunnerToken(t *testing.T) {
	type ActionRunnerToken struct {
		ID       int64
		Token    string `xorm:"UNIQUE"`
		OwnerID  int64  `xorm:"index"`
		RepoID   int64  `xorm:"index"`
		IsActive bool
		Created  timeutil.TimeStamp `xorm:"created"`
		Updated  timeutil.TimeStamp `xorm:"updated"`
		Deleted  timeutil.TimeStamp `xorm:"deleted"`
	}
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(ActionRunnerToken))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	require.NoError(t, removeSoftDeleteActionRunnerToken(x))

	var remainingRecords []*ActionRunnerToken
	require.NoError(t,
		db.GetEngine(t.Context()).
			Table("action_runner_token").
			Select("`id`, `owner_id`, `repo_id`").
			OrderBy("`id`").
			Unscoped(). // `Deleted` column doesn't exist anymore, so don't include in query
			Find(&remainingRecords))
	assert.Equal(t,
		[]*ActionRunnerToken{
			{ID: 4},
			{ID: 5, OwnerID: 1},
			{ID: 6, RepoID: 1},
		},
		remainingRecords)
}
