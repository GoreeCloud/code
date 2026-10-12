// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package a

import (
	"errors"
	"html/template"
	"math/rand/v2"

	"forgejo.org/services/context"
)

func controlFlowIf(ctx *context.APIContext) {
	if rand.Float64() > 0.1 {
		ctx.Error(500, "title", nil) // want "Invocation of (.*) and control flow continues afterwards."
	}

	work()

	if rand.Float64() > 0.1 {
		ctx.Error(500, "title", nil) // no want
		return
	}

	work()

	if rand.Float64() > 0.1 {
		ctx.Error(500, "title", nil) // no want
		return
	} else {
		ctx.Error(500, "title", nil) // want "Invocation of (.*) and control flow continues afterwards."
	}

	if rand.Float64() > 0.1 {
		ctx.Error(500, "title", nil) // no want
		return
	} else if rand.Float64() > 0.1 {
		ctx.Error(500, "title", nil) // no want
		return
	} else if rand.Float64() > 0.1 {
		ctx.InternalServerError(errors.New("something")) // want "Invocation of (.*) and control flow continues afterwards."
	} else {
		ctx.Error(500, "title", nil) // no want
		return
	}

	work()

	if rand.Float64() > 0.1 {
		ctx.Error(500, "title", nil) // no want -- method ends either way
	} else {
		ctx.Error(500, "title", nil) // no want -- method ends either way
	}
}

func controlFlowSwitch(ctx *context.APIContext) {
	switch {
	case rand.Float64() > 0.1:
		ctx.Error(500, "title", nil) // want "Invocation of (.*) and control flow continues afterwards."
	case rand.Float64() > 0.1:
		ctx.Error(500, "title", nil) // want "Invocation of (.*) and control flow continues afterwards."
	case rand.Float64() > 0.1:
		ctx.Error(500, "title", nil) // no want
		return
	}

	work()

	switch {
	case rand.Float64() > 0.1:
		ctx.Error(500, "title", nil) // want "Invocation of (.*) and control flow continues afterwards."
		fallthrough
	case rand.Float64() > 0.1:
		ctx.Error(500, "title", nil) // no want
		return
	}

	work()

	switch {
	case rand.Float64() > 0.1:
		ctx.Error(500, "title", nil) // no want -- method ends either way
	case rand.Float64() > 0.1:
		ctx.Error(500, "title", nil) // no want -- method ends either way
	}
}

func controlFlowLoop(ctx *context.APIContext) {
	for range []int{1, 2, 3} {
		ctx.Error(500, "title", nil) // want "Invocation of (.*) and control flow continues afterwards."
	}

	for range []int{1, 2, 3} {
		ctx.Error(500, "title", nil)
		return
	}

	for range []int{1, 2, 3} {
		ctx.Error(500, "title", nil)
		break
	}
	return
}

func controlFlowInternalDecl(ctx *context.Context) {
	work()
	renderWithError := func(msg template.HTML) {
		ctx.RenderWithErr(msg, "tplAccessTokenEdit", nil) // no want -- within the context of `renderWithError` this is fine
	}
	if rand.Float64() > 0.1 {
		renderWithError("")
		return
	}
	if rand.Float64() > 0.1 {
		// Future: would love if call to another method which calls ctx.*, followed by no `return`, could cause a
		// diagnostic -- but that may require a multi-pass analysis to do a good job generally.  With a local function
		// declaration it's more feasible but it's also not very common, may not be worth that effort.
		renderWithError("")
	}
	work()
}

var localFunc = func(ctx *context.Context) {
	ctx.PlainText(200, "title") // want "Invocation of (.*) / PlainText, and control flow continues afterwards."
	work()
}
