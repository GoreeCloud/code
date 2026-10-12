// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	"fmt"
	"net/http"

	auth_model "forgejo.org/models/auth"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/services/auth"
	"forgejo.org/services/authz"
)

var _ auth.Method = &AccessToken{}

type AccessToken struct {
	// Permit the use of `Authorization: Basic ...` to include an access token.
	PermitBasic bool
	// Permit the use of `Authorization: Bearer ...`, `Authorization: Token ...`, and form-based tokens.
	PermitBearer bool
}

func (a *AccessToken) Verify(req *http.Request, w http.ResponseWriter, _ auth.SessionStore) auth.MethodOutput {
	maybeAuthToken := a.getTokenFromRequest(req)
	if !maybeAuthToken.Has() {
		return &auth.AuthenticationNotAttempted{}
	}
	_, authToken := maybeAuthToken.Get()

	// check personal access token
	token, err := auth_model.GetAccessTokenBySHA(req.Context(), authToken)
	if auth_model.IsErrAccessTokenNotExist(err) || auth_model.IsErrAccessTokenEmpty(err) {
		return &auth.AuthenticationAttemptedIncorrectCredential{Error: err}
	} else if err != nil {
		return &auth.AuthenticationError{Error: fmt.Errorf("access token GetAccessTokenBySHA: %w", err)}
	}

	log.Trace("AccessToken: Valid AccessToken for user[%d]", token.UID)
	u, err := user_model.GetUserByID(req.Context(), token.UID)
	if err != nil {
		return &auth.AuthenticationError{Error: fmt.Errorf("access token GetUserByID: %w", err)}
	}

	if err = token.UpdateLastUsed(req.Context()); err != nil {
		log.Error("UpdateLastUsed:  %v", err)
	}

	reducer, err := authz.GetAuthorizationReducerForAccessToken(req.Context(), token)
	if err != nil {
		return &auth.AuthenticationError{Error: fmt.Errorf("access token GetAuthorizationReducerForAccessToken: %w", err)}
	}

	return &auth.AuthenticationSuccess{Result: &accessTokenAuthenticationResult{user: u, scope: token.Scope, reducer: reducer}}
}

func (a *AccessToken) getTokenFromRequest(req *http.Request) optional.Option[string] {
	if a.PermitBearer {
		if has, token := tokenFromForm(req).Get(); has {
			return optional.Some(token)
		}
		if has, token := tokenFromAuthorizationBearer(req).Get(); has {
			return optional.Some(token)
		}
	}
	if a.PermitBasic {
		if has, token := tokenFromAuthorizationBasic(req).Get(); has {
			return optional.Some(token)
		}
	}
	return optional.None[string]()
}
