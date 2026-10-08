// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"cmp"
	stdCtx "context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	access_model "forgejo.org/models/perm/access"
	repo_model "forgejo.org/models/repo"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/web"
	"forgejo.org/routers/api/v1/utils"
	"forgejo.org/routers/web/shared/user"
	"forgejo.org/services/authz"
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

	opts := auth_model.ListAccessTokensOptions{UserID: ctx.User().ID, ListOptions: utils.GetListOptions(ctx)}

	tokens, count, err := db.FindAndCount[auth_model.AccessToken](ctx, opts)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	// Load all the AccessTokenResourceRepo for the tokens that we're returning:
	repoModelsByTokenID, err := repo_model.BulkGetRepositoriesForAccessTokens(ctx, tokens,
		func(repo *repo_model.Repository) (bool, error) {
			// Repos associated with a repo-specific access token should already be visible to the token owner, but it's
			// possible that access has changed, such as a removed collaborator on a repo -- don't provide info on that
			// repo if so.
			permission, err := access_model.GetUserRepoPermissionWithReducer(ctx, repo, ctx.Doer(), ctx.Reducer())
			if err != nil {
				return false, err
			}
			return permission.HasAccess(), nil
		})
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	// Convert map[int64]*Repository -> map[int64]*RepositoryMeta...
	reposByTokenID := make(map[int64][]*api.RepositoryMeta)
	for tokenID, repoModels := range repoModelsByTokenID {
		repos := make([]*api.RepositoryMeta, len(repoModels))
		for i, repo := range repoModels {
			repos[i] = &api.RepositoryMeta{
				ID:       repo.ID,
				Name:     repo.Name,
				Owner:    repo.OwnerName,
				FullName: repo.FullName(),
			}
		}
		reposByTokenID[tokenID] = repos
	}

	apiTokens := make([]*api.AccessToken, len(tokens))
	for i := range tokens {
		apiTokens[i] = &api.AccessToken{
			ID:             tokens[i].ID,
			Name:           tokens[i].Name,
			TokenLastEight: tokens[i].TokenLastEight,
			Scopes:         tokens[i].Scope.StringSlice(),
			Repositories:   reposByTokenID[tokens[i].ID],
		}
		// Provide a consistent sort order on repositories, helpful for test consistency.  Hard to do any earlier
		// because of the bulk loading maps.
		slices.SortFunc(apiTokens[i].Repositories, func(a, b *api.RepositoryMeta) int {
			return cmp.Compare(a.ID, b.ID)
		})
	}

	ctx.SetTotalCountHeader(count)
	ctx.JSON(http.StatusOK, &apiTokens)
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

	form := web.GetForm(ctx).(*api.CreateAccessTokenOption)

	t := &auth_model.AccessToken{
		UID:  ctx.User().ID,
		Name: form.Name,
	}

	exist, err := auth_model.AccessTokenByNameExists(ctx, t)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	if exist {
		ctx.Error(http.StatusBadRequest, "AccessTokenByNameExists", errors.New("access token name has been used already"))
		return
	}

	scope, err := auth_model.AccessTokenScope(strings.Join(form.Scopes, ",")).Normalize()
	if err != nil {
		ctx.Error(http.StatusBadRequest, "AccessTokenScope.Normalize", fmt.Errorf("invalid access token scope provided: %w", err))
		return
	}
	if scope == "" {
		ctx.Error(http.StatusBadRequest, "AccessTokenScope", "access token must have a scope")
		return
	}
	t.Scope = scope

	var resourceRepos []*auth_model.AccessTokenResourceRepo
	var tokenRepositories []*api.RepositoryMeta

	if form.Repositories != nil {
		repos := make([]*repo_model.Repository, len(form.Repositories))
		for i, repoTarget := range form.Repositories {
			repo, err := repo_model.GetRepositoryByOwnerAndName(ctx, repoTarget.Owner, repoTarget.Name)
			if err != nil && repo_model.IsErrRepoNotExist(err) {
				ctx.Error(http.StatusBadRequest, "GetRepositoryByOwnerAndName", fmt.Errorf("repository %s/%s does not exist", repoTarget.Owner, repoTarget.Name))
				return
			} else if err != nil {
				ctx.ServerError("GetRepositoryByOwnerAndName", err)
				return
			}
			permission, err := access_model.GetUserRepoPermissionWithReducer(ctx, repo, ctx.Doer(), ctx.Reducer())
			if err != nil {
				ctx.ServerError("GetUserRepoPermissionWithReducer", err)
				return
			} else if !permission.HasAccess() {
				// Prevent data existence probing -- ensure this error is the exact same as the !IsErrRepoNotExist case above
				ctx.Error(http.StatusBadRequest, "GetRepositoryByOwnerAndName", fmt.Errorf("repository %s/%s does not exist", repoTarget.Owner, repoTarget.Name))
				return
			}
			repos[i] = repo
		}

		for _, repo := range repos {
			resourceRepos = append(resourceRepos, &auth_model.AccessTokenResourceRepo{RepoID: repo.ID})
			tokenRepositories = append(tokenRepositories, &api.RepositoryMeta{
				ID:       repo.ID,
				Name:     repo.Name,
				Owner:    repo.OwnerName,
				FullName: repo.FullName(),
			})
		}

		t.ResourceAllRepos = false
	} else {
		// token has access to all repository resources
		t.ResourceAllRepos = true
	}

	if err := authz.ValidateAccessToken(t, resourceRepos); err != nil {
		s := user.TranslateAccessTokenValidationError(ctx.Base, err)
		if has, str := s.Get(); has {
			ctx.Error(http.StatusBadRequest, "ValidateAccessToken", str)
			return
		}
		ctx.ServerError("ValidateAccessToken", err)
		return
	}

	err = db.WithTx(ctx, func(ctx stdCtx.Context) error {
		if err := auth_model.NewAccessToken(ctx, t); err != nil {
			return err
		}
		return auth_model.InsertAccessTokenResourceRepos(ctx, t.ID, resourceRepos)
	})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "NewAccessToken", err)
		return
	}
	ctx.JSON(http.StatusCreated, &api.AccessToken{
		Name:           t.Name,
		Token:          t.Token,
		ID:             t.ID,
		TokenLastEight: t.TokenLastEight,
		Scopes:         t.Scope.StringSlice(),
		Repositories:   tokenRepositories,
	})
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

	token := ctx.Params(":id")
	tokenID, _ := strconv.ParseInt(token, 0, 64)

	if tokenID == 0 {
		tokens, err := db.Find[auth_model.AccessToken](ctx, auth_model.ListAccessTokensOptions{
			Name:   token,
			UserID: ctx.User().ID,
		})
		if err != nil {
			ctx.Error(http.StatusInternalServerError, "ListAccessTokens", err)
			return
		}

		switch len(tokens) {
		case 0:
			ctx.NotFound()
			return
		case 1:
			tokenID = tokens[0].ID
		default:
			ctx.Error(http.StatusUnprocessableEntity, "DeleteAccessTokenByID", fmt.Errorf("multiple matches for token name '%s'", token))
			return
		}
	}
	if tokenID == 0 {
		ctx.Error(http.StatusInternalServerError, "Invalid TokenID", nil)
		return
	}

	if err := auth_model.DeleteAccessTokenByID(ctx, tokenID, ctx.User().ID); err != nil {
		if auth_model.IsErrAccessTokenNotExist(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "DeleteAccessTokenByID", err)
		}
		return
	}

	ctx.Status(http.StatusNoContent)
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
