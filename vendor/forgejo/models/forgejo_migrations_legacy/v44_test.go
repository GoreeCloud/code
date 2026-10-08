// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations_legacy

import (
	"testing"

	"forgejo.org/models/db"
	migration_tests "forgejo.org/models/gitea_migrations/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AddForeignKeysAccess(t *testing.T) {
	type AccessMode int
	type Access struct {
		ID     int64 `xorm:"pk autoincr"`
		UserID int64 `xorm:"UNIQUE(s)"`
		RepoID int64 `xorm:"UNIQUE(s)"`
		Mode   AccessMode
	}
	type User struct {
		ID int64 `xorm:"pk autoincr"`
	}
	type Repository struct {
		ID int64 `xorm:"pk autoincr"`
	}
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(User), new(Repository), new(Access))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	require.NoError(t, AddForeignKeysAccess(x))

	var remainingRecords []*Access
	require.NoError(t,
		db.GetEngine(t.Context()).
			Table("access").
			Select("`id`, `user_id`, `repo_id`").
			OrderBy("`id`").
			Find(&remainingRecords))
	assert.Equal(t,
		[]*Access{
			{ID: 1, UserID: 1, RepoID: 1},
		},
		remainingRecords)
}
