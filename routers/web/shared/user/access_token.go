// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package user

import (
	"errors"

	"forgejo.org/modules/optional"
	"forgejo.org/services/authz"
	"forgejo.org/services/context"
)

// Translate errors from [ValidateAccessToken] into user-visible messages, if the error is expected as a user-facing
// error.  None[string] may be returned to indicate the error is an unexpected server-side error, not a user validation
// error.
func TranslateAccessTokenValidationError(ctx *context.Base, err error) optional.Option[string] {
	switch {
	case errors.Is(err, authz.ErrSpecifiedReposNone):
		return optional.Some[string](ctx.Locale.TrString("access_token.error.specified_repos_none"))
	case errors.Is(err, authz.ErrSpecifiedReposNoPublicOnly):
		return optional.Some[string](ctx.Locale.TrString("access_token.error.specified_repos_and_public_only"))
	case errors.Is(err, authz.ErrSpecifiedReposInvalidScope):
		return optional.Some[string](ctx.Locale.TrString("access_token.error.specified_repos_and_invalid_scope"))
	default:
		return optional.None[string]()
	}
}
