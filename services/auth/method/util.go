// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	"net/http"
	"strings"

	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/util"
)

func tokenFromForm(req *http.Request) optional.Option[string] {
	_ = req.ParseForm()
	if !setting.DisableQueryAuthToken {
		// Check token.
		if token := req.Form.Get("token"); token != "" {
			return optional.Some(token)
		}
		// Check access token.
		if token := req.Form.Get("access_token"); token != "" {
			return optional.Some(token)
		}
	} else if req.Form.Get("token") != "" || req.Form.Get("access_token") != "" {
		log.Warn("API token sent in query string but DISABLE_QUERY_AUTH_TOKEN=true")
	}
	return optional.None[string]()
}

func tokenFromAuthorizationBearer(req *http.Request) optional.Option[string] {
	authorization := req.Header.Get("Authorization")
	if len(authorization) != 0 {
		auths := strings.Fields(authorization)
		if len(auths) == 2 && (util.ASCIIEqualFold(auths[0], "token") || util.ASCIIEqualFold(auths[0], "bearer")) {
			return optional.Some(auths[1])
		}
	}
	return optional.None[string]()
}

func tokenFromAuthorizationBasic(req *http.Request) optional.Option[string] {
	authorization := req.Header.Get("Authorization")
	if len(authorization) != 0 {
		auths := strings.SplitN(authorization, " ", 2)
		if len(auths) == 2 && strings.ToLower(auths[0]) == "basic" {
			uname, passwd, err := base.BasicAuthDecode(auths[1])
			if err != nil {
				// Client provided a `Authorization: Basic ...`, but then [...] either couldn't be base64 decoded, or
				// didn't contain a ":" for username/password separation.  If we return `None`, it'll indicate to the
				// caller that `Authorization: Basic [...]` wasn't present and skip authentication, so intead we'll
				// return Some with an empty token to trigger a 401 error.
				log.Debug("unexpected error in BasicAuthDecode(%q): %s", auths[1], err)
				return optional.Some("")
			}

			// Usually we'll use the password as the access token (or oauth token), but if the password is empty or
			// `x-oauth-basic`, we'll use the username as a token.  This behaviour is inherited from GitHub's OAuth Git
			// over HTTPS behaviour.
			if len(passwd) == 0 || passwd == "x-oauth-basic" {
				return optional.Some(uname)
			}

			return optional.Some(passwd)
		}
	}
	return optional.None[string]()
}
