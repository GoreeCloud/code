// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	actions_model "forgejo.org/models/actions"
	git_model "forgejo.org/models/git"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/cache"
	"forgejo.org/modules/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestCreateCommitStatus_IncompleteMatrix(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 192})

	// Normally this job will attempt to create a commit status on a commit that doesn't exist in the test repo,
	// resulting in an error due to the test fixture data not matching the related repos. But it tried.
	err := createCommitStatus(t.Context(), job)
	require.ErrorContains(t, err, "object does not exist [id: 7a3858dc7f059543a8807a8b551304b7e362a7ef")

	// Transition from HasIncompleteMatrix()=false to true...
	isIncomplete, _, err := job.HasIncompleteMatrix()
	require.NoError(t, err)
	require.False(t, isIncomplete)
	job.WorkflowPayload = append(job.WorkflowPayload, "\nincomplete_matrix: true\n"...)
	job.ClearCachedWorkflowPayload()
	isIncomplete, _, err = job.HasIncompleteMatrix()
	require.NoError(t, err)
	require.True(t, isIncomplete)

	// Now there should be no error since createCommitStatus will exit early due to the HasIncompleteMatrix() flag.
	err = createCommitStatus(t.Context(), job)
	require.NoError(t, err)
}

func TestCreateCommitStatus_AvoidsDuplicates(t *testing.T) {
	defer unittest.OverrideFixtures("services/actions/TestCreateCommitStatus")()
	require.NoError(t, unittest.PrepareTestDatabase())
	cache.Init()

	job := unittest.AssertExistsAndLoadBean(t, &actions_model.ActionRunJob{ID: 400})
	targetCommitStatusFilter := builder.Eq{"repo_id": 4, "sha": job.CommitSHA}

	// Begin with 0 commit statuses
	assert.Equal(t, 0, unittest.GetCount(t, &git_model.CommitStatus{}, targetCommitStatusFilter))

	err := createCommitStatus(t.Context(), job)
	require.NoError(t, err)

	// Should have 1 commit status now with one createCommitStatus call
	assert.Equal(t, 1, unittest.GetCount(t, &git_model.CommitStatus{}, targetCommitStatusFilter))
	status := unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{}, targetCommitStatusFilter)
	assert.EqualValues(t, 4, status.RepoID)
	assert.Equal(t, structs.CommitStatusState("pending"), status.State)
	assert.Equal(t, "c7cd3cd144e6d23c9d6f3d07e52b2c1a956e0338", status.SHA)
	assert.Equal(t, "/user5/repo4/actions/runs/200/jobs/0", status.TargetURL)
	assert.Equal(t, "Blocked by required conditions", status.Description)
	assert.Equal(t, "39c46bc9f0f68e0dcfbb59118e12778fa095b066", status.ContextHash)
	assert.Equal(t, "/ job_2 (push)", status.Context) // This test is intended to cover the runName = "" case, which trims whitespace in this context string -- don't change it in the future
	assert.EqualValues(t, 1, status.Index)

	// No status change, but createCommitStatus invoked again
	err = createCommitStatus(t.Context(), job)
	require.NoError(t, err)

	// Should have just the same 1 commit status since the context & state was unchanged.
	assert.Equal(t, 1, unittest.GetCount(t, &git_model.CommitStatus{}, targetCommitStatusFilter))

	// Change status, but still pending -- should add new commit status just for the Description change
	job.Status = actions_model.StatusWaiting // Blocked -> Waiting
	err = createCommitStatus(t.Context(), job)
	require.NoError(t, err)

	// New commit status added
	assert.Equal(t, 2, unittest.GetCount(t, &git_model.CommitStatus{}, targetCommitStatusFilter))
	status = unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{Index: 2}, targetCommitStatusFilter)
	assert.Equal(t, structs.CommitStatusState("pending"), status.State)
	assert.Equal(t, "Waiting to run", status.Description)

	// Invoke createCommitStatus again, check no new record created again
	err = createCommitStatus(t.Context(), job)
	require.NoError(t, err)
	assert.Equal(t, 2, unittest.GetCount(t, &git_model.CommitStatus{}, targetCommitStatusFilter))

	// Update job status & create new commit status
	job.Status = actions_model.StatusSuccess
	err = createCommitStatus(t.Context(), job)
	require.NoError(t, err)

	// New commit status added w/ updated state & description
	assert.Equal(t, 3, unittest.GetCount(t, &git_model.CommitStatus{}, targetCommitStatusFilter))
	status = unittest.AssertExistsAndLoadBean(t, &git_model.CommitStatus{Index: 3}, targetCommitStatusFilter)
	assert.Equal(t, structs.CommitStatusState("success"), status.State)
	assert.Equal(t, "Successful in 1m38s", status.Description)
}
