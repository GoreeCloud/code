// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package federation

import (
	"net/http"

	"forgejo.org/modules/log"

	ap "github.com/go-ap/activitypub"
)

func processPersonInboxAccept(activity *ap.Activity) (ServiceResult, error) {
	if activity.Object.GetType() != ap.FollowType {
		log.Error("Invalid object type for Accept activity: %v", activity.Object.GetType())
		return ServiceResult{}, NewErrNotAcceptablef("invalid object type for Accept activity: %v", activity.Object.GetType())
	}

	// We currently do not do anything here, we just drop it.
	return NewServiceResultStatusOnly(http.StatusNoContent), nil
}
