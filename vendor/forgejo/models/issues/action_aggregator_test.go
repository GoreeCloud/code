// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package issues

import (
	"testing"

	"forgejo.org/models/db"
	org_model "forgejo.org/models/organization"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"

	"github.com/stretchr/testify/assert"
)

func TestRequestReviewTarget(t *testing.T) {
	unittest.PrepareTestEnv(t)

	target := RequestReviewTarget{User: &user_model.User{ID: 1, Name: "user1"}}
	assert.Equal(t, int64(1), target.ID())
	assert.Equal(t, "user1", target.Name())
	assert.Equal(t, "user", target.Type())
	assert.Equal(t, "/user1", target.Link(db.DefaultContext))

	target = RequestReviewTarget{Team: &org_model.Team{ID: 2, Name: "Collaborators", OrgID: 3}}
	assert.Equal(t, int64(2), target.ID())
	assert.Equal(t, "Collaborators", target.Name())
	assert.Equal(t, "team", target.Type())
	assert.Equal(t, "/org/org3/teams/Collaborators", target.Link(db.DefaultContext))

	target = RequestReviewTarget{Team: org_model.NewGhostTeam()}
	assert.Equal(t, int64(-1), target.ID())
	assert.Equal(t, "Ghost team", target.Name())
	assert.Equal(t, "team", target.Type())
	assert.Empty(t, target.Link(db.DefaultContext))
}
