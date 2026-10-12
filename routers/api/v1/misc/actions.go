// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package misc

import (
	"net/http"

	"forgejo.org/models/actions"
	"forgejo.org/services/context"
	"forgejo.org/services/convert"
)

// Retrieves a workflow run associated with a token
func GetActionsRun(ctx *context.APIContext) {
	// swagger:operation GET /actions/run miscellaneous getActionsRun
	// ---
	// summary: Get a workflow run associated with a token
	// description: >
	//   The automatic actions token must be used as the authentication mechanism
	//   (<code>Authorization: Bearer $&#x7b;&#x7b; forgejo.token &#x7d;&#x7d;</code>); other
	//   types of tokens cannot be used. The token is associated with the job, which must be still
	//   running for the request to this endpoint to succeed.
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/ActionRun"
	hasTaskID, taskID := ctx.Authentication().ActionsTaskID().Get()
	if !hasTaskID {
		ctx.Error(http.StatusForbidden, "", "must use an automatic actions token")
		return
	}
	task, err := actions.GetTaskByID(ctx, taskID)
	if err != nil {
		ctx.ServerError("GetTaskByID", err)
		return
	}
	if err := task.LoadJob(ctx); err != nil {
		ctx.ServerError("LoadJob", err)
		return
	}
	if err := task.Job.LoadRun(ctx); err != nil {
		ctx.ServerError("LoadRun", err)
		return
	}
	if err := task.Job.Run.LoadAttributes(ctx); err != nil {
		ctx.ServerError("LoadAttributes", err)
		return
	}
	ctx.JSON(http.StatusOK, convert.ToActionRun(ctx, task.Job.Run, ctx.Doer()))
}
