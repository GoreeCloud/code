// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	"context"
	"fmt"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/perm"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/modules/structs"

	"xorm.io/builder"
)

// For specific repositories listed in [AccessTokenResourceRepo] models, all access is permitted.  For public
// repositories that aren't listed among the specific repos, read-only access is permitted.  For all other repos, no
// access is permitted.
type SpecificReposAuthorizationReducer struct {
	resourceRepos []*auth_model.AccessTokenResourceRepo
}

func (r *SpecificReposAuthorizationReducer) ReduceRepoAccess(ctx context.Context, repo *repo_model.Repository, accessMode perm.AccessMode) (perm.AccessMode, error) {
	for _, tokenRepo := range r.resourceRepos {
		if tokenRepo.RepoID == repo.ID {
			// No restrictions as this repo is within the scope of the access token.
			return accessMode, nil
		}
	}

	if err := repo.LoadOwner(ctx); err != nil {
		return 0, fmt.Errorf("failed to LoadOwner during ReduceRepoAccess: %w", err)
	}

	// Fine-grained access tokens remove access to any private repositories, or repository owned by non-public users,
	// that aren't listed in their resource list.
	if !repo.Owner.Visibility.IsPublic() || repo.IsPrivate {
		return perm.AccessModeNone, nil
	}

	// Public repos will be reduced to read access.
	return min(accessMode, perm.AccessModeRead), nil
}

func (r *SpecificReposAuthorizationReducer) RepoReadAccessFilter() builder.Cond {
	repoIDs := make([]int64, len(r.resourceRepos))
	for i, tokenRepo := range r.resourceRepos {
		repoIDs[i] = tokenRepo.RepoID
	}
	targetRepos := builder.In("repository.id", repoIDs)

	// We should also be able to see all non-private repositories that aren't in a private or limited organization.
	return builder.Or(
		targetRepos,
		builder.And(
			builder.Eq{"repository.is_private": false},
			builder.NotIn("repository.owner_id", builder.Select("id").From("`user`").Where(
				builder.Or(builder.Eq{"visibility": structs.VisibleTypeLimited}, builder.Eq{"visibility": structs.VisibleTypePrivate}),
			))))
}

func (*SpecificReposAuthorizationReducer) AllowAdminOverride() bool {
	return false
}
