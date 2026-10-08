// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"
	"time"

	"forgejo.org/models/db"
	"forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/models/user"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/webhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleCreateScheduleTask(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	user2 := unittest.AssertExistsAndLoadBean(t, &user.User{ID: 2})
	repo62 := unittest.AssertExistsAndLoadBean(t, &repo.Repository{ID: 62, Name: "test_workflows", OwnerID: user2.ID})

	content := `
on:
  push:
  schedule:
    - cron: "2 13 * * *"
    - cron: "03 13 * * *"
      timezone: Europe/Paris
jobs:
  test:
    runs-on: debian
    steps:
      - run: |
          echo "OK"
`

	referenceTime := time.Date(2026, 3, 27, 17, 41, 21, 0, time.UTC)

	specWithoutTZ, err := NewActionScheduleSpec("2 13 * * *", optional.None[string](), referenceTime)
	require.NoError(t, err)

	specWithTZ, err := NewActionScheduleSpec("3 13 * * *", optional.Some("Europe/Paris"), referenceTime)
	require.NoError(t, err)

	schedule := &ActionSchedule{
		Title:             ".forgejo/workflows/test.yaml",
		Specs:             []*ActionScheduleSpec{specWithoutTZ, specWithTZ},
		RepoID:            repo62.ID,
		OwnerID:           user2.ID,
		WorkflowID:        "test.yaml",
		WorkflowDirectory: ".forgejo/workflows",
		TriggerUserID:     -2,
		Ref:               "main",
		CommitSHA:         "6af834a5bc97c1a337eb3a21d26903c5cdceca0c",
		Event:             webhook.HookEventPush,
		EventPayload:      "{\"action\":\"schedule\"}",
		Content:           []byte(content),
	}

	err = CreateScheduleTask(t.Context(), []*ActionSchedule{schedule})
	require.NoError(t, err)

	schedules, err := db.Find[ActionSchedule](t.Context(), FindScheduleOptions{OwnerID: user2.ID, RepoID: repo62.ID})
	require.NoError(t, err)
	require.Len(t, schedules, 1)

	assert.NotZero(t, schedules[0].ID)
	assert.Equal(t, ".forgejo/workflows/test.yaml", schedules[0].Title)
	assert.Equal(t, "test.yaml", schedules[0].WorkflowID)
	assert.Equal(t, ".forgejo/workflows", schedules[0].WorkflowDirectory)
	assert.Equal(t, int64(-2), schedules[0].TriggerUserID)
	assert.Equal(t, "main", schedules[0].Ref)
	assert.Equal(t, "6af834a5bc97c1a337eb3a21d26903c5cdceca0c", schedules[0].CommitSHA)
	assert.Equal(t, webhook.HookEventPush, schedules[0].Event)
	assert.JSONEq(t, "{\"action\":\"schedule\"}", schedules[0].EventPayload)
	assert.Equal(t, []byte(content), schedules[0].Content)

	specs, total, err := FindSpecs(t.Context(), FindSpecOptions{RepoID: repo62.ID})
	require.NoError(t, err)

	assert.Equal(t, int64(2), total)

	assert.NotZero(t, specs[0].ID)
	assert.Equal(t, schedules[0].ID, specs[0].ScheduleID)
	assert.Equal(t, timeutil.TimeStamp(1774699380), specs[0].Next)
	assert.Equal(t, "3 13 * * *", specs[0].Spec)
	assert.Equal(t, optional.Some("Europe/Paris"), specs[0].TimeZone)
	assert.Zero(t, specs[0].Prev)

	assert.NotZero(t, specs[1].ID)
	assert.Equal(t, schedules[0].ID, specs[1].ScheduleID)
	assert.Equal(t, timeutil.TimeStamp(1774702920), specs[1].Next)
	assert.Equal(t, "2 13 * * *", specs[1].Spec)
	assert.Equal(t, optional.None[string](), specs[1].TimeZone)
	assert.Zero(t, specs[1].Prev)
}
