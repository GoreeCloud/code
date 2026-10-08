// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"encoding/base64"
	"testing"

	migration_tests "forgejo.org/models/gitea_migrations/test"
	"forgejo.org/modules/json"
	"forgejo.org/modules/keying"
	"forgejo.org/modules/migration"
	"forgejo.org/modules/structs"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MigrateTaskSecretsToKeying(t *testing.T) {
	type Task struct {
		ID             int64
		DoerID         int64 `xorm:"index"`
		OwnerID        int64 `xorm:"index"`
		RepoID         int64 `xorm:"index"`
		Type           structs.TaskType
		Status         structs.TaskStatus `xorm:"index"`
		StartTime      timeutil.TimeStamp
		EndTime        timeutil.TimeStamp
		PayloadContent string             `xorm:"TEXT"`
		Message        string             `xorm:"TEXT"`
		Created        timeutil.TimeStamp `xorm:"created"`
	}

	// Prepare and load the testing database
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(Task))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	cnt, err := x.Table("task").Count()
	require.NoError(t, err)
	assert.EqualValues(t, 3, cnt)

	require.NoError(t, migrateTaskSecrets(x))

	cnt, err = x.Table("task").Count()
	require.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	var task Task
	_, err = x.Table("task").ID(1).Get(&task)
	require.NoError(t, err)

	var opts migration.MigrateOptions
	require.NoError(t, json.Unmarshal([]byte(task.PayloadContent), &opts))
	key := keying.MigrateTask

	encryptedCloneAddr, err := base64.RawStdEncoding.DecodeString(opts.CloneAddrEncrypted)
	require.NoError(t, err)
	cloneAddr, err := key.Decrypt(encryptedCloneAddr, keying.ColumnAndJSONSelectorAndID("payload_content", "clone_addr_encrypted", task.ID))
	require.NoError(t, err)
	assert.Equal(t, "https://admin:password@example.com", string(cloneAddr))

	encryptedAuthPassword, err := base64.RawStdEncoding.DecodeString(opts.AuthPasswordEncrypted)
	require.NoError(t, err)
	authPassword, err := key.Decrypt(encryptedAuthPassword, keying.ColumnAndJSONSelectorAndID("payload_content", "auth_password_encrypted", task.ID))
	require.NoError(t, err)
	assert.Equal(t, "password", string(authPassword))

	encryptedAuthToken, err := base64.RawStdEncoding.DecodeString(opts.AuthTokenEncrypted)
	require.NoError(t, err)
	authToken, err := key.Decrypt(encryptedAuthToken, keying.ColumnAndJSONSelectorAndID("payload_content", "auth_token_encrypted", task.ID))
	require.NoError(t, err)
	assert.Equal(t, "token", string(authToken))
}
