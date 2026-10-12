// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancelAbandonedJobs(t *testing.T) {
	defer unittest.OverrideFixtures("services/actions/TestCancelAbandonedJobs")()
	require.NoError(t, unittest.PrepareTestDatabase())

	require.NoError(t, CancelAbandonedJobs(t.Context()))

	// status waiting, too long, ready to be abandoned
	job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 600})
	assert.Equal(t, actions_model.StatusCancelled, job.Status)

	// status blocked, too long, ready to be abandoned
	job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 601})
	assert.Equal(t, actions_model.StatusCancelled, job.Status)

	// status blocked, *not* too long, not to be abandoned
	job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 602})
	assert.Equal(t, actions_model.StatusBlocked, job.Status)

	// status running, not to be abandoned
	job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 603})
	assert.Equal(t, actions_model.StatusRunning, job.Status)

	// related run needs approval, not to be abandoned
	job = unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 604})
	assert.Equal(t, actions_model.StatusWaiting, job.Status)
}
