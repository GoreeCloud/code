// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import "xorm.io/builder"

type FindRunJobOptions struct {
	RepoID  int64
	OwnerID int64
}

func (opts FindRunJobOptions) Bad() builder.Cond {
	cond := builder.NewCond()
	if opts.RepoID > 0 {
		cond = cond.And(builder.Eq{"repo_id": opts.RepoID})
	}
	// ruleid:forgejo-logic-suspicious-OwnerID-check
	if opts.OwnerID > 0 {
		cond = cond.And(builder.Eq{"owner_id": opts.OwnerID})
	}
	return cond
}

func (opts FindRunJobOptions) Good() builder.Cond {
	cond := builder.NewCond()
	if opts.RepoID > 0 {
		cond = cond.And(builder.Eq{"repo_id": opts.RepoID})
	}
	// ok:forgejo-logic-suspicious-OwnerID-check
	if opts.OwnerID != 0 {
		cond = cond.And(builder.Eq{"owner_id": opts.OwnerID})
	}
	return cond
}
