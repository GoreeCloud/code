// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package stats

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"forgejo.org/modules/optional"
	"forgejo.org/modules/timeutil"
)

type recalcRequest struct {
	RecalcType      RecalcType
	ObjectID        int64
	UpdateTimestamp optional.Option[timeutil.TimeStamp]
}

func (r *recalcRequest) string() string {
	return fmt.Sprintf("recalcRequest:%d:%d:%d", r.RecalcType, r.ObjectID, r.UpdateTimestamp.ValueOrDefault(0))
}

func recalcRequestFromString(s string) (*recalcRequest, error) {
	tags := strings.Split(s, ":")
	if len(tags) != 4 {
		return nil, errors.New("expected three tags")
	} else if tags[0] != "recalcRequest" {
		return nil, fmt.Errorf("expected tag `recalcRequest`, but was %s", tags[0])
	}
	recalcType, err := strconv.ParseInt(tags[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to parse recalc type: %w", err)
	}
	objectID, err := strconv.ParseInt(tags[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse object ID: %w", err)
	}
	timestamp, err := strconv.ParseInt(tags[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse timestamp ID: %w", err)
	}
	var updateTimestamp optional.Option[timeutil.TimeStamp]
	if timestamp == 0 {
		updateTimestamp = optional.None[timeutil.TimeStamp]()
	} else {
		updateTimestamp = optional.Some(timeutil.TimeStamp(timestamp))
	}
	return &recalcRequest{
		RecalcType:      RecalcType(recalcType),
		ObjectID:        objectID,
		UpdateTimestamp: updateTimestamp,
	}, nil
}
