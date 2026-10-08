// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package authz

import (
	repo_model "forgejo.org/models/repo"
)

// Defines an API for reducing available permissions to specific resources.  Typically associated with a fine-grained
// access tokens and provides methods to reduce authorization that the access token provides down to specific resources.
type AuthorizationReducer interface {
	// Incorporate all the methods of [RepositoryAuthorizationReducer], which allows reducing permissions related to
	// repositories specifically.
	repo_model.RepositoryAuthorizationReducer

	// Controls whether the presence of an authorization reducer will prevent administrators from overriding permission
	// checks. Typically site administrators and repo administrators are exempted from permission checks, but if an
	// authorization reducer is present then it may be intended for its restrictions to apply even to administrators.
	//
	// `true` allows the typical case where administrators *can* override permissions. `false` disables administrator
	// overrides of permission checks.
	AllowAdminOverride() bool
}
