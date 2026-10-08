// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"context"

	"forgejo.org/models/db"
	"forgejo.org/modules/timeutil"
)

// Represents a many-to-many join table which indicates specific repositories (RepoID) that can be accessed by an access
// token (TokenID).  An access token's ResourceAllRepos field must be false for records in this table to become active.
type AccessTokenResourceRepo struct {
	ID      int64 `xorm:"pk autoincr"`
	TokenID int64 `xorm:"NOT NULL REFERENCES(access_token, id)"` // needs to be shortened from "AccessTokenID" for the index to fit MySQL table identifier length restrictions
	RepoID  int64 `xorm:"NOT NULL REFERENCES(repository, id)"`

	CreatedUnix timeutil.TimeStamp `xorm:"created NOT NULL"`
}

func init() {
	db.RegisterModel(new(AccessTokenResourceRepo))
}

func GetRepositoriesAccessibleWithToken(ctx context.Context, accessTokenID int64) ([]*AccessTokenResourceRepo, error) {
	var resources []*AccessTokenResourceRepo
	err := db.GetEngine(ctx).
		Where("token_id = ?", accessTokenID).
		Find(&resources)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func GetRepositoriesAccessibleWithTokens(ctx context.Context, accessTokens []*AccessToken) (map[int64][]*AccessTokenResourceRepo, error) {
	accessTokenIDs := make([]int64, len(accessTokens))
	for i, at := range accessTokens {
		accessTokenIDs[i] = at.ID
	}

	var resources []*AccessTokenResourceRepo
	err := db.GetEngine(ctx).
		In("token_id", accessTokenIDs).
		Find(&resources)
	if err != nil {
		return nil, err
	}
	retval := make(map[int64][]*AccessTokenResourceRepo)
	for _, resource := range resources {
		retval[resource.TokenID] = append(retval[resource.TokenID], resource)
	}
	return retval, nil
}

func InsertAccessTokenResourceRepos(ctx context.Context, accessTokenID int64, resources []*AccessTokenResourceRepo) error {
	return db.WithTx(ctx, func(ctx context.Context) error {
		for _, resourceRepo := range resources {
			resourceRepo.TokenID = accessTokenID
			if err := db.Insert(ctx, resourceRepo); err != nil {
				return err
			}
		}
		return nil
	})
}
