// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"testing"

	migration_tests "forgejo.org/models/gitea_migrations/test"
	"forgejo.org/modules/keying"
	"forgejo.org/modules/timeutil"
	webhook_module "forgejo.org/modules/webhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_MigrateWebhookSecrets(t *testing.T) {
	type HookContentType int
	type Webhook struct {
		ID              int64 `xorm:"pk autoincr"`
		RepoID          int64 `xorm:"INDEX"`
		OwnerID         int64 `xorm:"INDEX"`
		IsSystemWebhook bool
		URL             string `xorm:"url TEXT"`
		HTTPMethod      string `xorm:"http_method"`
		ContentType     HookContentType
		Secret          string                  `xorm:"TEXT"`
		Events          string                  `xorm:"TEXT"`
		IsActive        bool                    `xorm:"INDEX"`
		Type            webhook_module.HookType `xorm:"VARCHAR(16) 'type'"`
		Meta            string                  `xorm:"TEXT"`
		LastStatus      webhook_module.HookStatus

		HeaderAuthorizationEncrypted string `xorm:"TEXT"`

		CreatedUnix timeutil.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"INDEX updated"`
	}

	type NewWebhook struct {
		ID              int64 `xorm:"pk autoincr"`
		RepoID          int64 `xorm:"INDEX"`
		OwnerID         int64 `xorm:"INDEX"`
		IsSystemWebhook bool
		URL             string `xorm:"url TEXT"`
		HTTPMethod      string `xorm:"http_method"`
		ContentType     HookContentType
		Secret          string                  `xorm:"TEXT"`
		Events          string                  `xorm:"TEXT"`
		IsActive        bool                    `xorm:"INDEX"`
		Type            webhook_module.HookType `xorm:"VARCHAR(16) 'type'"`
		Meta            string                  `xorm:"TEXT"`
		LastStatus      webhook_module.HookStatus

		HeaderAuthorizationEncrypted []byte `xorm:"BLOB"`

		CreatedUnix timeutil.TimeStamp `xorm:"INDEX created"`
		UpdatedUnix timeutil.TimeStamp `xorm:"INDEX updated"`
	}

	// Prepare and load the testing database
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(Webhook))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	cnt, err := x.Table("webhook").Count()
	require.NoError(t, err)
	assert.EqualValues(t, 3, cnt)

	require.NoError(t, migrateWebhookSecrets(x))

	cnt, err = x.Table("webhook").Count()
	require.NoError(t, err)
	assert.EqualValues(t, 2, cnt)

	key := keying.Webhook

	t.Run("webhook 1", func(t *testing.T) {
		var webhook NewWebhook
		_, err = x.Table("webhook").ID(1).Get(&webhook)
		require.NoError(t, err)

		secret, err := key.Decrypt(webhook.HeaderAuthorizationEncrypted, keying.ColumnAndID("header_authorization_encrypted", webhook.ID))
		require.NoError(t, err)
		assert.EqualValues(t, "Bearer s3cr3t", secret)
	})

	t.Run("webhook 3", func(t *testing.T) {
		var webhook NewWebhook
		_, err = x.Table("webhook").ID(3).Get(&webhook)
		require.NoError(t, err)
		assert.Empty(t, webhook.HeaderAuthorizationEncrypted)
	})
}
