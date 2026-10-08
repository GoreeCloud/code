// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forms

import (
	"net/http"

	"forgejo.org/modules/web/middleware"
	"forgejo.org/services/context"

	"code.forgejo.org/go-chi/binding"
)

// CreateSecretForm needs to be filled in by the user to create a new secret.
type CreateSecretForm struct {
	Name string `binding:"Required;MaxSize(255)"`
	Data string `binding:"Required;MaxSize(65535)"`
}

// Validate validates the submitted CreateSecretForm.
func (f *CreateSecretForm) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	ctx := context.GetValidateContext(req)
	return middleware.Validate(errs, ctx.Data, f, ctx.Locale)
}

// EditSecretForm needs to be filled in by the user to change an existing secret.
type EditSecretForm struct {
	Name string `binding:"Required;MaxSize(255)"`
	Data string `binding:"MaxSize(65535)"`
}

// Validate validates the submitted EditSecretForm.
func (f *EditSecretForm) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	ctx := context.GetValidateContext(req)
	return middleware.Validate(errs, ctx.Data, f, ctx.Locale)
}
