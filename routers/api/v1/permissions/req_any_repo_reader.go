// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"
)

func ReqAnyRepoReader(ctx Context) {
	if !ctx.Permission().HasAccess() && !IsUserSiteAdmin(ctx) {
		ctx.Error(http.StatusForbidden, "reqAnyRepoReader", "user should have any permission to read repository or permissions of site admin")
		return
	}
}
