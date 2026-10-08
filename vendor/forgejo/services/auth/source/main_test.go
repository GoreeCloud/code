// Copyright 2025 The Forgejo Contributors. All rights reserved.
// SPDX-License-Identifier: MIT

package source

import (
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/setting"
	"forgejo.org/services/webhook"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m, &unittest.TestOptions{
		SetUp: func() error {
			setting.LoadQueueSettings()
			return webhook.Init()
		},
	})
}
