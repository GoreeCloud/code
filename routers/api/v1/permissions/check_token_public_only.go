// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"net/http"

	auth_model "forgejo.org/models/auth"
	user_model "forgejo.org/models/user"
	api "forgejo.org/modules/structs"
)

func CheckTokenPublicOnly(ctx Context, user, org, packageOwner *user_model.User) {
	if !ctx.PublicOnly() {
		return
	}

	requiredScopeCategories := ctx.RequiredScopeCategories()
	if len(requiredScopeCategories) == 0 {
		return
	}

	// public Only permission check
	switch {
	case auth_model.ContainsCategory(requiredScopeCategories, auth_model.AccessTokenScopeCategoryRepository):
		if ctx.Repository() != nil && ctx.Repository().IsPrivate {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public repos")
			return
		}
	case auth_model.ContainsCategory(requiredScopeCategories, auth_model.AccessTokenScopeCategoryIssue):
		if ctx.Repository() != nil && ctx.Repository().IsPrivate {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public issues")
			return
		}
	case auth_model.ContainsCategory(requiredScopeCategories, auth_model.AccessTokenScopeCategoryOrganization):
		if org != nil && org.Visibility != api.VisibleTypePublic {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public orgs")
			return
		}
		if user != nil && user.IsOrganization() && user.Visibility != api.VisibleTypePublic {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public orgs")
			return
		}
	case auth_model.ContainsCategory(requiredScopeCategories, auth_model.AccessTokenScopeCategoryUser):
		if user != nil && user.IsUser() && user.Visibility != api.VisibleTypePublic {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public users")
			return
		}
	case auth_model.ContainsCategory(requiredScopeCategories, auth_model.AccessTokenScopeCategoryActivityPub):
		if user != nil && user.IsUser() && user.Visibility != api.VisibleTypePublic {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public activitypub")
			return
		}
	case auth_model.ContainsCategory(requiredScopeCategories, auth_model.AccessTokenScopeCategoryNotification):
		if ctx.Repository() != nil && ctx.Repository().IsPrivate {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public notifications")
			return
		}
	case auth_model.ContainsCategory(requiredScopeCategories, auth_model.AccessTokenScopeCategoryPackage):
		if packageOwner != nil && packageOwner.Visibility.IsPrivate() {
			ctx.Error(http.StatusForbidden, "reqToken", "token scope is limited to public packages")
			return
		}
	}
}
