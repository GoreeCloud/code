// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package a

import (
	"errors"

	"forgejo.org/services/context"
)

func directWebCallFine(ctx *context.Context) {
	ctx.HTML(200, "tmpl")
}

// Directly call Context functions, then "do work", triggering a linting error:

func directWebCallHTML(ctx *context.Context) {
	ctx.HTML(200, "tmpl") // want "Invocation of (.*) / HTML, and control flow continues afterwards."
	work()
}

func directWebCallJSONError(ctx *context.Context) {
	ctx.JSONError("something") // want "Invocation of (.*) / JSONError, and control flow continues afterwards."
	work()
}

func directWebCallJSONOK(ctx *context.Context) {
	ctx.JSONOK() // want "Invocation of (.*) / JSONOK, and control flow continues afterwards."
	work()
}

func directWebCallJSONRedirect(ctx *context.Context) {
	ctx.JSONRedirect("/somewhere") // want "Invocation of (.*) / JSONRedirect, and control flow continues afterwards."
	work()
}

func directWebCallJSONTemplate(ctx *context.Context) {
	ctx.JSONTemplate("tmpl") // want "Invocation of (.*) / JSONTemplate, and control flow continues afterwards."
	work()
}

func directWebCallNotFound(ctx *context.Context) {
	ctx.NotFound("something", errors.New("something")) // want "Invocation of (.*) / NotFound, and control flow continues afterwards."
	work()
}

func directWebCallNotFoundOrServerError(ctx *context.Context) {
	ctx.NotFoundOrServerError("something", func(err error) bool { return false }, errors.New("something")) // want "Invocation of (.*) / NotFoundOrServerError, and control flow continues afterwards."
	work()
}

func directWebCallRedirectToFirst(ctx *context.Context) {
	ctx.RedirectToFirst("/somewhere") // want "Invocation of (.*) / RedirectToFirst, and control flow continues afterwards."
	work()
}

func directWebCallRenderWithErr(ctx *context.Context) {
	ctx.RenderWithErr("something", "tmpl", errors.New("something")) // want "Invocation of (.*) / RenderWithErr, and control flow continues afterwards."
	work()
}

func directWebCallServerError(ctx *context.Context) {
	ctx.ServerError("something", errors.New("something")) // want "Invocation of (.*) / ServerError, and control flow continues afterwards."
	work()
}

// Call methods on ctx that will go to the `*Base` implementation:

func indirectWebCallError(ctx *context.Context) {
	ctx.Error(500, "something") // want "Invocation of (.*).Base / Error, and control flow continues afterwards."
	work()
}

func indirectWebCallJSON(ctx *context.Context) {
	ctx.JSON(200, "something") // want "Invocation of (.*).Base / JSON, and control flow continues afterwards."
	work()
}

func indirectWebCallPlainText(ctx *context.Context) {
	ctx.PlainText(200, "something") // want "Invocation of (.*).Base / PlainText, and control flow continues afterwards."
	work()
}

func indirectWebCallPlainTextBytes(ctx *context.Context) {
	ctx.PlainTextBytes(200, []byte{}) // want "Invocation of (.*).Base / PlainTextBytes, and control flow continues afterwards."
	work()
}

func indirectWebCallRedirect(ctx *context.Context) {
	ctx.Redirect("/somewhere") // want "Invocation of (.*).Base / Redirect, and control flow continues afterwards."
	work()
}

func indirectWebCallServeContent(ctx *context.Context) {
	ctx.ServeContent(nil, nil) // want "Invocation of (.*).Base / ServeContent, and control flow continues afterwards."
	work()
}
