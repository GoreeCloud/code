// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin

import (
	"net/http"

	"forgejo.org/models/db"
	"forgejo.org/models/forgefed"
	"forgejo.org/modules/base"
	"forgejo.org/modules/setting"
	"forgejo.org/services/context"
)

const (
	tplFederationHosts base.TplName = "admin/federation/hosts"
)

func FederationHosts(ctx *context.Context) {
	sort := ctx.FormTrim("sort")
	page := max(ctx.FormInt("page"), 1)

	hosts, err := forgefed.FindFederationHosts(ctx, db.ListOptions{
		Page:     page,
		PageSize: setting.UI.Admin.FederationHostPagingNum,
	})
	if err != nil {
		ctx.ServerError("GetFederationHosts", err)
		return
	}

	total, err := forgefed.CountFederationHosts(ctx)
	if err != nil {
		ctx.ServerError("CountFederationHosts", err)
		return
	}

	ctx.Data["Title"] = ctx.Tr("admin.federation.hosts.title")
	ctx.Data["PageIsAdminFederationHosts"] = true
	ctx.Data["SortType"] = sort
	ctx.Data["TotalCount"] = int(total)
	ctx.Data["Hosts"] = hosts

	numPages := 0
	if total > 0 {
		numPages = (int(total) - 1/setting.UI.Admin.FederationHostPagingNum)
	}

	pager := context.NewPagination(int(total), setting.UI.Admin.FederationHostPagingNum, page, numPages)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager

	ctx.HTML(http.StatusOK, tplFederationHosts)
}
