// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgejo_migrations

import (
	"forgejo.org/modules/optional"

	"code.forgejo.org/xorm/xorm"
)

func init() {
	registerMigration(&Migration{
		Description: "add invited_id to team_invite",
		Upgrade:     addTeamInviteInvitedID,
	})
}

func addTeamInviteInvitedID(x *xorm.Engine) error {
	// the invited_id is set None if we are inviting someone who does not have an account yet
	type TeamInvite struct {
		InvitedID optional.Option[int64] `xorm:"index REFERENCES(user, id)"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{IgnoreDropIndices: true}, new(TeamInvite))
	return err
}
