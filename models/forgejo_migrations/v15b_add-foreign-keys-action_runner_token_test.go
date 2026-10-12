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

func Test_addForeignKeysActionRunnerToken(t *testing.T) {
	type ActionRunnerToken struct {
		ID       int64
		Token    string `xorm:"UNIQUE"`
		OwnerID  int64  `xorm:"index"`
		RepoID   int64  `xorm:"index"`
		IsActive bool
		Created  timeutil.TimeStamp `xorm:"created"`
		Updated  timeutil.TimeStamp `xorm:"updated"`
	}
	type User struct {
		ID int64 `xorm:"pk autoincr"`
	}
	type Repository struct {
		ID int64 `xorm:"pk autoincr"`
	}
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(User), new(Repository), new(ActionRunnerToken))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	require.NoError(t, addForeignKeysActionRunnerToken(x))

	var remainingRecords []*ActionRunnerToken
	require.NoError(t,
		db.GetEngine(t.Context()).
			Table("action_runner_token").
			Select("`id`, `owner_id`, `repo_id`").
			OrderBy("`id`").
			Find(&remainingRecords))
	assert.Equal(t,
		[]*ActionRunnerToken{
			{ID: 1},
			{ID: 2, OwnerID: 1},
			{ID: 3, RepoID: 1},
		},
		remainingRecords)
}
