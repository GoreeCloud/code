// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import (
	"time"
)

// ActionTask represents a ActionTask
type ActionTask struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	HeadBranch   string `json:"head_branch"`
	HeadSHA      string `json:"head_sha"`
	RunNumber    int64  `json:"run_number"`
	Event        string `json:"event"`
	DisplayTitle string `json:"display_title"`
	Status       string `json:"status"`
	WorkflowID   string `json:"workflow_id"`
	URL          string `json:"url"`
	// swagger:strfmt date-time
	CreatedAt time.Time `json:"created_at"`
	// swagger:strfmt date-time
	UpdatedAt time.Time `json:"updated_at"`
	// swagger:strfmt date-time
	RunStartedAt time.Time `json:"run_started_at"`
}

// ActionTaskResponse returns a ActionTask
type ActionTaskResponse struct {
	Entries    []*ActionTask `json:"workflow_runs"`
	TotalCount int64         `json:"total_count"`
}

type RunnerStatus int

const (
	// RunnerStatusOffline signals that the runner is not connected to Forgejo.
	RunnerStatusOffline RunnerStatus = iota

	// RunnerStatusIdle means that the runner is connected to Forgejo and waiting for jobs to run.
	RunnerStatusIdle

	// RunnerStatusActive signifies that the runner is connected to Forgejo and running a job.
	RunnerStatusActive
)

var statusName = map[RunnerStatus]string{
	RunnerStatusOffline: "offline",
	RunnerStatusIdle:    "idle",
	RunnerStatusActive:  "active",
}

func (status RunnerStatus) String() string {
	return statusName[status]
}

// ActionRunner represents a runner
// swagger:model
type ActionRunner struct {
	// ID uniquely identifies this runner.
	ID int64 `json:"id"`
	// UUID uniquely identifies this runner.
	UUID string `json:"uuid"`
	// OwnerID is the identifier of the user or organization this runner belongs to. O if the runner is owned by a
	// repository.
	OwnerID int64 `json:"owner_id"`
	// RepoID is the identifier of the repository this runner belongs to. 0 if the runner belongs to a user or
	// organization.
	RepoID int64 `json:"repo_id"`
	// Name of the runner; not unique.
	Name string `json:"name"`
	// Status indicates whether this runner is offline, or active, for example.
	// enum: ["offline", "idle", "active"]
	Status string `json:"status"`
	// Version is the self-reported version string of Forgejo Runner.
	Version string `json:"version"`
	// Labels is a list of labels attached to this runner.
	Labels []string `json:"labels"`
	// Description provides optional details about this runner.
	Description string `json:"description"`
	// Indicates if runner is ephemeral runner
	Ephemeral bool `json:"ephemeral"`
}
