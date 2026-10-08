// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	"forgejo.org/modules/setting"
)

func ReqExploreSignIn(ctx Context) {
	if (setting.Service.RequireSignInView || setting.Service.Explore.RequireSigninView) && !ctx.IsSigned() {
		ctx.Error(http.StatusUnauthorized, "reqExploreSignIn", "you must be signed in to search for users")
	}
}
