// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgefed_test

import (
	"testing"

	"forgejo.org/modules/forgefed"
	"forgejo.org/modules/validation"

	ap "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
)

func Test_UserActivityNoteValidation(t *testing.T) {
	sut := forgefed.ForgeUserActivityNote{}
	sut.Type = "Note"
	sut.Content = ap.NaturalLanguageValues{
		{
			Ref:   ap.NilLangRef,
			Value: ap.Content("Any Content!"),
		},
	}
	sut.URL = ap.IRI("example.org/user-id/57")

	valid, _ := validation.IsValid(sut)
	assert.True(t, valid, "sut expected to be valid: %v\n", sut.Validate())
}
