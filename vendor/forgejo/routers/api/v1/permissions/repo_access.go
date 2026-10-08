// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/perm"
	access_model "forgejo.org/models/perm/access"
	"forgejo.org/models/unit"
	user_model "forgejo.org/models/user"
)

func RepoAccess(ctx Context) {
	if ctx.Doer() != nil && ctx.Doer().ID == user_model.ActionsUserID && ctx.Authentication().ActionsTaskID().Has() {
		_, taskID := ctx.Authentication().ActionsTaskID().Get()
		task, err := actions_model.GetTaskByID(ctx.Context(), taskID)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "actions_model.GetTaskByID", err)
			return
		}
		if task.RepoID != ctx.Repository().ID {
			ctx.NotFound()
			return
		}

		if task.IsForkPullRequest {
			ctx.Permission().AccessMode = perm.AccessModeRead
		} else {
			ctx.Permission().AccessMode = perm.AccessModeWrite
		}

		if err := ctx.Repository().LoadUnits(ctx.Context()); err != nil {
			ctx.Error(http.StatusInternalServerError, "LoadUnits", err)
			return
		}
		ctx.Permission().Units = ctx.Repository().Units
		ctx.Permission().UnitsMode = make(map[unit.Type]perm.AccessMode)
		for _, u := range ctx.Repository().Units {
			ctx.Permission().UnitsMode[u.Type] = ctx.Permission().AccessMode
		}
	} else {
		permission, err := access_model.GetUserRepoPermissionWithReducer(ctx.Context(), ctx.Repository(), ctx.Doer(), ctx.Reducer())
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "GetUserRepoPermissionWithReducer", err)
			return
		}
		ctx.SetPermission(&permission)
	}

	if !ctx.Permission().HasAccess() {
		ctx.NotFound()
		return
	}
}
