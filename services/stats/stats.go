// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// Package stats implements a queue and registration model for recalculating statistics in the database asynchronously.
// Typically the statistics are simple counts of related objects which are used for later database sort operations --
// because of the use of sorting and pagination when querying, these data are not possible to convert into efficient
// real-time queries. The reasons that these calculations are performed asynchronously through a queue are:
//
// - User operations that are common and performance-sensitive don't have to wait for recalculations that don't need to
// be exactly up-to-date at all times.
//
// - Database deadlocks that can occur between concurrent operations -- for example, if you were holding a lock on an
// issue while recalculating a label's count of open issues -- can be broken by making the recalculation occur outside
// of the transaction.
//
// There are two elements to using the package; either you are requesting recalculations, or you are implementing
// statistics.
//
// If you're requesting recalculations, each object type has simple queue wrapper methods like `QueueRecalcLabelByID`,
// which are fire-and-forget operations that will make a best-effort to recalculate the requested statistic, but
// provides no guarantee on when.
//
// If you're implementing recalculations, then a new `RecalcType` enum value needs to be added and simple wrapper
// methods in the `stats` package, and then use the `RegisterRecalc` method implement the recalculation in your model
// package.
//
// The implementation of stats is currently simple, but may be enhanced (as needed) in the future with:
//
// - Bulk recalculations -- gather all the recalc requests of the same objects and perform them in one operation, which
// is typically more efficient for a database.
//
// - Retry operations -- if a recalculation fails, assume that it may be a transient failure and allow it to be retried
// soon.  If it continues to fail persistenly, fall back to logging errors.
//
// - Throttling and fairness -- in the event of a queue backup, don't allow available resources to be consumed entirely
// by single users.
package stats

import (
	"context"
	"errors"
	"time"

	"forgejo.org/models/db"
	"forgejo.org/modules/graceful"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/queue"
	"forgejo.org/modules/timeutil"
)

type RecalcType int

const (
	LabelByLabelID RecalcType = iota
	LabelByRepoID
	MilestoneByMilestoneID
)

type RecalcHandler func(context.Context, int64, optional.Option[timeutil.TimeStamp]) error

var (
	// string queue is used for consistent unique behaviour independent of json serialization
	statsQueue       *queue.WorkerPoolQueue[string]
	recalcHandlers   = make(map[RecalcType]RecalcHandler)
	recalcTimeout    = 1 * time.Minute
	testFlushTimeout = 30 * time.Second
)

// Initialize the stats queue
func Init() error {
	statsQueue = queue.CreateUniqueQueue(graceful.GetManager().ShutdownContext(), "stats_recalc", handler)
	if statsQueue == nil {
		return errors.New("unable to create stats queue")
	}
	go graceful.GetManager().RunWithCancel(statsQueue)
	return nil
}

// Register that a specific type of recalculation will be performed by the given handler.  Can only be performed once
// per recalc type.
func RegisterRecalc(recalcType RecalcType, handler RecalcHandler) {
	_, present := recalcHandlers[recalcType]
	if present {
		log.Fatal("RegisterRecalc invoked twice for RecalcType %d", recalcType)
	}
	recalcHandlers[recalcType] = handler
}

func handler(items ...string) []string {
	ctx, cancel := context.WithTimeout(graceful.GetManager().ShutdownContext(), recalcTimeout)
	defer cancel()

	for _, item := range items {
		req, err := recalcRequestFromString(item)
		if err != nil {
			log.Error("Unable to parse recalc request, ignoring: %v", err)
			continue
		}

		handler, ok := recalcHandlers[req.RecalcType]
		if !ok {
			log.Error("Unrecognized RecalcType %d, ignoring", req.RecalcType)
			continue
		}
		if err := handler(ctx, req.ObjectID, req.UpdateTimestamp); err != nil {
			log.Error("Error in stats recalc %v on object %d: %v", req.RecalcType, req.ObjectID, err)
		}
	}
	return nil
}

func safePush(ctx context.Context, recalc recalcRequest) {
	db.AfterTx(ctx, func() {
		err := statsQueue.Push(recalc.string())
		if err != nil && !errors.Is(err, queue.ErrAlreadyInQueue) {
			log.Error("error during stat queue push: %v", err)
		}
	})
}

// Only use for testing; do not use in production code
func Flush(ctx context.Context) error {
	return statsQueue.FlushWithContext(ctx, testFlushTimeout)
}
