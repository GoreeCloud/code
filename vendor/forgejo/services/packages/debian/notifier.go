// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package debian

import (
	"context"

	packages_model "forgejo.org/models/packages"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/log"
	debian_module "forgejo.org/modules/packages/debian"
	"forgejo.org/services/notify"
)

func init() {
	notify.RegisterNotifier(&debianPackageNotifier{})
}

type debianPackageNotifier struct {
	notify.NullNotifier
}

func (m *debianPackageNotifier) PackageDelete(ctx context.Context, doer *user_model.User, pd *packages_model.PackageDescriptor) {
	rebuildFromPackageEvent(ctx, pd)
}

func rebuildFromPackageEvent(ctx context.Context, pd *packages_model.PackageDescriptor) {
	if pd.Package == nil || pd.Package.Type != packages_model.TypeDebian {
		return
	}

	pv, err := GetOrCreateRepositoryVersion(ctx, pd.Owner.ID)
	if err != nil {
		log.Error("GetOrCreateRepositoryVersion failed with error: %v", err)
		return
	}

	for _, file := range pd.Files {
		distribution := file.Properties.GetByName(debian_module.PropertyDistribution)
		component := file.Properties.GetByName(debian_module.PropertyComponent)
		architecture := file.Properties.GetByName(debian_module.PropertyArchitecture)

		if distribution == "" || component == "" || architecture == "" {
			log.Warn("Debian package had file missing all expected properties; distribution = %q, component = %q, architecture = %q", distribution, component, architecture)
			return
		}

		log.Trace("Debian package change triggered rebuild of repository index for distribution = %q, component = %q, architecture = %q", distribution, component, architecture)
		err = buildRepositoryFiles(ctx, pd.Owner.ID, pv, distribution, component, architecture)
		if err != nil {
			log.Error("buildRepositoryFiles failed with error: %v", err)
		}
	}
}
