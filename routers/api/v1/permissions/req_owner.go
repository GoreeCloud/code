// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"
	"slices"

	"forgejo.org/models/unit"
)

func ReqOwner(ctx Context, unitTypes []unit.Type) {
	if len(unitTypes) > 0 && !slices.ContainsFunc(unitTypes, func(unitType unit.Type) bool {
		return ctx.Repository().UnitEnabled(ctx.Context(), unitType)
	}) {
		ctx.NotFound()
		return
	}
	if !ctx.Permission().IsOwner() && !IsUserSiteAdmin(ctx) {
		ctx.Error(http.StatusForbidden, "reqOwner", "user should be the owner of the repo")
		return
	}
}
