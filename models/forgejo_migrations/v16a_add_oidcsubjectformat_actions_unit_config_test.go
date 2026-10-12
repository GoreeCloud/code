// Copyright 2026 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"testing"

	"forgejo.org/models/db"
	migration_tests "forgejo.org/models/gitea_migrations/test"
	"forgejo.org/modules/timeutil"

	"code.forgejo.org/xorm/xorm/convert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_setOIDCSubjectFormatLegacy15(t *testing.T) {
	type Type int
	type UnitAccessMode int
	type RepoUnit struct { //revive:disable-line:exported
		ID                 int64
		RepoID             int64              `xorm:"INDEX(s)"`
		Type               Type               `xorm:"INDEX(s)"`
		Config             convert.Conversion `xorm:"TEXT"`
		CreatedUnix        timeutil.TimeStamp `xorm:"INDEX CREATED"`
		DefaultPermissions UnitAccessMode     `xorm:"NOT NULL DEFAULT 0"`
	}
	x, deferable := migration_tests.PrepareTestEnv(t, 0, new(RepoUnit))
	defer deferable()
	if x == nil || t.Failed() {
		return
	}

	require.NoError(t, setOIDCSubjectFormatLegacy15(x))

	var records []map[string]string
	require.NoError(t,
		db.GetEngine(t.Context()).
			Table("repo_unit").
			Select("`id`, `repo_id`, `config`").
			OrderBy("`id`").
			Find(&records))
	assert.Equal(t, []map[string]string{
		{
			"config":  "{\"OIDCSubjectFormat\":\"legacy-forgejo-v15\"}",
			"id":      "1",
			"repo_id": "4",
		},
		{
			"config":  "{\"DisabledWorkflows\":[\"renovate.yml\"],\"OIDCSubjectFormat\":\"legacy-forgejo-v15\"}",
			"id":      "2",
			"repo_id": "5",
		},
		{
			"config":  "",
			"id":      "3",
			"repo_id": "6",
		},
	}, records)
}
