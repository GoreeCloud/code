// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forms

import (
	"net/http"

	"forgejo.org/modules/web/middleware"
	"forgejo.org/services/context"

	"code.forgejo.org/go-chi/binding"
)

// CreateRunnerForm needs to be filled in by the user to create a new runner.
type CreateRunnerForm struct {
	RunnerName        string `binding:"Required;MaxSize(255)"`
	RunnerDescription string
}

// Validate validates the submitted form.
func (f *CreateRunnerForm) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	ctx := context.GetValidateContext(req)
	return middleware.Validate(errs, ctx.Data, f, ctx.Locale)
}

// EditRunnerForm can be filled in by the user to change an existing runner.
type EditRunnerForm struct {
	RunnerName        string `binding:"Required;MaxSize(255)"`
	RunnerDescription string
	RegenerateToken   bool
}

// Validate validates the submitted form.
func (f *EditRunnerForm) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	ctx := context.GetValidateContext(req)
	return middleware.Validate(errs, ctx.Data, f, ctx.Locale)
}
