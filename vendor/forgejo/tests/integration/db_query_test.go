// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"fmt"
	"testing"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/db"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These are basically unit tests, but by running them in the integration test suite they are tested against all
// supported database types.

func TestDatabaseDefaultMaxInSize(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	// Ensure there are more than db.DefaultMaxInSize objects in a table:
	targetCount := db.DefaultMaxInSize * 2
	for i := range targetCount {
		_, err := actions_model.InsertVariable(t.Context(), 2, 2, fmt.Sprintf("VAR_%d", i), fmt.Sprintf("Value %d", i))
		require.NoError(t, err)
	}

	t.Run("GetByIDs", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		allActionVariables := make([]*actions_model.ActionVariable, 0, targetCount)
		err := db.GetEngine(t.Context()).Find(&allActionVariables)
		require.NoError(t, err)

		allIDs := make([]int64, len(allActionVariables))
		for i := range allActionVariables {
			allIDs[i] = allActionVariables[i].ID
		}

		allActionVariablesAgain, err := db.GetByIDs(t.Context(), "id", allIDs, &actions_model.ActionVariable{})
		require.NoError(t, err)
		assert.Len(t, allActionVariablesAgain, len(allActionVariables))
	})

	t.Run("GetByFieldIn", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		allActionVariables := make([]*actions_model.ActionVariable, 0, targetCount)
		err := db.GetEngine(t.Context()).Find(&allActionVariables)
		require.NoError(t, err)

		allIDs := make([]int64, len(allActionVariables))
		for i := range allActionVariables {
			allIDs[i] = allActionVariables[i].ID
		}

		allActionVariablesAgain, err := db.GetByFieldIn(t.Context(), "id", allIDs, &actions_model.ActionVariable{})
		require.NoError(t, err)
		assert.Len(t, allActionVariablesAgain, len(allActionVariables))
	})
}
