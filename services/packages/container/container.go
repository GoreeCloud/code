// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"errors"
	"os"

	packages_model "forgejo.org/models/packages"
	container_model "forgejo.org/models/packages/container"
	"forgejo.org/modules/log"
	packages_module "forgejo.org/modules/packages"
	"forgejo.org/modules/util"
)

// FIXME: Workaround to be removed in v1.20
// https://github.com/go-gitea/gitea/issues/19586
func WorkaroundGetContainerBlob(ctx context.Context, opts *container_model.BlobSearchOptions) (*packages_model.PackageFileDescriptor, error) {
	blob, err := container_model.GetContainerBlob(ctx, opts)
	if err != nil {
		return nil, err
	}

	err = packages_module.NewContentStore().Has(packages_module.BlobHash256Key(blob.Blob.HashSHA256))
	if err != nil {
		if errors.Is(err, util.ErrNotExist) || errors.Is(err, os.ErrNotExist) {
			log.Debug("Package registry inconsistent: blob %s does not exist on file system", blob.Blob.HashSHA256)
			return nil, container_model.ErrContainerBlobNotExist
		}
		return nil, err
	}

	return blob, nil
}
