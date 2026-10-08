// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package federation

import (
	"context"
	"net/http"

	"forgejo.org/models/activities"
	"forgejo.org/models/user"
	fm "forgejo.org/modules/forgefed"
	"forgejo.org/modules/log"

	ap "github.com/go-ap/activitypub"
)

func processPersonInboxCreate(ctx context.Context, user *user.User, activity *ap.Activity) (ServiceResult, error) {
	createAct, err := fm.NewForgeUserActivityFromAp(*activity)
	if err != nil {
		log.Error("Invalid user activity: %v, %v", activity, err)
		return ServiceResult{}, NewErrNotAcceptablef("Invalid user activity: %v", err)
	}

	actorURI := createAct.Actor.GetLink().String()
	federatedBaseUser, _, _, err := findFederatedUser(ctx, actorURI)
	if err != nil {
		log.Error("Federated user not found (%s): %v", actorURI, err)
		return ServiceResult{}, NewErrNotAcceptablef("federated user not found (%s): %v", actorURI, err)
	}
	if federatedBaseUser == nil {
		log.Error("Federated user not found (%s): %v", actorURI, err)
		return ServiceResult{}, NewErrNotAcceptablef("federated user not found (%s): %v", actorURI, err)
	}

	federatedUserActivity, err := activities.NewFederatedUserActivity(
		user.ID,
		federatedBaseUser.ID,
		createAct.Actor.GetLink().String(),
		createAct.Note.Content.String(),
		createAct.Note.URL.GetID().String(),
		*activity,
	)
	if err != nil {
		log.Error("Error creating federatedUserActivity (%s): %v", actorURI, err)
		return ServiceResult{}, NewErrNotAcceptablef("Error creating federatedUserActivity: %v", err)
	}

	if err := activities.CreateUserActivity(ctx, &federatedUserActivity); err != nil {
		log.Error("Unable to record activity: %v", err)
		return ServiceResult{}, NewErrNotAcceptablef("Unable to record activity: %v", err)
	}

	return NewServiceResultStatusOnly(http.StatusNoContent), nil
}
