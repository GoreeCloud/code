// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package repo

import (
	"context"

	"forgejo.org/models/perm"

	"xorm.io/builder"
)

// Defines an API for reducing available permissions to specific repositories.
type RepositoryAuthorizationReducer interface {
	// Given a repository and an accessMode, ReduceRepoAccess will return a new, possibly reduced, AccessMode that
	// reflects the actual access that is currently permitted.  For example, when using a fine-grained access token that
	// only grants write access to one target repository, `ReduceRepoAccess(target, AccessModeWrite)` would return
	// `AccessModeWrite`, and `ReduceRepoAccess(other-repo, AccessModeWrite)` would return a lesser access mode,
	// restricting access to other repositories.
	ReduceRepoAccess(ctx context.Context, repo *Repository, accessMode perm.AccessMode) (perm.AccessMode, error)

	// If querying the repository table, apply the provided condition to query only repositories that the restriction
	// will allow AccessModeRead (or higher).  For example, when using a fine-grained access token that only grants
	// write access to one target repository, `RepoReadAccessFilter()` will return a query condition that provides
	// visibility for all the public repos (which have read access) and all the target private repos (which have write
	// access, which is greater-than read access).
	RepoReadAccessFilter() builder.Cond
}
