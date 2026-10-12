// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package templates_test

import (
	"context"
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/markup"
)

func TestMain(m *testing.M) {
	markup.Init(&markup.ProcessorHelper{
		IsUsernameMentionable: func(ctx context.Context, username string) bool {
			return username == "mention-user"
		},
	})
	unittest.MainTest(m)
}
