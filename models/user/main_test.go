// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user_test

import (
	"testing"

	"forgejo.org/models/unittest"

	_ "forgejo.org/modules/testimport"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
