// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	"forgejo.org/models/organization"
)

func ReqTeamMembership(ctx Context) {
	if IsUserSiteAdmin(ctx) {
		return
	}
	if ctx.Team() == nil {
		ctx.Error(http.StatusInternalServerError, "", "reqTeamMembership: unprepared context")
		return
	}

	orgID := ctx.Team().OrgID
	isOwner, err := organization.IsOrganizationOwner(ctx.Context(), orgID, ctx.Doer().ID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "IsOrganizationOwner", err)
		return
	} else if isOwner {
		return
	}

	if isTeamMember, err := organization.IsTeamMember(ctx.Context(), orgID, ctx.Team().ID, ctx.Doer().ID); err != nil {
		ctx.Error(http.StatusInternalServerError, "IsTeamMember", err)
		return
	} else if !isTeamMember {
		isOrgMember, err := organization.IsOrganizationMember(ctx.Context(), orgID, ctx.Doer().ID)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "IsOrganizationMember", err)
		} else if isOrgMember {
			ctx.Error(http.StatusForbidden, "", "Must be a team member")
		} else {
			ctx.NotFound()
		}
		return
	}
}
