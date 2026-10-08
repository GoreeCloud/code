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

func TestRecalcMilestoneByMilestoneID(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// Verify no error on recalc of a deleted/non-existent object; important because async recalcs can be queued and
	// then occur later after more state changes have happened.
	err := doRecalcMilestoneByID(t.Context(), -1000, optional.None[timeutil.TimeStamp]())
	require.NoError(t, err)

	// Intentionally corrupt counts from fixture, then recalc them
	milestone := unittest.AssertExistsAndLoadBean(t, &Milestone{ID: 1})
	updated, err := db.GetEngine(t.Context()).
		Table(&Milestone{}).
		Where("id = ?", milestone.ID).
		Update(map[string]any{
			"num_issues":        1000,
			"num_closed_issues": 1001,
			"completeness":      99,
			"updated_unix":      123,
		})
	require.NoError(t, err)
	require.EqualValues(t, 1, updated)
	err = doRecalcMilestoneByID(t.Context(), milestone.ID, optional.None[timeutil.TimeStamp]())
	require.NoError(t, err)
	milestone = unittest.AssertExistsAndLoadBean(t, &Milestone{ID: 1})
	assert.Equal(t, 1, milestone.NumIssues)
	assert.Equal(t, 0, milestone.NumClosedIssues)
	assert.Equal(t, 0, milestone.Completeness)
	assert.NotEqualValues(t, 123, milestone.UpdatedUnix)

	// Exercise the updateTimestamp option to the recalc
	updated, err = db.GetEngine(t.Context()).
		Table(&Milestone{}).
		Where("id = ?", milestone.ID).
		Update(map[string]any{
			"num_issues":        1000,
			"num_closed_issues": 1001,
			"completeness":      99,
			"updated_unix":      123,
		})
	require.NoError(t, err)
	require.EqualValues(t, 1, updated)
	err = doRecalcMilestoneByID(t.Context(), milestone.ID, optional.Some(timeutil.TimeStamp(456)))
	require.NoError(t, err)
	milestone = unittest.AssertExistsAndLoadBean(t, &Milestone{ID: 1})
	assert.Equal(t, 1, milestone.NumIssues)
	assert.Equal(t, 0, milestone.NumClosedIssues)
	assert.Equal(t, 0, milestone.Completeness)
	assert.EqualValues(t, 456, milestone.UpdatedUnix)
}
