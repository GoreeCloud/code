// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package utils

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
	"forgejo.org/routers/web/shared/user"
	"forgejo.org/services/authz"
	"forgejo.org/services/context"
)

// DeleteAccessToken deletes an access token for a user identified by ctx.ContextUser.
// Shared logic between user and admin token deletion endpoints.
func DeleteAccessToken(ctx *context.APIContext) {
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

// ListAccessTokens lists access tokens for a user identified by ctx.ContextUser.
// Shared logic between user and admin token listing endpoints.
func ListAccessTokens(ctx *context.APIContext) {
	opts := auth_model.ListAccessTokensOptions{UserID: ctx.User().ID, ListOptions: GetListOptions(ctx)}

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
			Created:        tokens[i].CreatedUnix.AsTime(),
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

// CreateAccessToken creates an access token for a user identified by ctx.ContextUser.
// Shared logic between user and admin token creation endpoints.
func CreateAccessToken(ctx *context.APIContext) {
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
		Created:        t.CreatedUnix.AsTime(),
		Repositories:   tokenRepositories,
	})
}
