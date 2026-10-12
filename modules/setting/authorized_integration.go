// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import "time"

var AuthorizedIntegration = struct {
	AllowedDomains     string
	BlockedDomains     string
	AllowLocalNetworks bool
	RequestTimeout     time.Duration
	CacheTTL           time.Duration
}{}

func loadAuthorizedIntegrationFrom(rootCfg ConfigProvider) {
	sec := rootCfg.Section("authorized_integration")
	AuthorizedIntegration.AllowedDomains = sec.Key("ALLOWED_DOMAINS").MustString("")
	AuthorizedIntegration.BlockedDomains = sec.Key("BLOCKED_DOMAINS").MustString("")
	AuthorizedIntegration.AllowLocalNetworks = sec.Key("ALLOW_LOCALNETWORKS").MustBool(false)
	AuthorizedIntegration.RequestTimeout = sec.Key("REQUEST_TIMEOUT").MustDuration(10 * time.Second)
	AuthorizedIntegration.CacheTTL = sec.Key("CACHE_TTL").MustDuration(10 * time.Minute)
}
