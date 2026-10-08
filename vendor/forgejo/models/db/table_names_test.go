// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"slices"
	"testing"

	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
)

func TestGetTableNames(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		defer test.MockVariableValue(&tables, []any{new(GPGKey)})()

		assert.Equal(t, []string{"gpg_key"}, GetTableNames().Values())
	})

	t.Run("Multiple tables", func(t *testing.T) {
		defer test.MockVariableValue(&tables, []any{new(GPGKey), new(User), new(BlockedUser)})()

		tableNames := GetTableNames().Values()
		slices.Sort(tableNames)

		assert.Equal(t, []string{"forgejo_blocked_user", "gpg_key", "user"}, tableNames)
	})
}

type GPGKey struct{}

type User struct{}

type BlockedUser struct{}

func (*BlockedUser) TableName() string {
	return "forgejo_blocked_user"
}
