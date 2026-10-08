// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package common

import (
	code_indexer "forgejo.org/modules/indexer/code"
	"forgejo.org/modules/setting"
	"forgejo.org/services/context"
)

type CodeSearchOptions struct {
	Language, Keyword, Path string
}

// Parses the common code search options from the context
// This functions takes care of the following ctx.Data fields
// - Keyword
// - Language
// - CodeSearchPath
func InitCodeSearchOptions(ctx *context.Context) (opts CodeSearchOptions) {
	opts.Language = ctx.FormTrim("l")
	opts.Keyword = ctx.FormTrim("q")
	opts.Path = ctx.FormTrim("path")

	ctx.Data["Keyword"] = opts.Keyword
	ctx.Data["Language"] = opts.Language
	ctx.Data["CodeSearchPath"] = opts.Path

	return opts
}

// Returns the indexer mode to be used by the code indexer
// Also sets the ctx.Data fields "CodeSearchMode" and "CodeSearchOptions"
//
// NOTE:
// This is separate from `InitCodeSearchOptions`
// since this is specific the indexer and only used
// where git-grep is not available.
func CodeSearchIndexerMode(ctx *context.Context) (mode code_indexer.SearchMode) {
	mode = code_indexer.SearchModeExact
	if m := ctx.FormTrim("mode"); m == "union" {
		mode = code_indexer.SearchModeUnion
	} else if m == "fuzzy" || ctx.FormBool("fuzzy") {
		if setting.Indexer.RepoIndexerEnableFuzzy {
			mode = code_indexer.SearchModeFuzzy
		} else {
			mode = code_indexer.SearchModeUnion
		}
	}
	ctx.Data["CodeSearchOptions"] = code_indexer.CodeSearchOptions
	ctx.Data["CodeSearchMode"] = mode.String()

	return mode
}
