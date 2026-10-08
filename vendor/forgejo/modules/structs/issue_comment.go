// Copyright 2016 The Gogs Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import (
	"time"
)

// Comment represents a comment on a commit or issue
type Comment struct {
	// The identifier of the comment
	ID int64 `json:"id"`
	// The HTML URL of the comment
	HTMLURL string `json:"html_url"`
	// The HTML URL of the pull request if the comment is posted on a pull request, else empty string
	PRURL string `json:"pull_request_url"`
	// The HTML URL of the issue if the comment is posted on an issue, else empty string
	IssueURL string `json:"issue_url"`
	// The user that posted the comment if it was posted locally
	Poster *User `json:"user"`
	// The original author that posted the comment if it was not posted locally, else empty string
	OriginalAuthor string `json:"original_author"`
	// The ID of the original author that posted the comment if it was not posted locally, else 0
	OriginalAuthorID int64 `json:"original_author_id"`
	// The body of the comment
	Body string `json:"body"`
	// The attachments to the comment
	Attachments []*Attachment `json:"assets"`
	// The time of the comment's creation
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
	// The time of the comment's update
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`
}

// CreateIssueCommentOption options for creating a comment on an issue
type CreateIssueCommentOption struct {
	// The body of the comment
	// required:true
	Body string `json:"body" binding:"Required"`
	// The time of the comment's update, needs admin or repository owner permission
	// swagger:strfmt date-time
	Updated *time.Time `json:"updated_at"`
}

// EditIssueCommentOption options for editing a comment
type EditIssueCommentOption struct {
	// The body of the comment
	// required: true
	Body string `json:"body" binding:"Required"`
	// The time of the comment's update, needs admin or repository owner permission
	// swagger:strfmt date-time
	Updated *time.Time `json:"updated_at"`
}

// TimelineComment represents a timeline comment (comment of any type) on a commit or issue
type TimelineComment struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`

	HTMLURL  string `json:"html_url"`
	PRURL    string `json:"pull_request_url"`
	IssueURL string `json:"issue_url"`
	Poster   *User  `json:"user"`
	Body     string `json:"body"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
	// swagger:strfmt date-time
	Updated time.Time `json:"updated_at"`

	OldProjectID int64        `json:"old_project_id"`
	ProjectID    int64        `json:"project_id"`
	OldMilestone *Milestone   `json:"old_milestone"`
	Milestone    *Milestone   `json:"milestone"`
	TrackedTime  *TrackedTime `json:"tracked_time"`
	OldTitle     string       `json:"old_title"`
	NewTitle     string       `json:"new_title"`
	OldRef       string       `json:"old_ref"`
	NewRef       string       `json:"new_ref"`

	RefIssue   *Issue   `json:"ref_issue"`
	RefComment *Comment `json:"ref_comment"`
	RefAction  string   `json:"ref_action"`
	// commit SHA where issue/PR was referenced
	RefCommitSHA string `json:"ref_commit_sha"`

	ReviewID int64 `json:"review_id"`

	Label *Label `json:"label"`

	Assignee     *User `json:"assignee"`
	AssigneeTeam *Team `json:"assignee_team"`
	// whether the assignees were removed or added
	RemovedAssignee bool `json:"removed_assignee"`

	ResolveDoer *User `json:"resolve_doer"`

	DependentIssue *Issue `json:"dependent_issue"`
}
