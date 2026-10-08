// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo_test

import (
	"testing"
	"time"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushMirrorsIterate(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	now := timeutil.TimeStampNow()

	db.Insert(db.DefaultContext, &repo_model.PushMirror{
		RemoteName:     "test-1",
		LastUpdateUnix: now,
		Interval:       1,
	})

	long, _ := time.ParseDuration("24h")
	db.Insert(db.DefaultContext, &repo_model.PushMirror{
		RemoteName:     "test-2",
		LastUpdateUnix: now,
		Interval:       long,
	})

	db.Insert(db.DefaultContext, &repo_model.PushMirror{
		RemoteName:     "test-3",
		LastUpdateUnix: now,
		Interval:       0,
	})

	repo_model.PushMirrorsIterate(db.DefaultContext, 1, func(idx int, bean any) error {
		m, ok := bean.(*repo_model.PushMirror)
		assert.True(t, ok)
		assert.Equal(t, "test-1", m.RemoteName)
		assert.Equal(t, m.RemoteName, m.GetRemoteName())
		return nil
	})
}

func TestPushMirrorPrivatekey(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	m := &repo_model.PushMirror{
		RemoteName: "test-privatekey",
	}
	require.NoError(t, db.Insert(db.DefaultContext, m))

	privateKey := []byte{0x00, 0x01, 0x02, 0x04, 0x08, 0x10}
	t.Run("Set privatekey", func(t *testing.T) {
		require.NoError(t, m.SetPrivatekey(db.DefaultContext, privateKey))
	})

	t.Run("Normal retrieval", func(t *testing.T) {
		actualPrivateKey, err := m.Privatekey()
		require.NoError(t, err)
		assert.Equal(t, privateKey, actualPrivateKey)
	})

	t.Run("Incorrect retrieval", func(t *testing.T) {
		m.ID++
		actualPrivateKey, err := m.Privatekey()
		require.Error(t, err)
		assert.Empty(t, actualPrivateKey)
	})
}

func TestPushMirrorBranchFilter(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("Create push mirror with branch filter", func(t *testing.T) {
		m := &repo_model.PushMirror{
			RepoID:       1,
			RemoteName:   "test-branch-filter",
			BranchFilter: "main,develop",
		}
		unittest.AssertSuccessfulInsert(t, m)
		assert.NotZero(t, m.ID)
		assert.Equal(t, "main,develop", m.BranchFilter)
	})

	t.Run("Create push mirror with empty branch filter", func(t *testing.T) {
		m := &repo_model.PushMirror{
			RepoID:       1,
			RemoteName:   "test-empty-filter",
			BranchFilter: "",
		}
		unittest.AssertSuccessfulInsert(t, m)
		assert.NotZero(t, m.ID)
		assert.Empty(t, m.BranchFilter)
	})

	t.Run("Create push mirror without branch filter", func(t *testing.T) {
		m := &repo_model.PushMirror{
			RepoID:     1,
			RemoteName: "test-no-filter",
			// BranchFilter: "",
		}
		unittest.AssertSuccessfulInsert(t, m)
		assert.NotZero(t, m.ID)
		assert.Empty(t, m.BranchFilter)
	})

	t.Run("Update branch filter", func(t *testing.T) {
		m := &repo_model.PushMirror{
			RepoID:       1,
			RemoteName:   "test-update",
			BranchFilter: "main",
		}
		unittest.AssertSuccessfulInsert(t, m)

		m.BranchFilter = "main,develop"
		require.NoError(t, repo_model.UpdatePushMirrorBranchFilter(db.DefaultContext, m))

		updated := unittest.AssertExistsAndLoadBean(t, &repo_model.PushMirror{ID: m.ID})
		assert.Equal(t, "main,develop", updated.BranchFilter)
	})

	t.Run("Retrieve push mirror with branch filter", func(t *testing.T) {
		original := &repo_model.PushMirror{
			RepoID:       1,
			RemoteName:   "test-retrieve",
			BranchFilter: "main,develop",
		}
		unittest.AssertSuccessfulInsert(t, original)

		retrieved := unittest.AssertExistsAndLoadBean(t, &repo_model.PushMirror{ID: original.ID})
		assert.Equal(t, original.BranchFilter, retrieved.BranchFilter)
		assert.Equal(t, "main,develop", retrieved.BranchFilter)
	})

	t.Run("GetPushMirrorsByRepoID includes branch filter", func(t *testing.T) {
		mirrors := []*repo_model.PushMirror{
			{
				RepoID:       2,
				RemoteName:   "mirror-1",
				BranchFilter: "main",
			},
			{
				RepoID:       2,
				RemoteName:   "mirror-2",
				BranchFilter: "develop,feature-*",
			},
			{
				RepoID:       2,
				RemoteName:   "mirror-3",
				BranchFilter: "",
			},
		}

		for _, mirror := range mirrors {
			unittest.AssertSuccessfulInsert(t, mirror)
		}

		retrieved, count, err := repo_model.GetPushMirrorsByRepoID(db.DefaultContext, 2, db.ListOptions{})
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
		assert.Len(t, retrieved, 3)

		filterMap := make(map[string]string)
		for _, mirror := range retrieved {
			filterMap[mirror.RemoteName] = mirror.BranchFilter
		}

		assert.Equal(t, "main", filterMap["mirror-1"])
		assert.Equal(t, "develop,feature-*", filterMap["mirror-2"])
		assert.Empty(t, filterMap["mirror-3"])
	})

	t.Run("GetPushMirrorsSyncedOnCommit includes branch filter", func(t *testing.T) {
		mirrors := []*repo_model.PushMirror{
			{
				RepoID:       3,
				RemoteName:   "sync-mirror-1",
				BranchFilter: "main,develop",
				SyncOnCommit: true,
			},
			{
				RepoID:       3,
				RemoteName:   "sync-mirror-2",
				BranchFilter: "feature-*",
				SyncOnCommit: true,
			},
		}

		for _, mirror := range mirrors {
			unittest.AssertSuccessfulInsert(t, mirror)
		}

		retrieved, err := repo_model.GetPushMirrorsSyncedOnCommit(db.DefaultContext, 3)
		require.NoError(t, err)
		assert.Len(t, retrieved, 2)

		filterMap := make(map[string]string)
		for _, mirror := range retrieved {
			filterMap[mirror.RemoteName] = mirror.BranchFilter
		}

		assert.Equal(t, "main,develop", filterMap["sync-mirror-1"])
		assert.Equal(t, "feature-*", filterMap["sync-mirror-2"])
	})
}
