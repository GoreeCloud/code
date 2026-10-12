// Copyright 2022 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"net/http"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/modules/base"
	code_indexer "forgejo.org/modules/indexer/code"
	"forgejo.org/modules/setting"
	"forgejo.org/routers/common"
	shared_user "forgejo.org/routers/web/shared/user"
	"forgejo.org/services/context"
)

const (
	tplUserCode base.TplName = "user/code"
)

// CodeSearch render user/organization code search page
func CodeSearch(ctx *context.Context) {
	if !setting.Indexer.RepoIndexerEnabled {
		ctx.Redirect(ctx.ContextUser.HomeLink())
		return
	}
	shared_user.PrepareContextForProfileBigAvatar(ctx)
	shared_user.RenderUserHeader(ctx)

	if err := shared_user.LoadHeaderCount(ctx); err != nil {
		ctx.ServerError("LoadHeaderCount", err)
		return
	}

	ctx.Data["IsPackageEnabled"] = setting.Packages.Enabled
	ctx.Data["IsRepoIndexerEnabled"] = setting.Indexer.RepoIndexerEnabled
	ctx.Data["IsCodePage"] = true
	ctx.Data["Title"] = ctx.Tr("explore.code")

	opts := common.InitCodeSearchOptions(ctx)
	mode := common.CodeSearchIndexerMode(ctx)

	if opts.Keyword == "" {
		ctx.HTML(http.StatusOK, tplUserCode)
		return
	}

	var (
		repoIDs []int64
		err     error
	)

	page := ctx.FormInt("page")
	if page <= 0 {
		page = 1
	}

	repoIDs, err = repo_model.FindUserCodeAccessibleOwnerRepoIDs(ctx, ctx.ContextUser.ID, ctx.Doer)
	if err != nil {
		ctx.ServerError("FindUserCodeAccessibleOwnerRepoIDs", err)
		return
	}

	var (
		total                 int
		searchResults         code_indexer.Results
		searchResultLanguages []*code_indexer.SearchResultLanguages
	)

	if len(repoIDs) > 0 {
		total, searchResults, searchResultLanguages, err = code_indexer.PerformSearch(ctx, &code_indexer.SearchOptions{
			RepoIDs:  repoIDs,
			Keyword:  opts.Keyword,
			Mode:     mode,
			Language: opts.Language,
			Filename: opts.Path,
			Paginator: &db.ListOptions{
				Page:     page,
				PageSize: setting.UI.RepoSearchPagingNum,
			},
		})
		if err != nil {
			if code_indexer.IsAvailable(ctx) {
				ctx.ServerError("SearchResults", err)
				return
			}
			ctx.Data["CodeIndexerUnavailable"] = true
		} else {
			ctx.Data["CodeIndexerUnavailable"] = !code_indexer.IsAvailable(ctx)
		}

		loadRepoIDs := searchResults.RepoIDs()
		repoMaps, err := repo_model.GetRepositoriesMapByIDs(ctx, loadRepoIDs)
		if err != nil {
			ctx.ServerError("GetRepositoriesMapByIDs", err)
			return
		}

		ctx.Data["RepoMaps"] = repoMaps
	}
	ctx.Data["SearchResults"] = searchResults
	ctx.Data["SearchResultLanguages"] = searchResultLanguages

	pager := context.NewPagination(total, setting.UI.RepoSearchPagingNum, page, 5)
	pager.SetDefaultParams(ctx)
	pager.AddParam(ctx, "l", "Language")
	pager.AddParam(ctx, "mode", "CodeSearchMode")
	ctx.Data["Page"] = pager

	ctx.HTML(http.StatusOK, tplUserCode)
}
