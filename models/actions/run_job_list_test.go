// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	"forgejo.org/modules/container"

	"github.com/stretchr/testify/assert"
)

func TestActionJobList_GetJobIDs(t *testing.T) {
	jobs := ActionJobList{
		&ActionRunJob{JobID: "job 1"},
		&ActionRunJob{JobID: "job 2"},
	}

	assert.Equal(t, container.SetOf("job 2", "job 1"), jobs.GetJobIDs())
}
