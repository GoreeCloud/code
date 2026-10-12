// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	"context"
	"errors"
	"fmt"

	auth_model "forgejo.org/models/auth"
)

func GetAuthorizationReducerForAccessToken(ctx context.Context, token *auth_model.AccessToken) (AuthorizationReducer, error) {
	if token.ResourceAllRepos {
		if publicOnly, err := token.Scope.PublicOnly(); err != nil {
			return nil, fmt.Errorf("PublicOnly: %w", err)
		} else if publicOnly {
			return &PublicReposAuthorizationReducer{}, nil
		}
		return &AllAccessAuthorizationReducer{}, nil
	}

	repos, err := auth_model.GetRepositoriesAccessibleWithToken(ctx, token.ID)
	if err != nil {
		return nil, fmt.Errorf("GetRepositoriesAccessibleWithToken: %w", err)
	}
	// Cast slice into []RepoGetter
	iface := make([]RepoGetter, len(repos))
	for i, r := range repos {
		iface[i] = r
	}
	return &SpecificReposAuthorizationReducer{ResourceRepos: iface}, nil
}

// Validate that an access token's state is valid for creation.  For example, that it doesn't have a conflicting set of
// resources (public-only and specific repositories), and other similar checks.
func ValidateAccessToken(token *auth_model.AccessToken, repoResources []*auth_model.AccessTokenResourceRepo) error {
	// Other validations may be added here in the future.
	return ValidateRepositoryResource(token.ResourceAllRepos, token.Scope, len(repoResources))
}

var (
	ErrSpecifiedReposNone         = errors.New("specified repository access token: must have at least one repository")
	ErrSpecifiedReposNoPublicOnly = errors.New("specified repository access token: cannot be combined with public-only scope")
	ErrSpecifiedReposInvalidScope = errors.New("specified repository access token: invalid scope")
)

func ValidateRepositoryResource(resourceAllRepos bool, scope auth_model.AccessTokenScope, numRepoResources int) error {
	// Access tokens with broad access to all resources don't have any relevant validation rules to apply.
	if resourceAllRepos {
		return nil
	}

	// Repo-specific access token must have at least one repository.
	if numRepoResources == 0 {
		return ErrSpecifiedReposNone
	}

	// Can't have public-only and specified repos -- that's a combination that doesn't make sense.
	if publicOnly, err := scope.PublicOnly(); err != nil {
		return err
	} else if publicOnly {
		return ErrSpecifiedReposNoPublicOnly
	}

	// Repo-specific access tokens are only effective at restricting permissions if they are limited to the scopes that
	// support repositories as a resource.  For example, if you had a repo-specific token but then gave it
	// `write:organization`, it would be able to do operations like delete an organization -- permission checks on the
	// repository resources wouldn't be applicable to the organization resources.
	for _, scope := range scope.StringSlice() {
		switch auth_model.AccessTokenScope(scope) {
		case auth_model.AccessTokenScopeReadIssue,
			auth_model.AccessTokenScopeWriteIssue,
			auth_model.AccessTokenScopeReadRepository,
			auth_model.AccessTokenScopeWriteRepository:
			continue
		default:
			return fmt.Errorf("%w: cannot be combined with scope %s", ErrSpecifiedReposInvalidScope, scope)
		}
	}

	return nil
}
