// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth_test

import (
	"testing"

	"forgejo.org/models/unittest"

	_ "forgejo.org/modules/testimport"
)

func TestMain(m *testing.M) {
	unittest.MainTest(m)
}
