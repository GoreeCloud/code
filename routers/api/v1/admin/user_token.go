// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"forgejo.org/routers/api/v1/utils"
	"forgejo.org/services/context"
)

// ListUserAccessTokens lists all access tokens for a given user.
// This endpoint is admin-only and does not require Basic auth.
func ListUserAccessTokens(ctx *context.APIContext) {
	// swagger:operation GET /admin/users/{username}/tokens admin adminListUserAccessTokens
	// ---
	// summary: List the specified user's access tokens
	// produces:
	// - application/json
	// parameters:
	// - name: username
	//   in: path
	//   description: username of user
	//   type: string
	//   required: true
	// - name: page
	//   in: query
	//   description: page number of results to return (1-based)
	//   type: integer
	// - name: limit
	//   in: query
	//   description: page size of results
	//   type: integer
	// responses:
	//   "200":
	//     "$ref": "#/responses/AccessTokenList"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	utils.ListAccessTokens(ctx)
}

// CreateUserAccessToken creates a new access token for a given user.
// This endpoint is admin-only and does not require Basic auth.
func CreateUserAccessToken(ctx *context.APIContext) {
	// swagger:operation POST /admin/users/{username}/tokens admin adminCreateUserAccessToken
	// ---
	// summary: Create an access token for the specified user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: username
	//   in: path
	//   description: username of user
	//   required: true
	//   type: string
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateAccessTokenOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/AccessToken"
	//   "400":
	//     "$ref": "#/responses/error"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	utils.CreateAccessToken(ctx)
}

// DeleteUserAccessToken deletes an access token for a given user.
// This endpoint is admin-only and does not require Basic auth.
func DeleteUserAccessToken(ctx *context.APIContext) {
	// swagger:operation DELETE /admin/users/{username}/tokens/{token} admin adminDeleteUserAccessToken
	// ---
	// summary: Delete an access token for the specified user
	// produces:
	// - application/json
	// parameters:
	// - name: username
	//   in: path
	//   description: username of user
	//   type: string
	//   required: true
	// - name: token
	//   in: path
	//   description: token to be deleted, identified by ID and if not available by name
	//   type: string
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	//   "422":
	//     "$ref": "#/responses/error"

	utils.DeleteAccessToken(ctx)
}
