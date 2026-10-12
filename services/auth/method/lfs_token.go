// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	"errors"
	"fmt"
	"net/http"

	user_model "forgejo.org/models/user"
	"forgejo.org/modules/lfs"
	"forgejo.org/modules/setting"
	"forgejo.org/services/auth"

	"github.com/golang-jwt/jwt/v5"
)

var _ auth.Method = &LFSToken{}

// LFSToken is an authentication method used when access a git repository over ssh which has LFS resources in-use.  The
// LFS client will issue a `git-lfs-authenticate` command over the ssh connection, and Forgejo will provide a
// supplemental HTTP header "Authorization: Bearer ..." with a JWT.  The LFS client can then make HTTP requests to LFS
// endpoints with that supplemental header in order to inherit the permissions of the SSH user and to retrieve LFS
// objects.
type LFSToken struct{}

func (a *LFSToken) Verify(req *http.Request, w http.ResponseWriter, _ auth.SessionStore) auth.MethodOutput {
	hasToken, tokenText := tokenFromAuthorizationBearer(req).Get()
	if !hasToken {
		return &auth.AuthenticationNotAttempted{}
	}

	token, err := jwt.ParseWithClaims(tokenText, &lfs.Claims{}, func(t *jwt.Token) (any, error) {
		k := setting.LFS.SigningKey
		if t.Method != k.SigningMethod() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return k.VerifyKey(), nil
	})
	if err != nil {
		return &auth.AuthenticationAttemptedIncorrectCredential{Error: err}
	}

	claims, claimsOk := token.Claims.(*lfs.Claims)
	if !token.Valid {
		return &auth.AuthenticationAttemptedIncorrectCredential{Error: errors.New("not a valid LFS JWT")}
	} else if !claimsOk {
		return &auth.AuthenticationError{Error: fmt.Errorf("claim object %v was not an lfs.Claims instance", token.Claims)}
	}

	u, err := user_model.GetUserByID(req.Context(), claims.UserID)
	if err != nil {
		return &auth.AuthenticationError{Error: fmt.Errorf("unable to load claim user %d: %w", claims.UserID, err)}
	}
	if !u.IsAccessAllowed(req.Context()) {
		return &auth.AuthenticationError{Error: fmt.Errorf("user access is not permitted")}
	}

	return &auth.AuthenticationSuccess{
		Result: &lfsTokenAuthenticationResult{
			user:   u,
			claims: claims,
		},
	}
}
