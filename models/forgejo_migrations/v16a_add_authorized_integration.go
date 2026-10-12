// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"forgejo.org/modules/timeutil"

	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add authorized_integration tables",
		Upgrade:     addAuthorizedIntegrationTables,
	})
}

func addAuthorizedIntegrationTables(x *xorm.Engine) error {
	type ClaimRules struct{}
	type AuthorizedIntegration struct {
		ID               int64              `xorm:"pk autoincr"`
		UserID           int64              `xorm:"NOT NULL REFERENCES(user, id)"`
		Scope            string             `xorm:"NOT NULL"`
		ResourceAllRepos bool               `xorm:"NOT NULL"`
		Issuer           string             `xorm:"NOT NULL UNIQUE(s)"`
		Audience         string             `xorm:"NOT NULL UNIQUE(s)"`
		ClaimRules       *ClaimRules        `xorm:"NOT NULL JSON"`
		CreatedUnix      timeutil.TimeStamp `xorm:"NOT NULL created"`
		UpdatedUnix      timeutil.TimeStamp `xorm:"NOT NULL updated"`
	}
	type AuthorizedIntegResourceRepo struct {
		ID          int64              `xorm:"pk autoincr"`
		IntegID     int64              `xorm:"NOT NULL REFERENCES(authorized_integration, id)"`
		RepoID      int64              `xorm:"NOT NULL REFERENCES(repository, id)"`
		CreatedUnix timeutil.TimeStamp `xorm:"created NOT NULL"`
	}

	_, err := x.SyncWithOptions(
		xorm.SyncOptions{IgnoreDropIndices: true},
		new(AuthorizedIntegration),
		new(AuthorizedIntegResourceRepo),
	)
	return err
}
