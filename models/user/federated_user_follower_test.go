// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"testing"

	"forgejo.org/modules/validation"

	"github.com/stretchr/testify/assert"
)

func Test_FederatedUserFollowerValidation(t *testing.T) {
	sut := FederatedUserFollower{
		FollowedUserID:  12,
		FollowingUserID: 1,
	}
	res, err := validation.IsValid(sut)
	assert.Truef(t, res, "sut should be valid but was %q", err)

	sut = FederatedUserFollower{
		FollowedUserID: 1,
	}
	res, _ = validation.IsValid(sut)
	assert.False(t, res, "sut should be invalid")
}
