// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package federation

import (
	"context"
	"fmt"
	"net/http"

	"forgejo.org/models/user"
	"forgejo.org/modules/forgefed"
	"forgejo.org/modules/log"

	ap "github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
)

func processPersonFollow(ctx context.Context, ctxUser *user.User, activity *ap.Activity) (ServiceResult, error) {
	follow, err := forgefed.NewForgeFollowFromAp(*activity)
	if err != nil {
		log.Error("Invalid follow activity: %s", err)
		return ServiceResult{}, NewErrNotAcceptablef("Invalid follow activity: %v", err)
	}

	actorURI := follow.Actor.GetLink().String()
	_, federatedUser, federationHost, err := FindOrCreateFederatedUser(ctx, actorURI)
	if err != nil {
		log.Error("Error finding or creating federated user (%s): %v", actorURI, err)
		return ServiceResult{}, NewErrNotAcceptablef("Federated user not found: %v", err)
	}

	following, err := user.IsFollowingAp(ctx, ctxUser, federatedUser)
	if err != nil {
		log.Error("forgefed.IsFollowing: %v", err)
		return ServiceResult{}, NewErrNotAcceptablef("forgefed.IsFollowing: %v", err)
	}
	if following {
		// If the user is already following, we're good, nothing to do.
		log.Trace("Local user[%d] is already following federated user[%d]", ctxUser.ID, federatedUser.ID)
		return NewServiceResultStatusOnly(http.StatusNoContent), nil
	}

	follower, err := user.AddFollower(ctx, ctxUser, federatedUser)
	if err != nil {
		log.Error("Unable to add follower: %v", err)
		return ServiceResult{}, NewErrNotAcceptablef("Unable to add follower: %v", err)
	}

	accept := ap.AcceptNew(ap.IRI(fmt.Sprintf(
		"%s#accepts/follow/%d", ctxUser.APActorID(), follower.ID,
	)), follow)
	accept.Actor = ap.IRI(ctxUser.APActorID())
	payload, err := jsonld.WithContext(jsonld.IRI(ap.ActivityBaseURI)).Marshal(accept)
	if err != nil {
		log.Error("Unable to Marshal JSON: %v", err)
		return ServiceResult{}, NewErrInternalf("MarshalJSON: %v", err)
	}

	hostURL := federationHost.AsURL()
	if err := deliveryQueue.Push(deliveryQueueItem{
		InboxURL: hostURL.JoinPath(federatedUser.InboxPath).String(),
		Doer:     ctxUser,
		Payload:  payload,
	}); err != nil {
		log.Error("Unable to push to pending queue: %v", err)
		return ServiceResult{}, NewErrInternalf("Unable to push to pending queue: %v", err)
	}

	// Respond back with an accept
	result := NewServiceResultWithBytes(http.StatusAccepted, []byte(`{"status":"Accepted"}`))
	return result, nil
}
