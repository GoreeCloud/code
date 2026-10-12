// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package container

import (
	"testing"

	packages_model "forgejo.org/models/packages"

	"github.com/stretchr/testify/assert"
)

func TestShouldBeSkipped(t *testing.T) {
	tests := []struct {
		name    string
		pv      *packages_model.PackageVersion
		skipped bool
	}{
		{
			name:    "latest: never cleaned up",
			pv:      &packages_model.PackageVersion{LowerVersion: "latest"},
			skipped: true,
		},
		{
			name:    "sha256:* never cleaned up",
			pv:      &packages_model.PackageVersion{LowerVersion: "sha256:8a4d01effd20bcc5c65857885013efdbbdd466ac0a5a26b3fac573095533e3fc"},
			skipped: true,
		},
		{
			name:    "sha256:* never cleaned up",
			pv:      &packages_model.PackageVersion{LowerVersion: "v1.0.0"},
			skipped: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipped := ShouldBeSkipped(tt.pv)
			assert.Equal(t, tt.skipped, skipped)
		})
	}
}
