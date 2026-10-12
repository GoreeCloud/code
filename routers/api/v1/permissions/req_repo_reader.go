// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	"forgejo.org/models/unit"
)

func ReqRepoReader(ctx Context, unitType unit.Type) {
	if !ctx.Repository().UnitEnabled(ctx.Context(), unitType) {
		ctx.NotFound()
		return
	}
	if !ctx.Permission().CanRead(unitType) && !IsUserRepoAdmin(ctx) && !IsUserSiteAdmin(ctx) {
		ctx.Error(http.StatusForbidden, "reqRepoReader", "user should have specific read permission or be a repo admin or a site admin")
		return
	}
}
