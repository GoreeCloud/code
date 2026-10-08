// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"time"

	packages_model "forgejo.org/models/packages"
	container_model "forgejo.org/models/packages/container"
	"forgejo.org/modules/optional"
	packages_service "forgejo.org/services/packages"

	digest "github.com/opencontainers/go-digest"
)

// Cleanup removes expired container data
func Cleanup(ctx context.Context, olderThan time.Duration) error {
	if err := cleanupExpiredBlobUploads(ctx, olderThan); err != nil {
		return err
	}
	if err := CleanupSHA256(ctx, olderThan); err != nil {
		return err
	}
	return cleanupExpiredUploadedBlobs(ctx, olderThan)
}

// cleanupExpiredBlobUploads removes expired blob uploads
func cleanupExpiredBlobUploads(ctx context.Context, olderThan time.Duration) error {
	pbus, err := packages_model.FindExpiredBlobUploads(ctx, olderThan)
	if err != nil {
		return err
	}

	for _, pbu := range pbus {
		if err := RemoveBlobUploadByID(ctx, pbu.ID); err != nil {
			return err
		}
	}

	return nil
}

// cleanupExpiredUploadedBlobs removes expired uploaded blobs not referenced by a manifest
func cleanupExpiredUploadedBlobs(ctx context.Context, olderThan time.Duration) error {
	pfs, err := container_model.SearchExpiredUploadedBlobs(ctx, olderThan)
	if err != nil {
		return err
	}

	for _, pf := range pfs {
		if err := packages_service.DeletePackageFile(ctx, pf); err != nil {
			return err
		}
	}

	pvs, _, err := packages_model.SearchVersions(ctx, &packages_model.PackageSearchOptions{
		Type: packages_model.TypeContainer,
		Version: packages_model.SearchValue{
			ExactMatch: true,
			Value:      container_model.UploadVersion,
		},
		IsInternal: optional.Some(true),
		HasFiles:   optional.Some(false),
	})
	if err != nil {
		return err
	}

	for _, pv := range pvs {
		if err := packages_model.DeleteAllProperties(ctx, packages_model.PropertyTypeVersion, pv.ID); err != nil {
			return err
		}

		if err := packages_model.DeleteVersionByID(ctx, pv.ID); err != nil {
			return err
		}
	}

	return nil
}

func ShouldBeSkipped(pv *packages_model.PackageVersion) bool {
	// Always skip the "latest" tag
	if pv.LowerVersion == "latest" {
		return true
	}

	// Check if the version is a digest (or untagged)
	if digest.Digest(pv.LowerVersion).Validate() == nil {
		// Don't apply PackageCleanupRules to image versions that aren't tags; rely on `CleanupSHA256` to do mass
		// cleanup of them once they're dangling due to not being referenced.
		return true
	}

	return false
}
