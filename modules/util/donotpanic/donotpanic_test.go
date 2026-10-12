// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package donotpanic

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoNotPanic_SafeFuncWithError(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		assert.NoError(t, SafeFuncWithError(func() error { return nil }))
	})

	t.Run("PanickString", func(t *testing.T) {
		errorMessage := "ERROR MESSAGE"
		assert.ErrorContains(t, SafeFuncWithError(func() error { panic(errorMessage) }), fmt.Sprintf("recover: %s", errorMessage))
	})

	t.Run("PanickError", func(t *testing.T) {
		errorMessage := "ERROR MESSAGE"
		assert.ErrorContains(t, SafeFuncWithError(func() error { panic(errors.New(errorMessage)) }), fmt.Sprintf("recover with error: %s", errorMessage))
	})
}
