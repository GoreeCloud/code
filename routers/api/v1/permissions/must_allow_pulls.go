// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"forgejo.org/models/unit"
	"forgejo.org/modules/log"
)

func MustAllowPulls(ctx Context) {
	if !ctx.Repository().CanEnablePulls() || !ctx.Permission().CanRead(unit.TypePullRequests) {
		if ctx.Repository().CanEnablePulls() && log.IsTrace() {
			if ctx.IsSigned() {
				log.Trace("Permission Denied: User %-v cannot read %-v in Repo %-v\n"+
					"User in Repo has Permissions: %-+v",
					ctx.Doer(),
					unit.TypePullRequests,
					ctx.Repository(),
					ctx.Permission())
			} else {
				log.Trace("Permission Denied: Anonymous user cannot read %-v in Repo %-v\n"+
					"Anonymous user in Repo has Permissions: %-+v",
					unit.TypePullRequests,
					ctx.Repository(),
					ctx.Permission())
			}
		}
		ctx.NotFound()
		return
	}
}
