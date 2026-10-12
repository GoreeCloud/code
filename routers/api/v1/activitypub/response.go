// Copyright 2023 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package activitypub

import (
	"net/http"

	"forgejo.org/modules/activitypub"
	"forgejo.org/modules/forgefed"
	"forgejo.org/modules/log"
	"forgejo.org/services/context"
	"forgejo.org/services/federation"

	ap "github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
)

// Respond with an ActivityStreams object
func responseServiceResult(ctx *context.APIContext, result federation.ServiceResult) {
	ctx.Resp.Header().Add("Content-Type", activitypub.ActivityStreamsContentType)

	switch {
	case result.StatusOnly():
		ctx.Resp.WriteHeader(result.HTTPStatus)
		return
	case result.WithBytes():
		ctx.Resp.WriteHeader(result.HTTPStatus)
		if _, err := ctx.Resp.Write(result.Bytes); err != nil {
			log.Error("Error writing a response: %v", err)
			ctx.Error(http.StatusInternalServerError, "Error writing a response", err)
			return
		}
	case result.WithActivity():
		binary, err := jsonld.WithContext(
			jsonld.IRI(ap.ActivityBaseURI),
			jsonld.IRI(ap.SecurityContextURI),
			jsonld.IRI(forgefed.ForgeFedNamespaceURI),
		).Marshal(result.Activity)
		if err != nil {
			ctx.ServerError("Marshal", err)
			return
		}
		ctx.Resp.WriteHeader(result.HTTPStatus)
		if _, err = ctx.Resp.Write(binary); err != nil {
			log.Error("write to resp err: %v", err)
		}
	}
}

// Respond with an ActivityStreams object
// Deprecated
func response(ctx *context.APIContext, v any) {
	binary, err := jsonld.WithContext(
		jsonld.IRI(ap.ActivityBaseURI),
		jsonld.IRI(ap.SecurityContextURI),
		jsonld.IRI(forgefed.ForgeFedNamespaceURI),
	).Marshal(v)
	if err != nil {
		ctx.ServerError("Marshal", err)
		return
	}

	ctx.Resp.Header().Add("Content-Type", activitypub.ActivityStreamsContentType)
	ctx.Resp.WriteHeader(http.StatusOK)
	if _, err = ctx.Resp.Write(binary); err != nil {
		log.Error("write to resp err: %v", err)
	}
}
