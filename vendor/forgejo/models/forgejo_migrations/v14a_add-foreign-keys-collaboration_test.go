// Copyright 2025 The Forgejo Authors. All rights reserved.
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

func Test_addForeignKeysCollaboration(t *testing.T) {
	type AccessMode int
	type Collaboration struct {
		ID          int64              `xorm:"pk autoincr"`
		RepoID      int64              `xorm:"UNIQUE(s) INDEX NOT NULL"`
		UserID      int64              `xorm:"UNIQUE(s) INDEX NOT NULL"`
		Mode        AccessMode         `xorm:"DEFAULT 2 NOT NULL"`
		CreatedUnix timeutil.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"INDEX updated"`
	}
	type Repository struct {
		ID int64 `xorm:"pk autoincr"`
	}
	type User struct {
		ID int64 `xorm:"pk autoincr"`
	}
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(User), new(Repository), new(Collaboration))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	require.NoError(t, addForeignKeysCollaboration(x))

	var remainingRecords []*Collaboration
	require.NoError(t,
		db.GetEngine(t.Context()).
			Table("collaboration").
			Select("`id`, `repo_id`, `user_id`").
			OrderBy("`id`").
			Find(&remainingRecords))
	assert.Equal(t,
		[]*Collaboration{
			{ID: 1, UserID: 1, RepoID: 1},
		},
		remainingRecords)
}
