// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package org

import (
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/organization"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_InviteTeamMember(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	defer test.MockVariableValue(&setting.Service.AddMembersByInvitations, true)()

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{ID: 2})
	invited := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})
	inviter := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	require.NoError(t, InviteOrAddTeamMember(db.DefaultContext, inviter, invited, team))

	unittest.AssertExistsAndLoadBean(t, &organization.TeamInvite{TeamID: team.ID, InviterID: inviter.ID, InvitedID: optional.Some(invited.ID)})
	isMember, err := organization.IsTeamMember(db.DefaultContext, team.OrgID, team.ID, invited.ID)
	require.NoError(t, err)
	assert.False(t, isMember)
}

func Test_AddTeamMember(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	defer test.MockVariableValue(&setting.Service.AddMembersByInvitations, false)()

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{ID: 2})
	invited := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 5})
	inviter := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	require.NoError(t, InviteOrAddTeamMember(db.DefaultContext, inviter, invited, team))

	unittest.AssertNotExistsBean(t, &organization.TeamInvite{TeamID: team.ID, InviterID: inviter.ID, InvitedID: optional.Some(invited.ID)})
	isMember, err := organization.IsTeamMember(db.DefaultContext, team.OrgID, team.ID, invited.ID)
	require.NoError(t, err)
	assert.True(t, isMember)
}

func Test_SelfInvite(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	defer test.MockVariableValue(&setting.Service.AddMembersByInvitations, true)()

	team := unittest.AssertExistsAndLoadBean(t, &organization.Team{ID: 2})
	inviter := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	isMember, err := organization.IsTeamMember(db.DefaultContext, team.OrgID, team.ID, inviter.ID)
	require.NoError(t, err)
	assert.False(t, isMember)

	// the inviter invites themselves to a team (they are owner of the org)
	require.NoError(t, InviteOrAddTeamMember(db.DefaultContext, inviter, inviter, team))

	// even though `ADD_MEMBERS_BY_INVITATIONS` is on, we don't generate an invite to the inviter,
	// they are added to the team directly
	unittest.AssertNotExistsBean(t, &organization.TeamInvite{TeamID: team.ID, InviterID: inviter.ID, InvitedID: optional.Some(inviter.ID)})
	isMember, err = organization.IsTeamMember(db.DefaultContext, team.OrgID, team.ID, inviter.ID)
	require.NoError(t, err)
	assert.True(t, isMember)
}
