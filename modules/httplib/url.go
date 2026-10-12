// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package httplib

import (
	"net/url"
	"path"
	"strings"

	"forgejo.org/modules/setting"
)

// Unfortunately browsers consider a redirect Location with preceding "//", "\\", "/\" and "\/" as meaning redirect to "http(s)://REST_OF_PATH"
// Therefore we should ignore these redirect locations to prevent open redirects.
func isBrowserRedirect(s string) bool {
	return len(s) > 1 && (s[0] == '/' || s[0] == '\\') && (s[1] == '/' || s[1] == '\\')
}

// IsRiskyRedirectURL returns true if the URL is considered risky for redirects
func IsRiskyRedirectURL(s string) bool {
	if isBrowserRedirect(s) {
		return true
	}

	u, err := url.Parse(s)
	if err != nil || ((u.Scheme != "" || u.Host != "") && !strings.HasPrefix(strings.ToLower(s), strings.ToLower(setting.AppURL))) {
		return true
	}

	// If the path contains `..` then it's still possible this is seen
	// as a browser redirect, use `path.Clean` to eliminate each inner `..`
	// and then check if that might be a browser redirect.
	if strings.Contains(u.Path, "..") {
		return isBrowserRedirect(path.Clean(u.Path))
	}

	return false
}
