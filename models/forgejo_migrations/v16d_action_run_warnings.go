// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add pre_execution_warning fields to action_run",
		Upgrade:     addActionRunWarnings,
	})
}

func addActionRunWarnings(x *xorm.Engine) error {
	type PreExecutionWarning int64
	type ActionRun struct {
		PreExecutionWarningCodes   []PreExecutionWarning
		PreExecutionWarningDetails [][]any `xorm:"JSON LONGTEXT"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(ActionRun))
	return err
}
