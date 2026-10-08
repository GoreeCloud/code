// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"net/http"

	"forgejo.org/modules/log"
	"forgejo.org/modules/private"
	"forgejo.org/modules/web"
	"forgejo.org/services/context"
)

// SSHLog hook to response ssh log
func SSHLog(ctx *context.PrivateContext) {
	logger := log.GetManager().GetLogger("ssh")
	if !logger.IsEnabled() {
		ctx.Status(http.StatusOK)
		return
	}

	opts := web.GetForm(ctx).(*private.SSHLogOption)
	logger.Log(0, opts.Level, "ssh: %v", opts.Message)
	ctx.Status(http.StatusOK)
}
