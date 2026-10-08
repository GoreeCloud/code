// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	"forgejo.org/models/organization"
)

func ReqOrgOwnership(ctx Context) {
	if IsUserSiteAdmin(ctx) {
		return
	}

	var orgID int64
	if ctx.Organization() != nil {
		orgID = ctx.Organization().ID
	} else if ctx.Team() != nil {
		orgID = ctx.Team().OrgID
	} else {
		ctx.Error(http.StatusInternalServerError, "", "reqOrgOwnership: unprepared context")
		return
	}

	isOwner, err := organization.IsOrganizationOwner(ctx.Context(), orgID, ctx.Doer().ID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "IsOrganizationOwner", err)
		return
	} else if !isOwner {
		if ctx.Organization() != nil {
			ctx.Error(http.StatusForbidden, "", "Must be an organization owner")
		} else {
			ctx.NotFound()
		}
		return
	}
}
