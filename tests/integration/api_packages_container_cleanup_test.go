// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"testing"
	"time"

	packages_model "forgejo.org/models/packages"
	"forgejo.org/models/unittest"
	packages_service "forgejo.org/services/packages"
	packages_cleanup "forgejo.org/services/packages/cleanup"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageContainerCleanup(t *testing.T) {
	// Test fixture data contains three images; one that is a standard single-platform image (v2.0), and two that are a
	// manifest multi-platform image (v1.0).
	//
	// The goal of testing here is to ensure that package cleanup can only remove the tags, and leave the `sha256:*`
	// versions for cleanup once they're dangling by having no tags referencing them.

	testCases := []struct {
		name                     string
		cleanupRule              *packages_model.PackageCleanupRule
		expectedCleanupVersion   []string
		expectedRemainingVersion []string
	}{
		{
			name: "keep single-platform",
			cleanupRule: &packages_model.PackageCleanupRule{
				KeepCount: 1,
			},
			expectedCleanupVersion:   []string{"v1.0"},
			expectedRemainingVersion: []string{"v2.0"},
		},
		{
			name: "keep multi-platform",
			cleanupRule: &packages_model.PackageCleanupRule{
				RemovePattern: "v2\\.0",
			},
			expectedCleanupVersion: []string{"v2.0"},
			expectedRemainingVersion: []string{
				"v1.0",
				"sha256:4759bcf56210784f3b2c2b7438a43a5771c5eda61a11825c80f7b32143ac9c12", // linux/arm64/v8 version of v1.0
				"sha256:a972a01ff13b1cb6a9c49b3c34a283b67b75f441b8ff17ad4833a472f7b0fe08", // linux/amd64 version of v1.0
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			defer unittest.OverrideFixtures("tests/integration/fixtures/TestPackageContainerCleanup")()
			defer tests.PrepareTestEnvWithPackageData(t)()

			// Reduce repetition in the test case array
			tt.cleanupRule.Type = packages_model.TypeContainer
			tt.cleanupRule.OwnerID = 1

			targets, err := packages_cleanup.GetCleanupTargets(t.Context(), tt.cleanupRule, true)
			require.NoError(t, err)

			// Verify we have the expected targets
			targetVersions := make([]string, len(targets))
			for i := range targets {
				targetVersions[i] = targets[i].PackageVersion.LowerVersion
			}
			assert.Len(t, targetVersions, len(tt.expectedCleanupVersion))
			for _, expected := range tt.expectedCleanupVersion {
				assert.Containsf(t, targetVersions, expected, "expected to cleanup %s, but didn't", expected)
			}

			// Delete the target packages
			for _, ct := range targets {
				err := packages_service.DeletePackageVersionAndReferences(t.Context(), ct.PackageVersion)
				require.NoError(t, err)
			}

			// Perform the sha256 cleanup routine
			require.NoError(t, packages_cleanup.CleanupExpiredData(t.Context(), -1*time.Hour))

			// Verify the remaining versions are correct
			remainingVersions, _, err := packages_model.SearchVersions(t.Context(), &packages_model.PackageSearchOptions{
				OwnerID: tt.cleanupRule.OwnerID,
				Type:    tt.cleanupRule.Type,
			})
			require.NoError(t, err)
			remainingVersionsStr := make([]string, len(remainingVersions))
			for i := range remainingVersions {
				remainingVersionsStr[i] = remainingVersions[i].LowerVersion
			}
			assert.Len(t, remainingVersionsStr, len(tt.expectedRemainingVersion))
			for _, expected := range tt.expectedRemainingVersion {
				assert.Contains(t, remainingVersionsStr, expected, "expected to find remaining version %s, but didn't", expected)
			}
		})
	}
}
