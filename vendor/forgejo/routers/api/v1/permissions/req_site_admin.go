// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"
)

func ReqSiteAdmin(ctx Context) {
	if !IsUserSiteAdmin(ctx) {
		ctx.Error(http.StatusForbidden, "reqSiteAdmin", "user should be the site admin")
		return
	}
}
