// Copyright 2017 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swagger

import (
	api "forgejo.org/modules/structs"
)

// Issue
// swagger:response Issue
type swaggerResponseIssue struct {
	// in:body
	Body api.Issue `json:"body"`
}

// IssueList
// swagger:response IssueList
type swaggerResponseIssueList struct {
	// in:body
	Body []api.Issue `json:"body"`
}

// IssueListWithoutPagination - Issues without pagination headers (used for pinned issues, dependencies, etc.)
// swagger:response IssueListWithoutPagination
type swaggerIssueListWithoutPagination struct {
	// in:body
	Body []api.Issue `json:"body"`
}

// Comment
// swagger:response Comment
type swaggerResponseComment struct {
	// in:body
	Body api.Comment `json:"body"`
}

// CommentList
// swagger:response CommentList
type swaggerResponseCommentList struct {
	// in:body
	Body []api.Comment `json:"body"`

	// The total number of comments
	TotalCount int64 `json:"X-Total-Count"`
}

// CommentListWithoutPagination - Comments without pagination headers
// swagger:response CommentListWithoutPagination
type swaggerCommentListWithoutPagination struct {
	// in:body
	Body []api.Comment `json:"body"`
}

// TimelineList
// swagger:response TimelineList
type swaggerResponseTimelineList struct {
	// in:body
	Body []api.TimelineComment `json:"body"`

	// The total number of timeline comments
	TotalCount int64 `json:"X-Total-Count"`
}

// Label
// swagger:response Label
type swaggerResponseLabel struct {
	// in:body
	Body api.Label `json:"body"`
}

// LabelList
// swagger:response LabelList
type swaggerResponseLabelList struct {
	// in:body
	Body []api.Label `json:"body"`

	// The total number of labels
	TotalCount int64 `json:"X-Total-Count"`
}

// LabelListWithoutPagination - Labels for a specific issue (no pagination headers)
// swagger:response LabelListWithoutPagination
type swaggerLabelListWithoutPagination struct {
	// in:body
	Body []api.Label `json:"body"`
}

// Milestone
// swagger:response Milestone
type swaggerResponseMilestone struct {
	// in:body
	Body api.Milestone `json:"body"`
}

// MilestoneList
// swagger:response MilestoneList
type swaggerResponseMilestoneList struct {
	// in:body
	Body []api.Milestone `json:"body"`

	// The total number of milestones
	TotalCount int64 `json:"X-Total-Count"`
}

// TrackedTime
// swagger:response TrackedTime
type swaggerResponseTrackedTime struct {
	// in:body
	Body api.TrackedTime `json:"body"`
}

// TrackedTimeList
// swagger:response TrackedTimeList
type swaggerResponseTrackedTimeList struct {
	// in:body
	Body []api.TrackedTime `json:"body"`

	// The total number of tracked times
	TotalCount int64 `json:"X-Total-Count"`
}

// TrackedTimeListWithoutPagination - Tracked times for a specific user (no pagination headers)
// swagger:response TrackedTimeListWithoutPagination
type swaggerTrackedTimeListWithoutPagination struct {
	// in:body
	Body []api.TrackedTime `json:"body"`
}

// IssueDeadline
// swagger:response IssueDeadline
type swaggerIssueDeadline struct {
	// in:body
	Body api.IssueDeadline `json:"body"`
}

// IssueTemplates
// swagger:response IssueTemplates
type swaggerIssueTemplates struct {
	// in:body
	Body []api.IssueTemplate `json:"body"`
}

// StopWatch
// swagger:response StopWatch
type swaggerResponseStopWatch struct {
	// in:body
	Body api.StopWatch `json:"body"`
}

// StopWatchList
// swagger:response StopWatchList
type swaggerResponseStopWatchList struct {
	// in:body
	Body []api.StopWatch `json:"body"`

	// The total number of stop watches
	TotalCount int64 `json:"X-Total-Count"`
}

// Reaction
// swagger:response Reaction
type swaggerReaction struct {
	// in:body
	Body api.Reaction `json:"body"`
}

// ReactionList
// swagger:response ReactionList
type swaggerReactionList struct {
	// in:body
	Body []api.Reaction `json:"body"`

	// The total number of reactions
	TotalCount int64 `json:"X-Total-Count"`
}

// ReactionListWithoutPagination - Reactions for a specific comment (no pagination headers)
// swagger:response ReactionListWithoutPagination
type swaggerReactionListWithoutPagination struct {
	// in:body
	Body []api.Reaction `json:"body"`
}
