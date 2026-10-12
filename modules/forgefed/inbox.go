// Copyright 2024, 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgefed

import (
	ap "github.com/go-ap/activitypub"
)

// ForgeFollow activity data type
// swagger:model
type ForgeInbox struct {
	// swagger:ignore
	ap.InboxStream
}
