// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package webhook

import (
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestListSystemWebhooks(t *testing.T) {
	defer unittest.OverrideFixtures("models/webhook/fixtures/TestListSystemWebhooks")()
	require.NoError(t, unittest.PrepareTestDatabase())
	testGetAdminWebhooks(t, true)
}

func TestListDefaultWebhooks(t *testing.T) {
	defer unittest.OverrideFixtures("models/webhook/fixtures/TestListDefaultWebhooks")()
	require.NoError(t, unittest.PrepareTestDatabase())
	testGetAdminWebhooks(t, false)
}

func testGetAdminWebhooks(t *testing.T, systemHooks bool) {
	hookCond := builder.Eq{"is_system_webhook": systemHooks}.And(builder.Eq{"repo_id": 0})
	unittest.AssertCountByCond(t, "webhook", hookCond, 5)

	t.Run("All hooks (Implicit)", func(t *testing.T) {
		hooks, count, err := getAdminWebhooks(t.Context(), systemHooks, db.ListOptionsAll)
		require.NoError(t, err)
		unittest.AssertCountByCond(t, "webhook", hookCond, int(count))
		require.Len(t, hooks, int(count))
		require.Equal(t, 5, int(count))
		for i := range hooks {
			require.Equal(t, systemHooks, hooks[i].IsSystemWebhook)
		}
	})

	t.Run("All hooks (Explicit)", func(t *testing.T) {
		hooks, count, err := getAdminWebhooks(t.Context(), systemHooks, db.ListOptionsAll, false)
		require.NoError(t, err)
		unittest.AssertCountByCond(t, "webhook", hookCond, int(count))
		require.Len(t, hooks, int(count))
		require.Equal(t, 5, int(count))
		for i := range hooks {
			require.Equal(t, systemHooks, hooks[i].IsSystemWebhook)
		}
	})

	t.Run("Active hooks", func(t *testing.T) {
		hooks, count, err := getAdminWebhooks(t.Context(), systemHooks, db.ListOptionsAll, true)
		require.NoError(t, err)
		unittest.AssertCountByCond(t, "webhook", hookCond.And(builder.Eq{"is_active": true}), int(count))
		require.Len(t, hooks, int(count))
		require.Equal(t, 3, int(count))
		for i := range hooks {
			require.Equal(t, systemHooks, hooks[i].IsSystemWebhook)
			require.True(t, hooks[i].IsActive)
		}
	})
}
