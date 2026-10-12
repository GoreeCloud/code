// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package permissions

import (
	"context"

	auth_model "forgejo.org/models/auth"
	org_model "forgejo.org/models/organization"
	"forgejo.org/models/perm"
	access_model "forgejo.org/models/perm/access"
	repo_model "forgejo.org/models/repo"
	user_model "forgejo.org/models/user"
	"forgejo.org/services/auth"
	"forgejo.org/services/authz"
)

type Context interface {
	Context() context.Context

	Repository() *repo_model.Repository

	Doer() *user_model.User

	User() *user_model.User

	Organization() *org_model.Organization

	Team() *org_model.Team

	PackageOwner() *user_model.User
	PackageAccessMode() perm.AccessMode

	Permission() *access_model.Permission
	SetPermission(*access_model.Permission)

	IsSigned() bool

	PublicOnly() bool
	SetPublicOnly(bool)

	Reducer() authz.AuthorizationReducer
	SetReducer(authz.AuthorizationReducer)

	Authentication() auth.AuthenticationResult

	RequiredScopeCategories() []auth_model.AccessTokenScopeCategory
	SetRequiredScopeCategories([]auth_model.AccessTokenScopeCategory)

	Error(status int, title string, obj any)
	InternalServerError(err error)
	NotFound(objs ...any)

	GetError() error
	WrittenStatus() int
}
