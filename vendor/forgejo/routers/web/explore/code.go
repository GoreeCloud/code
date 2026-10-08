// Copyright 2021 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package explore

import (
	"net/http"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/modules/base"
	code_indexer "forgejo.org/modules/indexer/code"
	"forgejo.org/modules/setting"
	"forgejo.org/routers/common"
	"forgejo.org/services/context"
)

const (
	// tplExploreCode explore code page template
	tplExploreCode base.TplName = "explore/code"
)

// Code render explore code page
func Code(ctx *context.Context) {
	if !setting.Indexer.RepoIndexerEnabled || setting.Service.Explore.DisableCodePage {
		ctx.Redirect(setting.AppSubURL + "/explore")
		return
	}

	ctx.Data["UsersPageIsDisabled"] = setting.Service.Explore.DisableUsersPage
	ctx.Data["OrganizationsPageIsDisabled"] = setting.Service.Explore.DisableOrganizationsPage
	ctx.Data["IsRepoIndexerEnabled"] = setting.Indexer.RepoIndexerEnabled
	ctx.Data["Title"] = ctx.Tr("explore")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreCode"] = true
	ctx.Data["PageIsViewCode"] = true

	opts := common.InitCodeSearchOptions(ctx)
	mode := common.CodeSearchIndexerMode(ctx)

	if opts.Keyword == "" {
		ctx.HTML(http.StatusOK, tplExploreCode)
		return
	}

	page := ctx.FormInt("page")
	if page <= 0 {
		page = 1
	}

	var (
		repoIDs []int64
		err     error
		isAdmin bool
	)
	if ctx.Doer != nil {
		isAdmin = ctx.Doer.IsAdmin
	}

	// guest user or non-admin user
	if ctx.Doer == nil || !isAdmin {
		repoIDs, err = repo_model.FindUserCodeAccessibleRepoIDs(ctx, ctx.Doer)
		if err != nil {
			ctx.ServerError("FindUserCodeAccessibleRepoIDs", err)
			return
		}
	}

	var (
		total                 int
		searchResults         code_indexer.Results
		searchResultLanguages []*code_indexer.SearchResultLanguages
	)

	if (len(repoIDs) > 0) || isAdmin {
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

		if len(loadRepoIDs) != len(repoMaps) {
			// Remove deleted repos from search results
			cleanedSearchResults := make([]*code_indexer.Result, 0, len(repoMaps))
			for _, sr := range searchResults {
				if _, found := repoMaps[sr.RepoID]; found {
					cleanedSearchResults = append(cleanedSearchResults, sr)
				}
			}

			searchResults = cleanedSearchResults
		}
	}

	ctx.Data["SearchResults"] = searchResults
	ctx.Data["SearchResultLanguages"] = searchResultLanguages

	pager := context.NewPagination(total, setting.UI.RepoSearchPagingNum, page, 5)
	pager.SetDefaultParams(ctx)
	pager.AddParam(ctx, "l", "Language")
	pager.AddParam(ctx, "mode", "CodeSearchMode")
	ctx.Data["Page"] = pager

	ctx.HTML(http.StatusOK, tplExploreCode)
}
