// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	auth_model "forgejo.org/models/auth"
	user_model "forgejo.org/models/user"
)

func TokenRequiresRepoOwnerScope(ctx Context, owner *user_model.User, requiredScopeLevel auth_model.AccessTokenScopeLevel) {
	var category auth_model.AccessTokenScopeCategory
	if owner.IsOrganization() {
		category = auth_model.AccessTokenScopeCategoryOrganization
	} else {
		category = auth_model.AccessTokenScopeCategoryUser
	}
	TokenRequiresScopes(ctx, []auth_model.AccessTokenScopeCategory{category}, requiredScopeLevel)
}
