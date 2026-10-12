// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package project

import (
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveIssuesOnProjectColumn(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// Get column 1 which belongs to project 1 and has issue 1
	column := unittest.AssertExistsAndLoadBean(t, &Column{ID: 1})
	require.Equal(t, int64(1), column.ProjectID)

	t.Run("Success", func(t *testing.T) {
		// Issue 1 is in column 1 (from fixtures)
		sortedIssueIDs := map[int64]int64{
			0: 1, // sorting position 0 -> issue_id 1
		}
		err := MoveIssuesOnProjectColumn(db.DefaultContext, column, sortedIssueIDs)
		require.NoError(t, err)

		// Verify the sorting was updated using direct DB query
		var card ProjectIssue
		has, err := db.GetEngine(db.DefaultContext).Where("project_id=? AND issue_id=?", column.ProjectID, 1).Get(&card)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, int64(0), card.Sorting)
	})

	t.Run("MoveIssueFromDifferentColumn", func(t *testing.T) {
		// Issue 3 is in column 2, not column 1 — but same project, so cross-column move should succeed
		sortedIssueIDs := map[int64]int64{
			0: 3,
		}
		err := MoveIssuesOnProjectColumn(db.DefaultContext, column, sortedIssueIDs)
		require.NoError(t, err)

		// Verify the card was moved to column 1 and sorting updated
		var card ProjectIssue
		has, err := db.GetEngine(db.DefaultContext).Where("project_id=? AND issue_id=?", column.ProjectID, 3).Get(&card)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, column.ID, card.ProjectColumnID)
		assert.Equal(t, int64(0), card.Sorting)
	})

	t.Run("ErrorIssueNotInProject", func(t *testing.T) {
		// Issue 999 doesn't exist
		sortedIssueIDs := map[int64]int64{
			0: 999,
		}
		err := MoveIssuesOnProjectColumn(db.DefaultContext, column, sortedIssueIDs)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "all issues must belong to the specified project")
	})
}

func TestMoveIssuesOnProjectColumnSwap(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	column := unittest.AssertExistsAndLoadBean(t, &Column{ID: 1})

	// Setup: insert two cards at distinct positions using direct DB inserts
	card1 := &ProjectIssue{
		IssueID:         14,
		ProjectID:       1,
		ProjectColumnID: column.ID,
		Sorting:         10,
	}
	card2 := &ProjectIssue{
		IssueID:         15,
		ProjectID:       1,
		ProjectColumnID: column.ID,
		Sorting:         11,
	}
	_, err := db.GetEngine(db.DefaultContext).Insert(card1)
	require.NoError(t, err)
	_, err = db.GetEngine(db.DefaultContext).Insert(card2)
	require.NoError(t, err)

	// Swap them: card at 10→11, card at 11→10
	sortedIssueIDs := map[int64]int64{
		11: 14, // issue 14 goes to position 11
		10: 15, // issue 15 goes to position 10
	}
	err = MoveIssuesOnProjectColumn(db.DefaultContext, column, sortedIssueIDs)
	require.NoError(t, err)

	var resultCard14 ProjectIssue
	has, err := db.GetEngine(db.DefaultContext).Where("project_id=? AND issue_id=?", 1, 14).Get(&resultCard14)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, int64(11), resultCard14.Sorting)

	var resultCard15 ProjectIssue
	has, err = db.GetEngine(db.DefaultContext).Where("project_id=? AND issue_id=?", 1, 15).Get(&resultCard15)
	require.NoError(t, err)
	require.True(t, has)
	assert.Equal(t, int64(10), resultCard15.Sorting)
}

func TestMoveIssuesOnProjectColumnEmptyMap(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	column := unittest.AssertExistsAndLoadBean(t, &Column{ID: 1})
	err := MoveIssuesOnProjectColumn(db.DefaultContext, column, map[int64]int64{})
	require.NoError(t, err) // empty map should be a no-op
}

func TestMoveIssuesOnProjectColumnDuplicateIssueIDs(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	column := unittest.AssertExistsAndLoadBean(t, &Column{ID: 1})
	err := MoveIssuesOnProjectColumn(db.DefaultContext, column, map[int64]int64{
		0: 1,
		1: 1, // duplicate issue ID
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate issue IDs")
}

func TestMoveIssuesToAnotherColumnErrorPaths(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("DifferentProject", func(t *testing.T) {
		col1 := unittest.AssertExistsAndLoadBean(t, &Column{ID: 1, ProjectID: 1})
		col5 := unittest.AssertExistsAndLoadBean(t, &Column{ID: 5, ProjectID: 2})

		err := col1.moveIssuesToAnotherColumn(db.DefaultContext, col5)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "columns have to be in the same project")
	})

	t.Run("SameColumnIsNoOp", func(t *testing.T) {
		col1 := unittest.AssertExistsAndLoadBean(t, &Column{ID: 1, ProjectID: 1})

		err := col1.moveIssuesToAnotherColumn(db.DefaultContext, col1)
		require.NoError(t, err)
	})
}
