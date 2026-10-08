// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"context"
	"fmt"
	"time"

	"forgejo.org/models/db"
	"forgejo.org/modules/timeutil"

	"xorm.io/builder"
)

type ActionUser struct {
	ID     int64 `xorm:"pk autoincr"`
	UserID int64 `xorm:"INDEX UNIQUE(action_user_index) REFERENCES(user, id)"`
	RepoID int64 `xorm:"INDEX UNIQUE(action_user_index) REFERENCES(repository, id)"`

	TrustedWithPullRequests bool

	LastAccess timeutil.TimeStamp `xorm:"INDEX"`
}

func init() {
	db.RegisterModel(new(ActionUser))
}

type ErrUserNotExist struct {
	UserID int64
	RepoID int64
}

func IsErrUserNotExist(err error) bool {
	_, ok := err.(ErrUserNotExist)
	return ok
}

func (err ErrUserNotExist) Error() string {
	return fmt.Sprintf("ActionUser does not exist [user_id: %d, repo_id: %d]", err.UserID, err.RepoID)
}

func InsertActionUser(ctx context.Context, user *ActionUser) error {
	user.LastAccess = timeutil.TimeStampNow()
	return db.Insert(ctx, user)
}

func DeleteActionUserByUserIDAndRepoID(ctx context.Context, userID, repoID int64) error {
	_, err := db.GetEngine(ctx).Table(&ActionUser{}).Where("user_id=? AND repo_id=?", userID, repoID).Delete()
	return err
}

var updateFrequency = 24 * time.Hour

func MaybeUpdateAccess(ctx context.Context, user *ActionUser) error {
	// Keep track of the last time the record was accessed to identify which one
	// are never accessed so they can be removed eventually. But only every updateFrequency
	// to not stress the underlying database.
	if timeutil.TimeStampNow() > user.LastAccess.AddDuration(updateFrequency) {
		user.LastAccess = timeutil.TimeStampNow()
		if _, err := db.GetEngine(ctx).ID(user.ID).Cols("last_access").Update(user); err != nil {
			return err
		}
	}

	return nil
}

func GetActionUserByUserIDAndRepoID(ctx context.Context, userID, repoID int64) (*ActionUser, error) {
	user := new(ActionUser)
	has, err := db.GetEngine(ctx).Where("user_id=? AND repo_id=?", userID, repoID).Get(user)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrUserNotExist{userID, repoID}
	}
	return user, nil
}

func GetActionUserByUserIDAndRepoIDAndUpdateAccess(ctx context.Context, userID, repoID int64) (*ActionUser, error) {
	user, err := GetActionUserByUserIDAndRepoID(ctx, userID, repoID)
	if err != nil {
		return nil, err
	}
	return user, MaybeUpdateAccess(ctx, user)
}

var expire = 3 * 30 * 24 * time.Hour

func RevokeInactiveActionUser(ctx context.Context) error {
	olderThan := timeutil.TimeStampNow().AddDuration(-expire)

	_, err := db.GetEngine(ctx).Where(builder.Lt{"last_access": olderThan}).Delete(&ActionUser{})
	return err
}
