// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add ui to authorized_integration",
		Upgrade:     addAuthorizedIntegrationUI,
	})
}

func addAuthorizedIntegrationUI(x *xorm.Engine) error {
	type AuthorizedIntegration struct {
		UI string `xorm:"NOT NULL default('generic')"`
	}

	_, err := x.SyncWithOptions(
		xorm.SyncOptions{IgnoreDropIndices: true},
		new(AuthorizedIntegration),
	)
	return err
}
