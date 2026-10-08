// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	"forgejo.org/models/perm"
)

func ReqPackageAccess(ctx Context, accessMode perm.AccessMode) {
	if ctx.PackageAccessMode() < accessMode && !IsUserSiteAdmin(ctx) {
		ctx.Error(http.StatusForbidden, "reqPackageAccess", "user should have specific permission or be a site admin")
		return
	}
}
