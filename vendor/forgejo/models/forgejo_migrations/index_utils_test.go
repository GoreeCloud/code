// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"testing"

	migration_tests "forgejo.org/models/gitea_migrations/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDropIndexIfExists(t *testing.T) {
	type Table struct {
		ID      int64 `xorm:"pk"`
		DoerID  int64 `xorm:"INDEX INDEX(s)"`
		OwnerID int64 `xorm:"INDEX"`
		RepoID  int64 `xorm:"INDEX(s)"`
	}

	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(Table))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	exists, err := indexExists(x, "table", "IDX_table_doer_id")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = indexExists(x, "table", "IDX_table_owner_id")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = indexExists(x, "table", "IDX_table_repo_id")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = indexExists(x, "table", "IDX_table_s")
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, dropIndexIfExists(x, "table", "IDX_table_repo_id"))

	require.NoError(t, dropIndexIfExists(x, "table", "IDX_table_doer_id"))
	exists, err = indexExists(x, "table", "IDX_table_doer_id")
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, dropIndexIfExists(x, "table", "IDX_table_s"))
	exists, err = indexExists(x, "table", "IDX_table_s")
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, dropIndexIfExists(x, "table", "IDX_table_owner_id"))
	exists, err = indexExists(x, "table", "IDX_table_owner_id")
	require.NoError(t, err)
	assert.False(t, exists)
}
