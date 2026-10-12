// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// Skeletal implementation of forgejo.org/services/context which allows for test data to access realistic methods on
// Base, Context, APIContext, etc.

package context

import (
	"io"
	"net/http"
	"time"
)

type Base struct{}

func (*Base) Status(status int) {}

func (*Base) Error(status int, contents ...string) {}

func (*Base) JSON(status int, content any) {}

func (*Base) PlainTextBytes(status int, bs []byte) {}

func (*Base) PlainText(status int, text string) {}

func (*Base) Redirect(location string, status ...int) {}

func (*Base) ServeContent(r io.ReadSeeker, opts *ServeHeaderOptions) {}

type ServeHeaderOptions struct {
	ContentType        string // defaults to "application/octet-stream"
	ContentTypeCharset string
	ContentLength      *int64
	Disposition        string // defaults to "attachment"
	Filename           string
	CacheDuration      time.Duration // defaults to 5 minutes
	LastModified       time.Time
	AdditionalHeaders  http.Header
	RedirectStatusCode int
}
