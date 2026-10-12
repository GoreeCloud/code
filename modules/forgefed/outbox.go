// Copyright 2024, 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgefed

import (
	ap "github.com/go-ap/activitypub"
)

// ActivityStream OrderedCollection of activities
// swagger:model
type ForgeOutbox struct {
	// swagger:ignore
	ap.OutboxStream
}
