// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPrioritizationStrategy(t *testing.T) {
	runs := []*ActionRun{
		{ID: 2, Priority: 89, Prioritize: true},
		{ID: 1, Priority: MinRunPriority},
		{ID: 5, Priority: DefaultRunPriority},
		{ID: 3, Priority: MaxRunPriority, Prioritize: true},
	}

	strategy := DefaultPrioritizationStrategy{}
	changedRuns, err := strategy.PrioritizeRuns(runs)
	require.NoError(t, err)

	assert.Len(t, changedRuns, 2)
	assert.Contains(t, changedRuns, int64(1))
	assert.Contains(t, changedRuns, int64(2))

	assert.Len(t, runs, 4)
	assert.Contains(t, runs, &ActionRun{ID: 1, Priority: DefaultRunPriority, Prioritize: false})
	assert.Contains(t, runs, &ActionRun{ID: 2, Priority: MaxRunPriority, Prioritize: true})
	assert.Contains(t, runs, &ActionRun{ID: 3, Priority: MaxRunPriority, Prioritize: true})
	assert.Contains(t, runs, &ActionRun{ID: 5, Priority: DefaultRunPriority, Prioritize: false})
}
