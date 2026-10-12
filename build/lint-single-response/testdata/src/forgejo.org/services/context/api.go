// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package context

type APIContext struct {
	*Base
}

func (ctx *APIContext) Error(status int, title string, obj any) {}

func (ctx *APIContext) InternalServerError(err error) {}

func (ctx *APIContext) NotFound(objs ...any) {}

func (ctx *APIContext) NotFoundOrServerError(logMsg string, errCheck func(error) bool, logErr error) {
}

func (ctx *APIContext) ServerError(title string, err error) {}
