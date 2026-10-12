// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import "code.forgejo.org/xorm/xorm"

func init() {
	registerMigration(&Migration{
		Description: "add extra_lines_count to comment for multi-line review comments",
		Upgrade:     addCommentExtraLinesCount,
	})
}

func addCommentExtraLinesCount(x *xorm.Engine) error {
	type Comment struct {
		ExtraLinesCount int64 `xorm:"NOT NULL DEFAULT 0"`
	}

	_, err := x.SyncWithOptions(
		xorm.SyncOptions{IgnoreDropIndices: true},
		new(Comment),
	)
	return err
}
