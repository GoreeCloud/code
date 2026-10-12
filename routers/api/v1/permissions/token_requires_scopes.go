// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"fmt"
	"net/http"

	auth_model "forgejo.org/models/auth"
)

func TokenRequiresScopes(ctx Context, requiredScopeCategories []auth_model.AccessTokenScopeCategory, requiredScopeLevel auth_model.AccessTokenScopeLevel) {
	// no scope required
	if len(requiredScopeCategories) == 0 {
		return
	}

	// Need OAuth2 token to be present.
	hasScope, scope := ctx.Authentication().Scope().Get()
	if !hasScope {
		return
	}

	// get the required scope for the given access level and category
	requiredScopes := auth_model.GetRequiredScopes(requiredScopeLevel, requiredScopeCategories...)
	allow, err := scope.HasScope(requiredScopes...)
	if err != nil {
		ctx.Error(http.StatusForbidden, "tokenRequiresScope", "checking scope failed: "+err.Error())
		return
	}

	if !allow {
		ctx.Error(http.StatusForbidden, "tokenRequiresScope", fmt.Sprintf("token does not have at least one of required scope(s): %v", requiredScopes))
		return
	}

	ctx.SetRequiredScopeCategories(requiredScopeCategories)
}
