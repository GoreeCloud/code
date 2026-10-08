// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"
	"time"

	actions_model "forgejo.org/models/actions"
	dbfs_model "forgejo.org/models/dbfs"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/test"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestServicesActions_transferLingeringLogs(t *testing.T) {
	// it would be easier to dynamically create fixtures instead of injecting them
	// in the database for testing, but the dbfs API does not have what is needed to
	// create them
	defer unittest.OverrideFixtures("services/actions/TestServicesActions_TransferLingeringLogs")()
	require.NoError(t, unittest.PrepareTestDatabase())
	defer test.MockVariableValue(&transferLingeringLogsMax, 2)()
	defer test.MockVariableValue(&transferLingeringLogsOld, 2*24*time.Hour)()
	defer test.MockVariableValue(&transferLingeringLogsSleep, time.Millisecond)()

	now, err := time.Parse("2006-01-02", "2024-12-01")
	require.NoError(t, err)
	old := timeutil.TimeStamp(now.Add(-transferLingeringLogsOld).Unix())

	// a task has a lingering log but was updated more recently than
	// transferLingeringLogsOld
	recentID := int64(2000)
	recent := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: recentID}, builder.Eq{"log_in_storage": false})
	require.Greater(t, recent.Updated, old)

	// a task has logs already in storage but would be garbage collected if it was not
	inStorageID := int64(3000)
	inStorage := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: inStorageID}, builder.Eq{"log_in_storage": true})
	require.Greater(t, old, inStorage.Updated)

	taskWithLingeringLogIDs := []int64{1000, 4000, 5000}
	for _, taskWithLingeringLogID := range taskWithLingeringLogIDs {
		lingeringLog := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionTask{ID: taskWithLingeringLogID}, builder.Eq{"log_in_storage": false})
		require.Greater(t, old, lingeringLog.Updated)
	}
	lingeringLogIDs := []int64{1, 4, 5}

	assert.True(t, unittest.BeanExists(t, &dbfs_model.DbfsMeta{}, builder.In("id", []any{lingeringLogIDs}...)))

	// first pass transfer logs for transferLingeringLogsMax tasks
	require.NoError(t, transferLingeringLogs(t.Context(), transferLingeringLogsOpts(now)))
	assert.True(t, unittest.BeanExists(t, &dbfs_model.DbfsMeta{}, builder.In("id", []any{lingeringLogIDs[transferLingeringLogsMax:]}...)))
	for _, lingeringLogID := range lingeringLogIDs[:transferLingeringLogsMax] {
		unittest.AssertNotExistsBean(t, &dbfs_model.DbfsMeta{ID: lingeringLogID})
	}

	// second pass transfer logs for the remainder tasks and there are none left
	require.NoError(t, transferLingeringLogs(t.Context(), transferLingeringLogsOpts(now)))
	for _, lingeringLogID := range lingeringLogIDs {
		unittest.AssertNotExistsBean(t, &dbfs_model.DbfsMeta{ID: lingeringLogID})
	}

	// third pass is happily doing nothing
	require.NoError(t, transferLingeringLogs(t.Context(), transferLingeringLogsOpts(now)))

	// verify the tasks that are not to be garbage collected are still present
	assert.True(t, unittest.BeanExists(t, &actions_model.ActionTask{ID: recentID}, builder.Eq{"log_in_storage": false}))
	assert.True(t, unittest.BeanExists(t, &actions_model.ActionTask{ID: inStorageID}, builder.Eq{"log_in_storage": true}))
}
