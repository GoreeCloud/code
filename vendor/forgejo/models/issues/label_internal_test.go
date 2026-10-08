// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package issues

import (
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecalcLabelByLabelID(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// Verify no error on recalc of a deleted/non-existent object; important because async recalcs can be queued and
	// then occur later after more state changes have happened.
	err := doRecalcLabelByID(t.Context(), -1000, optional.None[timeutil.TimeStamp]())
	require.NoError(t, err)

	// Intentionally corrupt counts from fixture, then recalc them
	label := unittest.AssertExistsAndLoadBean(t, &Label{ID: 1})
	updated, err := db.GetEngine(t.Context()).
		Table(&Label{}).
		Where("id = ?", label.ID).
		Update(map[string]any{"num_issues": 1000, "num_closed_issues": 1001})
	require.NoError(t, err)
	require.EqualValues(t, 1, updated)
	err = doRecalcLabelByID(t.Context(), label.ID, optional.None[timeutil.TimeStamp]())
	require.NoError(t, err)
	label = unittest.AssertExistsAndLoadBean(t, &Label{ID: 1})
	assert.Equal(t, 2, label.NumIssues)
	assert.Equal(t, 0, label.NumClosedIssues)
}

func TestRecalcLabelByRepoID(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// Verify no error on recalc of a deleted/non-existent object; important because async recalcs can be queued and
	// then occur later after more state changes have happened.
	err := doRecalcLabelByRepoID(t.Context(), -1000, optional.None[timeutil.TimeStamp]())
	require.NoError(t, err)

	// Intentionally corrupt counts from fixture, then recalc them
	label1 := unittest.AssertExistsAndLoadBean(t, &Label{ID: 1})
	updated, err := db.GetEngine(t.Context()).
		Table(&Label{}).
		Where("id = ?", label1.ID).
		Update(map[string]any{"num_issues": 1000, "num_closed_issues": 1001})
	require.NoError(t, err)
	require.EqualValues(t, 1, updated)
	label2 := unittest.AssertExistsAndLoadBean(t, &Label{ID: 2})
	require.Equal(t, label1.RepoID, label2.RepoID) // sanity check
	updated, err = db.GetEngine(t.Context()).
		Table(&Label{}).
		Where("id = ?", label2.ID).
		Update(map[string]any{"num_issues": 1000, "num_closed_issues": 1001})
	require.NoError(t, err)
	require.EqualValues(t, 1, updated)
	err = doRecalcLabelByRepoID(t.Context(), label1.RepoID, optional.None[timeutil.TimeStamp]())
	require.NoError(t, err)
	label1 = unittest.AssertExistsAndLoadBean(t, &Label{ID: 1})
	label2 = unittest.AssertExistsAndLoadBean(t, &Label{ID: 2})
	assert.Equal(t, 2, label1.NumIssues)
	assert.Equal(t, 0, label1.NumClosedIssues)
	assert.Equal(t, 1, label2.NumIssues)
	assert.Equal(t, 1, label2.NumClosedIssues)
}
