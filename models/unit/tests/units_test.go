// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package tests

import (
	"testing"

	unit_model "forgejo.org/models/unit"
	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/assert"
)

func TestSaveUnits(t *testing.T) {
	restoreUnits := SaveUnits()

	unit_model.DisabledRepoUnitsSet([]unit_model.Type{unit_model.TypeInvalid})
	setting.Repository.DisabledRepoUnits = []string{"invalid"}

	unit_model.DefaultRepoUnits = []unit_model.Type{unit_model.TypeInvalid}
	setting.Repository.DefaultRepoUnits = []string{"invalid"}

	unit_model.DefaultForkRepoUnits = []unit_model.Type{unit_model.TypeInvalid}
	setting.Repository.DefaultForkRepoUnits = []string{"invalid"}

	restoreUnits()

	assert.NotEqual(t, []unit_model.Type{unit_model.TypeInvalid}, unit_model.DisabledRepoUnitsGet())
	assert.NotEqual(t, []string{"invalid"}, setting.Repository.DisabledRepoUnits)

	assert.NotEqual(t, []unit_model.Type{unit_model.TypeInvalid}, unit_model.DefaultRepoUnits)
	assert.NotEqual(t, []string{"invalid"}, setting.Repository.DefaultRepoUnits)

	assert.NotEqual(t, []unit_model.Type{unit_model.TypeInvalid}, unit_model.DefaultForkRepoUnits)
	assert.NotEqual(t, []string{"invalid"}, setting.Repository.DefaultForkRepoUnits)
}
