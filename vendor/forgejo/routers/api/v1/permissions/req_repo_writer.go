// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"
	"slices"

	"forgejo.org/models/unit"
)

func ReqRepoWriter(ctx Context, unitTypes []unit.Type) {
	if !slices.ContainsFunc(unitTypes, func(unitType unit.Type) bool {
		return ctx.Repository().UnitEnabled(ctx.Context(), unitType)
	}) {
		ctx.NotFound()
		return
	}
	if !IsUserRepoWriter(ctx, unitTypes) && !IsUserRepoAdmin(ctx) && !IsUserSiteAdmin(ctx) {
		ctx.Error(http.StatusForbidden, "reqRepoWriter", "user should have a permission to write to a repo")
		return
	}
}
