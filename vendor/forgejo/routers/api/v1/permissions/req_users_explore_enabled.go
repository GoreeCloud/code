// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"forgejo.org/modules/setting"
)

func ReqUsersExploreEnabled(ctx Context) {
	if setting.Service.Explore.DisableUsersPage {
		ctx.NotFound()
	}
}
