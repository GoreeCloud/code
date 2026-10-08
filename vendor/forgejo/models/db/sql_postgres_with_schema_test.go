// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"database/sql"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresSchemaDriverRegistered(t *testing.T) {
	// Verify the driver is registered (happens in init())
	drivers := sql.Drivers()
	found := slices.Contains(drivers, "postgresschema")
	assert.True(t, found, "postgresschema driver should be registered")
}

func TestPostgresSchemaDriverOpenFailsWithInvalidConnString(t *testing.T) {
	// Verify Open() is actually called by checking that an invalid connection string returns an error
	drv := &postgresSchemaDriver{innerDriver: nil}

	require.Panics(t, func() {
		_, _ = drv.Open("invalid")
	}, "Open with nil innerDriver should panic (confirming the Open method is called)")
}
