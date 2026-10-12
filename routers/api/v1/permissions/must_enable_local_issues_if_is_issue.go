// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	issues_model "forgejo.org/models/issues"
	"forgejo.org/models/unit"
)

func MustEnableLocalIssuesIfIsIssue(ctx Context, index int64) {
	if ctx.Repository().UnitEnabled(ctx.Context(), unit.TypeIssues) {
		return
	}

	issue, err := issues_model.GetIssueByIndex(ctx.Context(), ctx.Repository().ID, index)
	if err != nil {
		if issues_model.IsErrIssueNotExist(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "GetIssueByIndex", err)
		}
		return
	}
	if !issue.IsPull {
		ctx.NotFound()
		return
	}
}
