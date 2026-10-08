// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations_legacy

import (
	"testing"

	migration_tests "forgejo.org/models/gitea_migrations/test"
	"forgejo.org/models/secret"
	"forgejo.org/modules/keying"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MigrateActionSecretToKeying(t *testing.T) {
	type Secret struct {
		ID          int64
		OwnerID     int64              `xorm:"INDEX UNIQUE(owner_repo_name) NOT NULL"`
		RepoID      int64              `xorm:"INDEX UNIQUE(owner_repo_name) NOT NULL DEFAULT 0"`
		Name        string             `xorm:"UNIQUE(owner_repo_name) NOT NULL"`
		Data        string             `xorm:"LONGTEXT"` // encrypted data
		CreatedUnix timeutil.TimeStamp `xorm:"created NOT NULL"`
	}

	// Prepare and load the testing database
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(Secret))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	cnt, err := x.Table("secret").Count()
	require.NoError(t, err)
	assert.EqualValues(t, 2, cnt)

	require.NoError(t, MigrateActionSecretsToKeying(x))

	cnt, err = x.Table("secret").Count()
	require.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var secret secret.Secret
	_, err = x.Table("secret").ID(1).Get(&secret)
	require.NoError(t, err)

	secretBytes, err := keying.ActionSecret.Decrypt(secret.Data, keying.ColumnAndID("data", secret.ID))
	require.NoError(t, err)
	assert.Equal(t, []byte("A deep dark secret"), secretBytes)
}
