// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"forgejo.org/modules/optional"

	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add login_source_id column to forgejo_auth_token",
		Upgrade:     addLoginSourceIDToForgejoAuthToken,
	})
}

func addLoginSourceIDToForgejoAuthToken(x *xorm.Engine) error {
	type ForgejoAuthToken struct {
		LoginSourceID optional.Option[int64] `xorm:"INDEX REFERENCES(login_source, id)"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(ForgejoAuthToken))
	return err
}
