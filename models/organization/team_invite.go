// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package organization

import (
	"context"
	"fmt"

	"forgejo.org/models/db"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/util"

	"xorm.io/builder"
)

type ErrTeamInviteAlreadyExist struct {
	TeamID        int64
	Email         string
	InvitedUserID int64
}

func IsErrTeamInviteAlreadyExist(err error) bool {
	_, ok := err.(ErrTeamInviteAlreadyExist)
	return ok
}

func (err ErrTeamInviteAlreadyExist) Error() string {
	return fmt.Sprintf("team invite already exists [team_id: %d, email: %s, invited_user_id: %d]", err.TeamID, err.Email, err.InvitedUserID)
}

func (err ErrTeamInviteAlreadyExist) Unwrap() error {
	return util.ErrAlreadyExist
}

type ErrTeamInviteNotFound struct {
	Token string
}

func IsErrTeamInviteNotFound(err error) bool {
	_, ok := err.(ErrTeamInviteNotFound)
	return ok
}

func (err ErrTeamInviteNotFound) Error() string {
	return fmt.Sprintf("team invite was not found [token: %s]", err.Token)
}

func (err ErrTeamInviteNotFound) Unwrap() error {
	return util.ErrNotExist
}

// ErrInvitedUserAlreadyAdded indicates that a user is already part of a team and can not be invited again.
type ErrInvitedUserAlreadyAdded struct {
	Email         string
	InvitedUserID optional.Option[int64]
}

// IsErrUserEmailAlreadyAdded checks if an error is a ErrUserEmailAlreadyAdded.
func IsErrUserEmailAlreadyAdded(err error) bool {
	_, ok := err.(ErrInvitedUserAlreadyAdded)
	return ok
}

func (err ErrInvitedUserAlreadyAdded) Error() string {
	return fmt.Sprintf("user with email already added [email: %s, invited_user_id: %d]", err.Email, err.InvitedUserID)
}

func (err ErrInvitedUserAlreadyAdded) Unwrap() error {
	return util.ErrAlreadyExist
}

// TeamInvite represents an invite to a team
type TeamInvite struct {
	ID          int64                  `xorm:"pk autoincr"`
	Token       string                 `xorm:"UNIQUE(token) INDEX NOT NULL DEFAULT ''"`
	InviterID   int64                  `xorm:"NOT NULL DEFAULT 0"`
	OrgID       int64                  `xorm:"INDEX NOT NULL DEFAULT 0"`
	TeamID      int64                  `xorm:"UNIQUE(team_mail) INDEX NOT NULL DEFAULT 0"`
	Email       string                 `xorm:"UNIQUE(team_mail) NOT NULL DEFAULT ''"`
	InvitedID   optional.Option[int64] `xorm:"index REFERENCES(user, id)"`
	InvitedUser *user_model.User       `xorm:"-"`
	CreatedUnix timeutil.TimeStamp     `xorm:"INDEX created"`
	UpdatedUnix timeutil.TimeStamp     `xorm:"INDEX updated"`
}

// CreateTeamInviteByEmail creates a TeamInvite for someone who does not have an account yet.
func CreateTeamInviteByEmail(ctx context.Context, doer *user_model.User, team *Team, email string) (*TeamInvite, error) {
	has, err := db.GetEngine(ctx).Exist(&TeamInvite{
		TeamID: team.ID,
		Email:  email,
	})
	if err != nil {
		return nil, err
	}
	if has {
		return nil, ErrTeamInviteAlreadyExist{
			TeamID: team.ID,
			Email:  email,
		}
	}

	// check if the user is already a team member by email
	exist, err := db.GetEngine(ctx).
		Where(builder.Eq{
			"team_user.org_id":  team.OrgID,
			"team_user.team_id": team.ID,
			"`user`.email":      email,
		}).
		Join("INNER", "`user`", "`user`.id = team_user.uid").
		Table("team_user").
		Exist()
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, ErrInvitedUserAlreadyAdded{
			Email: email,
		}
	}

	token := util.CryptoRandomString(util.RandomStringMedium)

	invite := &TeamInvite{
		Token:     token,
		InviterID: doer.ID,
		OrgID:     team.OrgID,
		TeamID:    team.ID,
		Email:     email,
	}

	return invite, db.Insert(ctx, invite)
}

// CreateTeamInviteForUser creates a TeamInvite for someone who already has an account on the instance.
func CreateTeamInviteForUser(ctx context.Context, doer, invited *user_model.User, team *Team) (*TeamInvite, error) {
	has, err := db.GetEngine(ctx).Exist(&TeamInvite{
		TeamID:    team.ID,
		InvitedID: optional.Some(invited.ID),
	})
	if err != nil {
		return nil, err
	}
	if has {
		return nil, ErrTeamInviteAlreadyExist{
			TeamID: team.ID,
			Email:  invited.Email,
		}
	}

	// check if the user is already a team member
	exist, err := db.GetEngine(ctx).
		Where(builder.Eq{
			"org_id":  team.OrgID,
			"team_id": team.ID,
			"uid":     invited.ID,
		}).
		Table("team_user").
		Exist()
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, ErrInvitedUserAlreadyAdded{
			InvitedUserID: optional.Some(invited.ID),
		}
	}

	token := util.CryptoRandomString(util.RandomStringMedium)

	invite := &TeamInvite{
		Token:       token,
		InviterID:   doer.ID,
		OrgID:       team.OrgID,
		TeamID:      team.ID,
		Email:       invited.Email,
		InvitedID:   optional.Some(invited.ID),
		InvitedUser: invited,
	}

	return invite, db.Insert(ctx, invite)
}

func RemoveInviteByID(ctx context.Context, inviteID, teamID int64) error {
	_, err := db.DeleteByBean(ctx, &TeamInvite{
		ID:     inviteID,
		TeamID: teamID,
	})
	return err
}

func GetInvitesByTeamID(ctx context.Context, teamID int64) ([]*TeamInvite, error) {
	invites := make([]*TeamInvite, 0, 10)
	return invites, db.GetEngine(ctx).
		Where("team_id=?", teamID).
		Find(&invites)
}

func GetInviteByToken(ctx context.Context, token string) (*TeamInvite, error) {
	invite := &TeamInvite{}

	has, err := db.GetEngine(ctx).Where("token=?", token).Get(invite)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrTeamInviteNotFound{Token: token}
	}
	return invite, nil
}

func (i *TeamInvite) LoadInvitedUser(ctx context.Context) error {
	if i.InvitedUser == nil {
		hasInvitedUser, userID := i.InvitedID.Get()
		if hasInvitedUser {
			user, err := user_model.GetUserByID(ctx, userID)
			if err != nil {
				return err
			}
			i.InvitedUser = user
		}
	}
	return nil
}
