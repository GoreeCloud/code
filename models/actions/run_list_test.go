// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/translation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionStatusList(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	translation.InitLocales(t.Context())

	statusInfoList := GetStatusInfoList(t.Context(), translation.NewLocale("en-US"))
	assert.Len(t, statusInfoList, 7)
	statuses := []string{"Blocked", "Canceled", "Failure", "Running", "Skipped", "Success", "Waiting"}
	statusInts := []int{7, 3, 2, 6, 4, 1, 5}
	for i, statusString := range statuses {
		assert.Equal(t, statusInfoList[i].Status, statusInts[i])
		assert.Equal(t, statusInfoList[i].DisplayedStatus, statusString)
	}

	statusInfoList = GetStatusInfoList(t.Context(), translation.NewLocale("de-DE"))
	assert.Len(t, statusInfoList, 7)
	statuses = []string{"Abgebrochen", "Blockiert", "Erfolg", "Fehler", "Laufend", "Übersprungen", "Wartend"}
	statusInts = []int{3, 7, 1, 2, 6, 4, 5}
	for i, statusString := range statuses {
		assert.Equal(t, statusInfoList[i].Status, statusInts[i])
		assert.Equal(t, statusInfoList[i].DisplayedStatus, statusString)
	}
}
