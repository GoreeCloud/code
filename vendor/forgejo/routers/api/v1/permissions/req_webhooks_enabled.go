// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	"forgejo.org/modules/setting"
)

func ReqWebhooksEnabled(ctx Context) {
	if setting.DisableWebhooks {
		ctx.Error(http.StatusForbidden, "", "webhooks disabled by administrator")
		return
	}
}
