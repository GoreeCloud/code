// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	user_model "forgejo.org/models/user"
)

func ReqSelfOrAdmin(ctx Context) {
	getID := func(user *user_model.User) int64 {
		if user == nil {
			return 0
		}
		return user.ID
	}

	if !IsUserSiteAdmin(ctx) && getID(ctx.User()) != getID(ctx.Doer()) {
		ctx.Error(http.StatusForbidden, "reqSelfOrAdmin", "doer should be the site admin or be same as the contextUser")
		return
	}
}
