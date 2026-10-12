// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"forgejo.org/modules/container"
)

const (
	// MaxRunPriority is the highest possible priority of an ActionRun.
	MaxRunPriority int8 = 127
	// DefaultRunPriority is the default priority assigned to ActionRun instances.
	DefaultRunPriority int8 = 0
	// MinRunPriority is the lowest possible priority of an ActionRun.
	MinRunPriority int8 = -128
)

type RunPrioritizationStrategy interface {
	// PrioritizeRuns updates the priority of all ActionRun instances passed as argument. It returns a set containing
	// the IDs of all ActionRun instances whose priority was changed, or an error.
	//
	// It is the responsibility of each implementation to handle the ActionRun's Prioritized field appropriately.
	// Ignoring it is explicitly allowed.
	//
	// Forgejo sorts jobs by the ActionRun's priority followed by the time they were last updated and their ID, which
	// results in FIFO order. That behaviour cannot be influenced by implementations. It also means that they only have
	// to change an ActionRun's priority if FIFO order is not desired.
	//
	// PrioritizeRuns participates in an ongoing transaction. Implementations are free to query the database, but should
	// refrain from writing to it. Changes to any other aspect of the ActionRun besides its priority are discarded.
	PrioritizeRuns(runs []*ActionRun) (container.Set[int64], error)
}

var _ RunPrioritizationStrategy = DefaultPrioritizationStrategy{}

// DefaultPrioritizationStrategy boosts the priority of manually prioritized jobs, but retains the default order
// otherwise.
type DefaultPrioritizationStrategy struct{}

func (s DefaultPrioritizationStrategy) PrioritizeRuns(runs []*ActionRun) (container.Set[int64], error) {
	changedRuns := container.SetOf[int64]()
	for _, run := range runs {
		oldPriority := run.Priority
		if run.Prioritize {
			run.Priority = MaxRunPriority
		} else {
			run.Priority = DefaultRunPriority
		}
		if run.Priority != oldPriority {
			changedRuns.Add(run.ID)
		}
	}

	return changedRuns, nil
}
