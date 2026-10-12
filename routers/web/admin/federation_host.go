// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin

import (
	"net/http"

	"forgejo.org/models/db"
	"forgejo.org/models/forgefed"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/base"
	"forgejo.org/modules/setting"
	"forgejo.org/services/context"
)

const (
	tplFederationHost base.TplName = "admin/federation/host"
)

func FederationHost(ctx *context.Context) {
	federationHostID := ctx.ParamsInt64("id")
	page := max(ctx.FormInt("page"), 1)

	host, err := forgefed.GetFederationHost(ctx, federationHostID)
	if err != nil {
		ctx.ServerError("GetFederationHost", err)
		return
	}

	users, err := user_model.FindFederatedUsersByHostID(ctx, federationHostID, db.ListOptions{
		PageSize: setting.UI.Admin.FederationUserPagingNum,
		Page:     page,
	})
	if err != nil {
		ctx.ServerError("FindFederatedUsersByHostID", err)
		return
	}

	total, err := user_model.CountFederatedUsersByHostID(ctx, federationHostID)
	if err != nil {
		ctx.ServerError("CountFederatedUsersByHostID", err)
		return
	}

	ctx.Data["Host"] = host
	ctx.Data["Users"] = users
	ctx.Data["UsersTotal"] = int(total)
	ctx.Data["Title"] = ctx.Tr("admin.federation.hosts.details_panel")
	ctx.Data["PageIsAdminFederationHosts"] = true

	numPages := 0
	if total > 0 {
		numPages = (int(total) - 1/setting.UI.Admin.FederationUserPagingNum)
	}

	pager := context.NewPagination(int(total), setting.UI.Admin.FederationUserPagingNum, page, numPages)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager

	ctx.HTML(http.StatusOK, tplFederationHost)
}
