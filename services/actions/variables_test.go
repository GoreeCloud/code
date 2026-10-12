// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicesAction_envNameCIRegexMatch(t *testing.T) {
	require.ErrorContains(t, envNameCIRegexMatch("ci"), "cannot be ci")
	require.ErrorContains(t, envNameCIRegexMatch("CI"), "cannot be ci")
	assert.NoError(t, envNameCIRegexMatch("CI_SOMETHING"))
}
