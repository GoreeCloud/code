// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package a

import (
	"errors"

	"forgejo.org/services/context"
)

func work() {}

func directApiCallFine(ctx *context.APIContext) {
	ctx.Error(500, "title", nil)
}

// Directly call APIContext functions, then "do work", triggering a linting error:

func directApiCallError(ctx *context.APIContext) {
	ctx.Error(500, "title", nil) // want "Invocation of (.*) / Error, and control flow continues afterwards."
	work()
}

func directApiCallInternalServerError(ctx *context.APIContext) {
	ctx.InternalServerError(errors.New("something")) // want "Invocation of (.*) / InternalServerError, and control flow continues afterwards."
	work()
}

func directApiCallNotFound(ctx *context.APIContext) {
	ctx.NotFound("title") // want "Invocation of (.*) / NotFound, and control flow continues afterwards."
	work()
}

func directApiCallNotFoundOrServerError(ctx *context.APIContext) {
	ctx.NotFoundOrServerError("logMsg", func(err error) bool { return false }, errors.New("something")) // want "Invocation of (.*) / NotFoundOrServerError, and control flow continues afterwards."
	work()
}

func directApiCallServerError(ctx *context.APIContext) {
	ctx.ServerError("something", errors.New("something")) // want "Invocation of (.*) / ServerError, and control flow continues afterwards."
	work()
}

// Call methods on ctx that will go to the `*Base` implementation:

func indirectApiCallJSON(ctx *context.APIContext) {
	ctx.JSON(200, "something") // want "Invocation of (.*).Base / JSON, and control flow continues afterwards."
	work()
}

func indirectApiCallPlainText(ctx *context.APIContext) {
	ctx.PlainText(200, "something") // want "Invocation of (.*).Base / PlainText, and control flow continues afterwards."
	work()
}

func indirectApiCallPlainTextBytes(ctx *context.APIContext) {
	ctx.PlainTextBytes(200, []byte{}) // want "Invocation of (.*).Base / PlainTextBytes, and control flow continues afterwards."
	work()
}

func indirectApiCallRedirect(ctx *context.APIContext) {
	ctx.Redirect("/somewhere") // want "Invocation of (.*).Base / Redirect, and control flow continues afterwards."
	work()
}

func indirectApiCallServeContent(ctx *context.APIContext) {
	ctx.ServeContent(nil, nil) // want "Invocation of (.*).Base / ServeContent, and control flow continues afterwards."
	work()
}
