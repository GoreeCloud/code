// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	issues_model "forgejo.org/models/issues"
)

func ReqValidCommentID(ctx Context, comment *issues_model.Comment) {
	if comment.Issue == nil || comment.Issue.RepoID != ctx.Repository().ID {
		ctx.NotFound()
		return
	}

	if !ctx.Permission().CanReadIssuesOrPulls(comment.Issue.IsPull) {
		ctx.NotFound()
		return
	}
}
