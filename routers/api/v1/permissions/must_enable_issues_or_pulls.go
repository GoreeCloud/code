// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"forgejo.org/models/unit"
	"forgejo.org/modules/log"
)

func MustEnableIssuesOrPulls(ctx Context) {
	if !ctx.Permission().CanRead(unit.TypeIssues) &&
		(!ctx.Repository().CanEnablePulls() || !ctx.Permission().CanRead(unit.TypePullRequests)) {
		if ctx.Repository().CanEnablePulls() && log.IsTrace() {
			if ctx.IsSigned() {
				log.Trace("Permission Denied: User %-v cannot read %-v and %-v in Repo %-v\n"+
					"User in Repo has Permissions: %-+v",
					ctx.Doer(),
					unit.TypeIssues,
					unit.TypePullRequests,
					ctx.Repository(),
					ctx.Permission())
			} else {
				log.Trace("Permission Denied: Anonymous user cannot read %-v and %-v in Repo %-v\n"+
					"Anonymous user in Repo has Permissions: %-+v",
					unit.TypeIssues,
					unit.TypePullRequests,
					ctx.Repository(),
					ctx.Permission())
			}
		}
		ctx.NotFound()
	}
}
