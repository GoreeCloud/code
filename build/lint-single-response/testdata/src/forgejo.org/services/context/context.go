// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// Skeletal implementation of forgejo.org/services/context which allows for test data to access realistic methods on
// Base, Context, APIContext, etc.

package context

type Context struct {
	*Base
}

func (*Context) HTML(status int, name string) {}

func (*Context) JSONError(msg any) {}

func (*Context) JSONOK() {}

func (*Context) JSONRedirect(redirect string) {}

func (*Context) JSONTemplate(tmpl string) {}

func (*Context) NotFound(logMsg string, logErr error) {}

func (*Context) NotFoundOrServerError(logMsg string, errCheck func(error) bool, logErr error) {}

func (*Context) RedirectToFirst(location ...string) string {
	return ""
}

func (*Context) RenderWithErr(msg any, tpl string, form any) {}

func (*Context) ServerError(logMsg string, logErr error) {}
