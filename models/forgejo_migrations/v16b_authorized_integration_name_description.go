// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"context"
	"fmt"

	"forgejo.org/models/db"
	"forgejo.org/modules/timeutil"

	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add name & description to authorized_integration",
		Upgrade:     addAuthorizedIntegrationNameDescription,
	})
}

func addAuthorizedIntegrationNameDescription(x *xorm.Engine) error {
	type AuthorizedIntegration struct {
		// New fields:
		Name        string
		Description string `xorm:"LONGTEXT"`

		// Existing fields, used for UPDATE in migration:
		ID          int64              `xorm:"pk autoincr"`
		Issuer      string             `xorm:"NOT NULL UNIQUE(s)"`
		Audience    string             `xorm:"NOT NULL UNIQUE(s)"`
		CreatedUnix timeutil.TimeStamp `xorm:"NOT NULL created"`
		// don't include `UpdatedUnix`, so the updated timestamp isn't bumped when Name is set in migration
	}

	_, err := x.SyncWithOptions(
		xorm.SyncOptions{IgnoreDropIndices: true},
		new(AuthorizedIntegration),
	)
	if err != nil {
		return err
	}

	// As v16a has creating this table, v16b will likely have no records for any users.  But for developers working on
	// v16, populate "Name" with a quick computed value:
	return db.Iterate(db.DefaultContext, nil, func(ctx context.Context, ai *AuthorizedIntegration) error {
		ai.Name = fmt.Sprintf("%s created %s", ai.Issuer, ai.CreatedUnix.FormatDate())
		r, err := db.GetEngine(ctx).
			ID(ai.ID).
			Cols("name").
			Update(ai)
		if err != nil {
			return err
		} else if r != 1 {
			return fmt.Errorf("UPDATE expected to affect 1 row, but was %d", r)
		}
		return nil
	})
}
