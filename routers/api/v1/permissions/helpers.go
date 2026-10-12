// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"slices"

	"forgejo.org/models/unit"
)

func IsUserSiteAdmin(ctx Context) bool {
	if !ctx.Reducer().AllowAdminOverride() {
		return false
	}
	return ctx.IsSigned() && ctx.Doer().IsAdmin
}

func IsUserRepoAdmin(ctx Context) bool {
	if !ctx.Reducer().AllowAdminOverride() {
		return false
	}
	return ctx.Permission().IsAdmin()
}

func IsUserRepoWriter(ctx Context, unitTypes []unit.Type) bool {
	return slices.ContainsFunc(unitTypes, ctx.Permission().CanWrite)
}
