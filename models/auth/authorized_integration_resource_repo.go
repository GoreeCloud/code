// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"context"

	"forgejo.org/models/db"
	"forgejo.org/modules/timeutil"
)

// Represents a many-to-many join table which indicates specific repositories (RepoID) that can be accessed by an
// authorized integration (IntegID).  An authorized integrations's ResourceAllRepos field must be false for records in
// this table to become active.
//
// Model name is shortend (from AuthorizedIntegrationResourceRepo) to accomodate recreate-tables + MySQL, where the
// "tmp_recreate_" + foreign key index name would exceed the max identifier length.
type AuthorizedIntegResourceRepo struct {
	ID      int64 `xorm:"pk autoincr"`
	IntegID int64 `xorm:"NOT NULL REFERENCES(authorized_integration, id)"` // field name shortened (AuthorizationIntegrationID) for max identifier length
	RepoID  int64 `xorm:"NOT NULL REFERENCES(repository, id)"`

	CreatedUnix timeutil.TimeStamp `xorm:"created NOT NULL"`
}

func init() {
	db.RegisterModel(new(AuthorizedIntegResourceRepo))
}

func (air *AuthorizedIntegResourceRepo) GetTargetRepoID() int64 {
	return air.RepoID
}

func GetRepositoriesAccessibleWithIntegration(ctx context.Context, aiID int64) ([]*AuthorizedIntegResourceRepo, error) {
	var resources []*AuthorizedIntegResourceRepo
	err := db.GetEngine(ctx).
		Where("integ_id = ?", aiID).
		Find(&resources)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func InsertAuthorizedIntegrationResourceRepos(ctx context.Context, aiID int64, resources []*AuthorizedIntegResourceRepo) error {
	return db.WithTx(ctx, func(ctx context.Context) error {
		for _, resourceRepo := range resources {
			resourceRepo.IntegID = aiID
			if err := db.Insert(ctx, resourceRepo); err != nil {
				return err
			}
		}
		return nil
	})
}

func UpdateAuthorizedIntegrationResourceRepos(ctx context.Context, aiID int64, resources []*AuthorizedIntegResourceRepo) error {
	return db.WithTx(ctx, func(ctx context.Context) error {
		if _, err := db.GetEngine(ctx).Delete(&AuthorizedIntegResourceRepo{IntegID: aiID}); err != nil {
			return err
		}
		return InsertAuthorizedIntegrationResourceRepos(ctx, aiID, resources)
	})
}
