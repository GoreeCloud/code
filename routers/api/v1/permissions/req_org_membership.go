// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	"forgejo.org/models/organization"
)

func ReqOrgMembership(ctx Context) {
	if IsUserSiteAdmin(ctx) {
		return
	}

	var orgID int64
	if ctx.Organization() != nil {
		orgID = ctx.Organization().ID
	} else if ctx.Team() != nil {
		orgID = ctx.Team().OrgID
	} else {
		ctx.Error(http.StatusInternalServerError, "", "reqOrgMembership: unprepared context")
		return
	}

	if isMember, err := organization.IsOrganizationMember(ctx.Context(), orgID, ctx.Doer().ID); err != nil {
		ctx.Error(http.StatusInternalServerError, "IsOrganizationMember", err)
		return
	} else if !isMember {
		if ctx.Organization() != nil {
			ctx.Error(http.StatusForbidden, "", "Must be an organization member")
		} else {
			ctx.NotFound()
		}
		return
	}
}
