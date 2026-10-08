// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package federation

import (
	"context"
	"net/http"

	"forgejo.org/models/user"
	"forgejo.org/modules/log"

	ap "github.com/go-ap/activitypub"
)

func processPersonInboxUndo(ctx context.Context, ctxUser *user.User, activity *ap.Activity) (ServiceResult, error) {
	if activity.Object.GetType() != ap.FollowType {
		log.Error("Invalid object type for Undo activity: %v", activity.Object.GetType())
		return ServiceResult{}, NewErrNotAcceptablef("Invalid object type for Undo activity: %v", activity.Object.GetType())
	}

	actorURI := activity.Actor.GetLink().String()
	_, federatedUser, _, err := findFederatedUser(ctx, actorURI)
	if err != nil {
		log.Error("User not found: %v", err)
		return ServiceResult{}, NewErrInternalf("User not found: %v", err)
	}

	if federatedUser != nil {
		following, err := user.IsFollowingAp(ctx, ctxUser, federatedUser)
		if err != nil {
			log.Error("forgefed.IsFollowing: %v", err)
			return ServiceResult{}, NewErrInternalf("forgefed.IsFollowing: %v", err)
		}
		if !following {
			// The local user is not following the federated one, nothing to do.
			log.Trace("Local user[%d] is not following federated user[%d]", ctxUser.ID, federatedUser.ID)
			return NewServiceResultStatusOnly(http.StatusNoContent), nil
		}
		if err := user.RemoveFollower(ctx, ctxUser, federatedUser); err != nil {
			log.Error("Unable to remove follower", err)
			return ServiceResult{}, NewErrInternalf("Unable to remove follower: %v", err)
		}
	}

	return NewServiceResultStatusOnly(http.StatusNoContent), nil
}
