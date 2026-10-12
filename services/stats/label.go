// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package stats

import "context"

// Queue a recalculation of the stats on a `Label` for a given label by its ID
func QueueRecalcLabelByID(ctx context.Context, labelID int64) {
	safePush(ctx, recalcRequest{
		RecalcType: LabelByLabelID,
		ObjectID:   labelID,
	})
}

// Queue a recalculation of the stats on all `Label` in a given repository
func QueueRecalcLabelByRepoID(ctx context.Context, repoID int64) {
	safePush(ctx, recalcRequest{
		RecalcType: LabelByRepoID,
		ObjectID:   repoID,
	})
}
