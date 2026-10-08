// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
package auth

import (
	"testing"

	"forgejo.org/models/unittest"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasTwoFactorByUID(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	t.Run("No twofactor", func(t *testing.T) {
		ok, err := HasTwoFactorByUID(t.Context(), 2)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("WebAuthn credential", func(t *testing.T) {
		ok, err := HasTwoFactorByUID(t.Context(), 32)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("TOTP", func(t *testing.T) {
		ok, err := HasTwoFactorByUID(t.Context(), 24)
		require.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestNewTwoFactor(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	otpKey, err := totp.Generate(totp.GenerateOpts{
		SecretSize:  40,
		Issuer:      "forgejo-test",
		AccountName: "user2",
	})
	require.NoError(t, err)

	t.Run("Transaction failed", func(t *testing.T) {
		reset := unittest.SetFaultInjector(2)
		require.ErrorIs(t, NewTwoFactor(t.Context(), &TwoFactor{UID: 44}, otpKey.Secret()), unittest.ErrFaultInjected)
		reset()

		unittest.AssertExistsIf(t, false, &TwoFactor{UID: 44})
	})

	t.Run("Normal", func(t *testing.T) {
		reset := unittest.SetFaultInjector(4)
		require.NoError(t, NewTwoFactor(t.Context(), &TwoFactor{UID: 44}, otpKey.Secret()))
		reset()

		unittest.AssertExistsIf(t, true, &TwoFactor{UID: 44})
	})
}
