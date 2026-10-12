// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package templates

import (
	"context"

	"forgejo.org/modules/translation"
)

type Context struct {
	context.Context
	Locale      translation.Locale
	AvatarUtils *AvatarUtils
	Data        map[string]any
}

var _ context.Context = Context{}

func NewContext(ctx context.Context) *Context {
	return &Context{Context: ctx}
}
