// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"
	"time"

	"forgejo.org/models/unittest"

	runnerv1 "code.forgejo.org/forgejo/actions-proto/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRemoveLogs(t *testing.T) {
	t.Run("Logs removed", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())

		_, err := WriteLogs(t.Context(), "test.log", 0,
			[]*runnerv1.LogRow{{Time: timestamppb.New(time.Now()), Content: "Hello world"}})
		require.NoError(t, err)

		exist, err := ExistsLogs(t.Context(), "test.log")
		require.NoError(t, err)

		assert.True(t, exist)

		require.NoError(t, RemoveLogs(t.Context(), false, "test.log"))

		exist, err = ExistsLogs(t.Context(), "test.log")
		require.NoError(t, err)

		assert.False(t, exist)
	})

	t.Run("Error if filename is empty", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())

		_, err := WriteLogs(t.Context(), "test.log", 0,
			[]*runnerv1.LogRow{{Time: timestamppb.New(time.Now()), Content: "Hello world"}})
		require.NoError(t, err)

		exist, err := ExistsLogs(t.Context(), "test.log")
		require.NoError(t, err)

		assert.True(t, exist)

		err = RemoveLogs(t.Context(), false, "")
		require.ErrorContains(t, err, "cannot remove logs because filename is empty")

		exist, err = ExistsLogs(t.Context(), "test.log")
		require.NoError(t, err)

		assert.True(t, exist)
	})
}
