// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package utils

import (
	"forgejo.org/models/db"
	"forgejo.org/services/context"
)

func GetDbSearchOrder(ctx *context.APIContext) db.SearchOrderBy {
	switch ctx.FormString("sort") {
	case "oldest":
		return db.SearchOrderByOldest
	case "newest":
		return db.SearchOrderByNewest
	case "alphabetically":
		return db.SearchOrderByAlphabetically
	case "reversealphabetically":
		return db.SearchOrderByAlphabeticallyReverse
	case "recentupdate":
		return db.SearchOrderByRecentUpdated
	case "leastupdate":
		return db.SearchOrderByLeastUpdated
	default:
		return db.SearchOrderByAlphabetically
	}
}
