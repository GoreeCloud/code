// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package method

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/lfs"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/services/auth"
	"forgejo.org/services/authz"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authTokenOptions struct {
	Op     string
	UserID int64
	RepoID int64
}

func getLFSAuthTokenWithBearer(opts authTokenOptions) (string, error) {
	now := time.Now()
	claims := lfs.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(setting.LFS.HTTPAuthExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
		RepoID: opts.RepoID,
		Op:     opts.Op,
		UserID: opts.UserID,
	}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := setting.LFS.SigningKey.JWT(claims)
	if err != nil {
		return "", fmt.Errorf("failed to sign LFS JWT token: %w", err)
	}
	return "Bearer " + tokenString, nil
}

func testAuthenticate(t *testing.T, cfg string) {
	require.NoError(t, unittest.PrepareTestDatabase())
	var err error
	setting.CfgProvider, err = setting.NewConfigProviderFromData(cfg)
	require.NoError(t, err, "Config")
	setting.LoadCommonSettings()

	t.Run("no bearer token", func(t *testing.T) {
		a := &LFSToken{}
		req := httptest.NewRequest("GET", "https://example.org", nil)
		output := a.Verify(req, nil, nil)
		requireOutput[*auth.AuthenticationNotAttempted](t, output)
	})

	t.Run("bearer not a JWT", func(t *testing.T) {
		a := &LFSToken{}
		req := httptest.NewRequest("GET", "https://example.org", nil)
		req.Header.Set("Authorization", "Bearer abc")
		output := a.Verify(req, nil, nil)
		err := requireOutput[*auth.AuthenticationAttemptedIncorrectCredential](t, output).Error
		require.ErrorContains(t, err, "token is malformed")
	})

	t.Run("token valid op=download", func(t *testing.T) {
		bearerAuth, err := getLFSAuthTokenWithBearer(authTokenOptions{Op: "download", UserID: 2, RepoID: 1})
		require.NoError(t, err)
		a := &LFSToken{}
		req := httptest.NewRequest("GET", "https://example.org", nil)
		req.Header.Set("Authorization", bearerAuth)
		output := a.Verify(req, nil, nil)
		result := requireOutput[*auth.AuthenticationSuccess](t, output).Result
		assert.EqualValues(t, 2, result.User().ID)
		assert.Equal(t, optional.Some(auth_model.AccessTokenScopeReadRepository), result.Scope())
		require.NotNil(t, result.Reducer())

		// No direct way to query an authz.Reducer for its specific repos permitted, so this is a bit of a workaround to
		// see if it's targeting the repo from the JWT:
		repoGetter, isRepoGetter := result.(authz.RepoGetter)
		require.True(t, isRepoGetter)
		assert.EqualValues(t, 1, repoGetter.GetTargetRepoID())
	})

	t.Run("token valid op=upload", func(t *testing.T) {
		bearerAuth, err := getLFSAuthTokenWithBearer(authTokenOptions{Op: "upload", UserID: 2, RepoID: 24})
		require.NoError(t, err)
		a := &LFSToken{}
		req := httptest.NewRequest("GET", "https://example.org", nil)
		req.Header.Set("Authorization", bearerAuth)
		output := a.Verify(req, nil, nil)
		result := requireOutput[*auth.AuthenticationSuccess](t, output).Result
		assert.EqualValues(t, 2, result.User().ID)
		assert.Equal(t, optional.Some(auth_model.AccessTokenScopeWriteRepository), result.Scope())
		require.NotNil(t, result.Reducer())

		// No direct way to query an authz.Reducer for its specific repos permitted, so this is a bit of a workaround to
		// see if it's targeting the repo from the JWT:
		repoGetter, isRepoGetter := result.(authz.RepoGetter)
		require.True(t, isRepoGetter)
		assert.EqualValues(t, 24, repoGetter.GetTargetRepoID())
	})

	t.Run("token signature malformed", func(t *testing.T) {
		bearerAuth, err := getLFSAuthTokenWithBearer(authTokenOptions{Op: "download", UserID: 2, RepoID: 1})
		bearerAuth += "malformed"

		require.NoError(t, err)
		a := &LFSToken{}
		req := httptest.NewRequest("GET", "https://example.org", nil)
		req.Header.Set("Authorization", bearerAuth)
		output := a.Verify(req, nil, nil)
		err = requireOutput[*auth.AuthenticationAttemptedIncorrectCredential](t, output).Error
		require.ErrorContains(t, err, "token signature is invalid")
	})

	t.Run("invalid user", func(t *testing.T) {
		bearerAuth, err := getLFSAuthTokenWithBearer(authTokenOptions{Op: "download", UserID: 999, RepoID: 1})
		require.NoError(t, err)
		a := &LFSToken{}
		req := httptest.NewRequest("GET", "https://example.org", nil)
		req.Header.Set("Authorization", bearerAuth)
		output := a.Verify(req, nil, nil)
		err = requireOutput[*auth.AuthenticationError](t, output).Error
		require.ErrorContains(t, err, "user does not exist")
	})
}

type namedCfg struct {
	name, cfg string
}

var iniCommon = `[security]
INSTALL_LOCK = true
INTERNAL_TOKEN = ForgejoForgejoForgejoForgejoForgejoForgejo_	# don't use in prod
[oauth2]
JWT_SECRET = ForgejoForgejoForgejoForgejoForgejoForgejo_	# don't use in prod
[server]
LFS_START_SERVER = true
	`

var cfgVariants = []namedCfg{
	{name: "HS256_default", cfg: `LFS_JWT_SECRET = ForgejoForgejoForgejoForgejoForgejoForgejo_`},
	{name: "RS256", cfg: `LFS_JWT_SIGNING_ALGORITHM = RS256`},
}

func TestAuthenticate(t *testing.T) {
	for _, v := range cfgVariants {
		cfg := iniCommon + v.cfg
		t.Run(v.name, func(t *testing.T) { testAuthenticate(t, cfg) })
	}
}
