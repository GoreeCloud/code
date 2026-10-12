// Copyright 2026 The Forgejo Authors
// SPDX-License-Identifier: MIT

package testimport

// ensure the init() function of those modules are called in a test
// environment that may not include them. It matters when the engine
// is trying to figure out the ordering of foreign keys, for instance

import ( //revive:disable:blank-imports
	_ "forgejo.org/models/actions"
	_ "forgejo.org/models/activities"
	_ "forgejo.org/models/auth"
	_ "forgejo.org/models/forgefed"
	_ "forgejo.org/models/perm/access"
	_ "forgejo.org/models/repo"
)
