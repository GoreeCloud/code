// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"testing"

	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestWindowConfig(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	resp := MakeRequest(t, NewRequest(t, "GET", "/"), http.StatusOK)
	assert.Contains(t, resp.Body.String(), `customEmojis: new Set(["git","gitea","codeberg","gitlab","github","gogs","forgejo"]),`)
}
