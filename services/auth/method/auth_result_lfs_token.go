// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	auth_model "forgejo.org/models/auth"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/lfs"
	"forgejo.org/modules/optional"
	"forgejo.org/services/auth"
	"forgejo.org/services/authz"
)

var _ auth.AuthenticationResult = &lfsTokenAuthenticationResult{}

type lfsTokenAuthenticationResult struct {
	*auth.BaseAuthenticationResult
	user   *user_model.User
	claims *lfs.Claims
}

func (r *lfsTokenAuthenticationResult) User() *user_model.User {
	return r.user
}

func (r *lfsTokenAuthenticationResult) Scope() optional.Option[auth_model.AccessTokenScope] {
	if r.claims.Op == "download" {
		return optional.Some(auth_model.AccessTokenScopeReadRepository)
	}
	return optional.Some(auth_model.AccessTokenScopeWriteRepository)
}

func (r *lfsTokenAuthenticationResult) Reducer() authz.AuthorizationReducer {
	return &authz.SpecificReposAuthorizationReducer{
		ResourceRepos: []authz.RepoGetter{r},
	}
}

func (r *lfsTokenAuthenticationResult) GetTargetRepoID() int64 {
	return r.claims.RepoID
}
