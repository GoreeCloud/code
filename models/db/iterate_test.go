// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package db_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"forgejo.org/models/db"
	git_model "forgejo.org/models/git"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func TestIterate(t *testing.T) {
	defer test.MockVariableValue(&setting.Database.IterateBufferSize, 50)()

	t.Run("No Modifications", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())
		xe, err := unittest.GetXORMEngine()
		require.NoError(t, err)
		require.NoError(t, xe.Sync(&repo_model.RepoUnit{}))

		// Fetch all the repo unit IDs...
		var remainingRepoIDs []int64
		db.GetEngine(t.Context()).Table(&repo_model.RepoUnit{}).Cols("id").Find(&remainingRepoIDs)

		// Ensure that every repo unit ID is found when doing iterate:
		err = db.Iterate(t.Context(), nil, func(ctx context.Context, repo *repo_model.RepoUnit) error {
			remainingRepoIDs = slices.DeleteFunc(remainingRepoIDs, func(n int64) bool {
				return repo.ID == n
			})
			return nil
		})
		require.NoError(t, err)
		assert.Empty(t, remainingRepoIDs)
	})

	t.Run("Concurrent Delete", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())
		xe, err := unittest.GetXORMEngine()
		require.NoError(t, err)
		require.NoError(t, xe.Sync(&repo_model.RepoUnit{}))

		// Fetch all the repo unit IDs...
		var remainingRepoIDs []int64
		db.GetEngine(t.Context()).Table(&repo_model.RepoUnit{}).Cols("id").Find(&remainingRepoIDs)

		// Ensure that every repo unit ID is found, even if someone else performs a DELETE on the table while we're
		// iterating.  In real-world usage the deleted record may or may not be returned, but the important
		// subject-under-test is that no *other* record is skipped.
		didDelete := false
		err = db.Iterate(t.Context(), nil, func(ctx context.Context, repo *repo_model.RepoUnit) error {
			// While on page 2 (assuming ID ordering, 50 record buffer size)...
			if repo.ID == 51 {
				// Delete a record that would have been on page 1.
				affected, err := db.GetEngine(t.Context()).ID(25).Delete(&repo_model.RepoUnit{})
				if err != nil {
					return err
				} else if affected != 1 {
					return fmt.Errorf("expected to delete 1 record, but affected %d records", affected)
				}
				didDelete = true
			}
			remainingRepoIDs = slices.DeleteFunc(remainingRepoIDs, func(n int64) bool {
				return repo.ID == n
			})
			return nil
		})
		require.NoError(t, err)
		assert.True(t, didDelete, "didDelete")
		assert.Empty(t, remainingRepoIDs)
	})

	t.Run("Verify cond applied", func(t *testing.T) {
		require.NoError(t, unittest.PrepareTestDatabase())
		xe, err := unittest.GetXORMEngine()
		require.NoError(t, err)
		require.NoError(t, xe.Sync(&repo_model.RepoUnit{}))

		// Fetch all the repo unit IDs...
		var remainingRepoIDs []int64
		db.GetEngine(t.Context()).Table(&repo_model.RepoUnit{}).Cols("id").Find(&remainingRepoIDs)

		// Remove those that we're not expecting to find based upon `Iterate`'s condition.  We'll trim the front few
		// records and last few records, which will confirm that cond is applied on all pages.
		remainingRepoIDs = slices.DeleteFunc(remainingRepoIDs, func(n int64) bool {
			return n <= 15 || n > 1000
		})
		err = db.Iterate(t.Context(), builder.Gt{"id": 15}.And(builder.Lt{"id": 1000}), func(ctx context.Context, repo *repo_model.RepoUnit) error {
			removedRecord := false
			// Remove the record from remainingRepoIDs, but track to make sure we did actually remove a record
			remainingRepoIDs = slices.DeleteFunc(remainingRepoIDs, func(n int64) bool {
				if repo.ID == n {
					removedRecord = true
					return true
				}
				return false
			})
			if !removedRecord {
				return fmt.Errorf("unable to find record in remainingRepoIDs for repo %d, indicating a cond application failure", repo.ID)
			}
			return nil
		})
		require.NoError(t, err)
		assert.Empty(t, remainingRepoIDs)
	})
}

func TestIterateMultipleFields(t *testing.T) {
	for _, bufferSize := range []int{1, 2, 3, 10} { // 8 records in fixture
		t.Run(fmt.Sprintf("No Modifications bufferSize=%d", bufferSize), func(t *testing.T) {
			require.NoError(t, unittest.PrepareTestDatabase())

			// Fetch all the commit status IDs...
			var remainingIDs []int64
			err := db.GetEngine(t.Context()).Table(&git_model.CommitStatus{}).Cols("id").Find(&remainingIDs)
			require.NoError(t, err)
			require.NotEmpty(t, remainingIDs)

			// Ensure that every repo unit ID is found when doing iterate:
			err = db.IterateByKeyset(t.Context(),
				nil,
				[]string{"repo_id", "sha", "context", "index", "id"},
				bufferSize,
				func(ctx context.Context, commit_status *git_model.CommitStatus) error {
					remainingIDs = slices.DeleteFunc(remainingIDs, func(n int64) bool {
						return commit_status.ID == n
					})
					return nil
				})
			require.NoError(t, err)
			assert.Empty(t, remainingIDs)
		})
	}
}
