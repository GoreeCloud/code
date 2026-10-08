// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
package templates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	type ctxKey struct{}

	// Test that the original context is used for its context functions.
	ctx := NewContext(context.WithValue(t.Context(), ctxKey{}, "there"))
	assert.Equal(t, "there", ctx.Value(ctxKey{}))
}
