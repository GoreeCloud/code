// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"net/http"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/web"
	"forgejo.org/routers/api/v1/utils"
	"forgejo.org/services/context"
	"forgejo.org/services/convert"
)

// ListAccessTokens list all the access tokens
func ListAccessTokens(ctx *context.APIContext) {
	// swagger:operation GET /users/{username}/tokens user userGetTokens
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

// CreateAccessToken creates an access token
func CreateAccessToken(ctx *context.APIContext) {
	// swagger:operation POST /users/{username}/tokens user userCreateToken
	// ---
	// summary: Generate an access token for the specified user
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

// DeleteAccessToken deletes an access token
func DeleteAccessToken(ctx *context.APIContext) {
	// swagger:operation DELETE /users/{username}/tokens/{token} user userDeleteAccessToken
	// ---
	// summary: Delete an access token from the specified user's account
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

// CreateOauth2Application is the handler to create a new OAuth2 Application for the authenticated user
func CreateOauth2Application(ctx *context.APIContext) {
	// swagger:operation POST /user/applications/oauth2 user userCreateOAuth2Application
	// ---
	// summary: Creates a new OAuth2 application
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/CreateOAuth2ApplicationOptions"
	// responses:
	//   "201":
	//     "$ref": "#/responses/OAuth2Application"
	//   "400":
	//     "$ref": "#/responses/error"
	//   "401":
	//     "$ref": "#/responses/unauthorized"
	//   "403":
	//     "$ref": "#/responses/forbidden"

	data := web.GetForm(ctx).(*api.CreateOAuth2ApplicationOptions)

	app, err := auth_model.CreateOAuth2Application(ctx, auth_model.CreateOAuth2ApplicationOptions{
		Name:               data.Name,
		UserID:             ctx.Doer().ID,
		RedirectURIs:       data.RedirectURIs,
		ConfidentialClient: data.ConfidentialClient,
	})
	if err != nil {
		ctx.Error(http.StatusBadRequest, "", "error creating oauth2 application")
		return
	}
	secret, err := app.GenerateClientSecret(ctx)
	if err != nil {
		ctx.Error(http.StatusBadRequest, "", "error creating application secret")
		return
	}
	app.ClientSecret = secret

	ctx.JSON(http.StatusCreated, convert.ToOAuth2Application(app))
}

// ListOauth2Applications list all the Oauth2 application
func ListOauth2Applications(ctx *context.APIContext) {
	// swagger:operation GET /user/applications/oauth2 user userGetOAuth2Applications
	// ---
	// summary: List the authenticated user's oauth2 applications
	// produces:
	// - application/json
	// parameters:
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
	//     "$ref": "#/responses/OAuth2ApplicationList"
	//   "401":
	//     "$ref": "#/responses/unauthorized"
	//   "403":
	//     "$ref": "#/responses/forbidden"

	apps, total, err := db.FindAndCount[auth_model.OAuth2Application](ctx, auth_model.FindOAuth2ApplicationsOptions{
		ListOptions: utils.GetListOptions(ctx),
		OwnerID:     ctx.Doer().ID,
	})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "ListOAuth2Applications", err)
		return
	}

	apiApps := make([]*api.OAuth2Application, len(apps))
	for i := range apps {
		apiApps[i] = convert.ToOAuth2Application(apps[i])
		apiApps[i].ClientSecret = "" // Hide secret on application list
	}

	ctx.SetTotalCountHeader(total)
	ctx.JSON(http.StatusOK, &apiApps)
}

// DeleteOauth2Application delete OAuth2 application
func DeleteOauth2Application(ctx *context.APIContext) {
	// swagger:operation DELETE /user/applications/oauth2/{id} user userDeleteOAuth2Application
	// ---
	// summary: Delete an OAuth2 application
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: token to be deleted
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "401":
	//     "$ref": "#/responses/unauthorized"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	appID := ctx.ParamsInt64(":id")
	if err := auth_model.DeleteOAuth2Application(ctx, appID, ctx.Doer().ID); err != nil {
		if auth_model.IsErrOAuthApplicationNotFound(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "DeleteOauth2ApplicationByID", err)
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetOauth2Application returns an OAuth2 application
func GetOauth2Application(ctx *context.APIContext) {
	// swagger:operation GET /user/applications/oauth2/{id} user userGetOAuth2Application
	// ---
	// summary: Get an OAuth2 application
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: Application ID to be found
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/OAuth2Application"
	//   "401":
	//     "$ref": "#/responses/unauthorized"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	appID := ctx.ParamsInt64(":id")
	app, err := auth_model.GetOAuth2ApplicationByID(ctx, appID)
	if err != nil {
		if auth_model.IsErrOauthClientIDInvalid(err) || auth_model.IsErrOAuthApplicationNotFound(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "GetOauth2ApplicationByID", err)
		}
		return
	}
	if app.UID != ctx.Doer().ID {
		ctx.NotFound()
		return
	}

	app.ClientSecret = ""

	ctx.JSON(http.StatusOK, convert.ToOAuth2Application(app))
}

// UpdateOauth2Application updates an OAuth2 application
func UpdateOauth2Application(ctx *context.APIContext) {
	// swagger:operation PATCH /user/applications/oauth2/{id} user userUpdateOAuth2Application
	// ---
	// summary: Update an OAuth2 application, this includes regenerating the client secret
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: application to be updated
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/CreateOAuth2ApplicationOptions"
	// responses:
	//   "200":
	//     "$ref": "#/responses/OAuth2Application"
	//   "401":
	//     "$ref": "#/responses/unauthorized"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	appID := ctx.ParamsInt64(":id")

	data := web.GetForm(ctx).(*api.CreateOAuth2ApplicationOptions)

	app, err := auth_model.UpdateOAuth2Application(ctx, auth_model.UpdateOAuth2ApplicationOptions{
		Name:               data.Name,
		UserID:             ctx.Doer().ID,
		ID:                 appID,
		RedirectURIs:       data.RedirectURIs,
		ConfidentialClient: data.ConfidentialClient,
	})
	if err != nil {
		if auth_model.IsErrOauthClientIDInvalid(err) || auth_model.IsErrOAuthApplicationNotFound(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "UpdateOauth2ApplicationByID", err)
		}
		return
	}
	app.ClientSecret, err = app.GenerateClientSecret(ctx)
	if err != nil {
		ctx.Error(http.StatusBadRequest, "", "error updating application secret")
		return
	}

	ctx.JSON(http.StatusOK, convert.ToOAuth2Application(app))
}
