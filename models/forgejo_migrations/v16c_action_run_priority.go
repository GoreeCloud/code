// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add priority action_run",
		Upgrade:     addActionRunPriority,
	})
}

func addActionRunPriority(x *xorm.Engine) error {
	type ActionRun struct {
		Priority   int8 `xorm:"NOT NULL DEFAULT 0"`
		Prioritize bool `xorm:"NOT NULL DEFAULT false"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(ActionRun))
	return err
}
