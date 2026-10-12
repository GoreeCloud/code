// Copyright 2023, 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package activitypub

import (
	"fmt"
	"net/http"
	"strings"

	"forgejo.org/modules/activitypub"
	"forgejo.org/modules/forgefed"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/web"
	"forgejo.org/services/context"
	"forgejo.org/services/federation"

	ap "github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
)

// Repository function returns the Repository actor for a repo
func Repository(ctx *context.APIContext) {
	// swagger:operation GET /activitypub/repository-id/{repository-id} activitypub activitypubRepository
	// ---
	// summary: Returns the Repository actor for a repo
	// produces:
	// - application/json
	// parameters:
	// - name: repository-id
	//   in: path
	//   description: repository ID of the repo
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/ActivityPub"

	link := fmt.Sprintf("%s/api/v1/activitypub/repository-id/%d", strings.TrimSuffix(setting.AppURL, "/"), ctx.Repo().Repository.ID)
	repo := forgefed.RepositoryNew(ap.IRI(link))

	repo.Inbox = ap.IRI(link + "/inbox")
	repo.Outbox = ap.IRI(link + "/outbox")

	repo.Name = ap.NaturalLanguageValuesNew()
	err := repo.Name.Set(ap.NilLangRef, ap.Content(ctx.Repo().Repository.Name))
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "Set Name", err)
		return
	}
	response(ctx, repo)
}

// PersonInbox function handles the incoming data for a repository inbox
func RepositoryInbox(ctx *context.APIContext) {
	// swagger:operation POST /activitypub/repository-id/{repository-id}/inbox activitypub activitypubRepositoryInbox
	// ---
	// summary: Send to the inbox
	// produces:
	// - application/json
	// parameters:
	// - name: repository-id
	//   in: path
	//   description: repository ID of the repo
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/ForgeLike"
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"

	repository := ctx.Repo().Repository
	log.Info("RepositoryInbox: repo: %v", repository)
	form := web.GetForm(ctx)
	activity := form.(*ap.Activity)
	result, err := federation.ProcessRepositoryInbox(ctx, activity, repository.ID)
	if err != nil {
		ctx.Error(federation.HTTPStatus(err), "Processing Repository Inbox failed", result)
		return
	}
	responseServiceResult(ctx, result)
}

func RepositoryOutbox(ctx *context.APIContext) {
	// swagger:operation POST /activitypub/repository-id/{repository-id}/outbox activitypub activitypubRepositoryOutbox
	// ---
	// summary: Display the outbox
	// produces:
	// - application/ld+json
	// parameters:
	// - name: repository-id
	//   in: path
	//   description: repository ID of the repo
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Outbox"

	repository := ctx.Repo().Repository
	outbox := ap.OrderedCollectionNew(ap.IRI(repository.APActorID() + "/outbox"))

	binary, err := jsonld.WithContext(
		jsonld.IRI(ap.ActivityBaseURI),
	).Marshal(outbox)
	if err != nil {
		ctx.ServerError("MarshalJSON", err)
		return
	}

	ctx.Resp.Header().Add("Content-Type", activitypub.ActivityStreamsContentType)
	ctx.Resp.WriteHeader(http.StatusOK)

	_, err = ctx.Resp.Write(binary)
	if err != nil {
		log.Error("write to resp err: %s", err)
	}
}
