// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swagger

import (
	api "forgejo.org/modules/structs"
)

// OAuth2Application
// swagger:response OAuth2Application
type swaggerResponseOAuth2Application struct {
	// in:body
	Body api.OAuth2Application `json:"body"`
}

// AccessToken represents an API access token.
// swagger:response AccessToken
type swaggerResponseAccessToken struct {
	// in:body
	Body api.AccessToken `json:"body"`
}

// AccessTokenList
// swagger:response AccessTokenList
type swaggerResponseAccessTokenList struct {
	// in:body
	Body []api.AccessToken `json:"body"`

	// The total number of access tokens
	TotalCount int64 `json:"X-Total-Count"`
}

// OAuth2ApplicationList
// swagger:response OAuth2ApplicationList
type swaggerResponseOAuth2ApplicationList struct {
	// in:body
	Body []api.OAuth2Application `json:"body"`

	// The total number of OAuth2 applications
	TotalCount int64 `json:"X-Total-Count"`
}
