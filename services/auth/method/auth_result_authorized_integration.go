// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	auth_model "forgejo.org/models/auth"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"
	"forgejo.org/services/auth"
	"forgejo.org/services/authz"
)

var _ auth.AuthenticationResult = &authorizedIntegrationAuthenticationResult{}

type authorizedIntegrationAuthenticationResult struct {
	*auth.BaseAuthenticationResult
	user      *user_model.User
	scope     auth_model.AccessTokenScope
	reducer   authz.AuthorizationReducer
	expiresAt optional.Option[timeutil.TimeStamp]
}

func (r *authorizedIntegrationAuthenticationResult) User() *user_model.User {
	return r.user
}

func (r *authorizedIntegrationAuthenticationResult) Scope() optional.Option[auth_model.AccessTokenScope] {
	return optional.Some(r.scope)
}

func (r *authorizedIntegrationAuthenticationResult) Reducer() authz.AuthorizationReducer {
	return r.reducer
}

func (r *authorizedIntegrationAuthenticationResult) ExpiresAt() optional.Option[timeutil.TimeStamp] {
	return r.expiresAt
}
