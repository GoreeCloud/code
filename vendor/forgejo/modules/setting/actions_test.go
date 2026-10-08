// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"fmt"
	"path/filepath"
	"testing"

	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getStorageInheritNameSectionTypeForActions(t *testing.T) {
	iniStr := `
	[storage]
	STORAGE_TYPE = minio
	`
	cfg, err := NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "minio", Actions.LogStorage.Type)
	assert.Equal(t, "actions_log/", Actions.LogStorage.MinioConfig.BasePath)
	assert.EqualValues(t, "minio", Actions.ArtifactStorage.Type)
	assert.Equal(t, "actions_artifacts/", Actions.ArtifactStorage.MinioConfig.BasePath)

	iniStr = `
[storage.actions_log]
STORAGE_TYPE = minio
`
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "minio", Actions.LogStorage.Type)
	assert.Equal(t, "actions_log/", Actions.LogStorage.MinioConfig.BasePath)
	assert.EqualValues(t, "local", Actions.ArtifactStorage.Type)
	assert.Equal(t, "actions_artifacts", filepath.Base(Actions.ArtifactStorage.Path))

	iniStr = `
[storage.actions_log]
STORAGE_TYPE = my_storage

[storage.my_storage]
STORAGE_TYPE = minio
`
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "minio", Actions.LogStorage.Type)
	assert.Equal(t, "actions_log/", Actions.LogStorage.MinioConfig.BasePath)
	assert.EqualValues(t, "local", Actions.ArtifactStorage.Type)
	assert.Equal(t, "actions_artifacts", filepath.Base(Actions.ArtifactStorage.Path))

	iniStr = `
[storage.actions_artifacts]
STORAGE_TYPE = my_storage

[storage.my_storage]
STORAGE_TYPE = minio
`
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "local", Actions.LogStorage.Type)
	assert.Equal(t, "actions_log", filepath.Base(Actions.LogStorage.Path))
	assert.EqualValues(t, "minio", Actions.ArtifactStorage.Type)
	assert.Equal(t, "actions_artifacts/", Actions.ArtifactStorage.MinioConfig.BasePath)

	iniStr = `
[storage.actions_artifacts]
STORAGE_TYPE = my_storage

[storage.my_storage]
STORAGE_TYPE = minio
`
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "local", Actions.LogStorage.Type)
	assert.Equal(t, "actions_log", filepath.Base(Actions.LogStorage.Path))
	assert.EqualValues(t, "minio", Actions.ArtifactStorage.Type)
	assert.Equal(t, "actions_artifacts/", Actions.ArtifactStorage.MinioConfig.BasePath)

	iniStr = ``
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "local", Actions.LogStorage.Type)
	assert.Equal(t, "actions_log", filepath.Base(Actions.LogStorage.Path))
	assert.EqualValues(t, "local", Actions.ArtifactStorage.Type)
	assert.Equal(t, "actions_artifacts", filepath.Base(Actions.ArtifactStorage.Path))
}

func Test_getDefaultActionsURLForActions(t *testing.T) {
	oldActions := Actions
	oldAppURL := AppURL
	defer func() {
		Actions = oldActions
		AppURL = oldAppURL
	}()

	AppURL = "http://test_get_default_actions_url_for_actions:3000/"

	tests := []struct {
		name    string
		iniStr  string
		wantURL string
	}{
		{
			name: "default",
			iniStr: `
[actions]
`,
			wantURL: "https://data.forgejo.org",
		},
		{
			name: "github",
			iniStr: `
[actions]
DEFAULT_ACTIONS_URL = github
`,
			wantURL: "https://github.com",
		},
		{
			name: "self",
			iniStr: `
[actions]
DEFAULT_ACTIONS_URL = self
`,
			wantURL: "http://test_get_default_actions_url_for_actions:3000",
		},
		{
			name: "custom urls",
			iniStr: `
[actions]
DEFAULT_ACTIONS_URL = https://example.com
`,
			wantURL: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewConfigProviderFromData(tt.iniStr)
			require.NoError(t, err)
			require.NoError(t, loadActionsFrom(cfg))

			assert.Equal(t, tt.wantURL, Actions.DefaultActionsURL.URL())
		})
	}
}

func Test_getIDTokenSettingsForActions(t *testing.T) {
	defer test.MockVariableValue(&AppDataPath, "/home/app/data")()

	oldActions := Actions
	oldAppURL := AppURL
	defer func() {
		Actions = oldActions
		AppURL = oldAppURL
	}()

	iniStr := `
  [actions]
  `
	cfg, err := NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "RS256", Actions.IDTokenSigningAlgorithm)
	assert.Equal(t, "/home/app/data/actions_id_token/private.pem", Actions.IDTokenSigningPrivateKeyFile)
	assert.EqualValues(t, 3600, Actions.IDTokenExpirationTime)

	iniStr = `
  [actions]
	ID_TOKEN_SIGNING_ALGORITHM = ES256
  ID_TOKEN_SIGNING_PRIVATE_KEY_FILE = /test/test.pem
  ID_TOKEN_EXPIRATION_TIME = 120
	`
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "ES256", Actions.IDTokenSigningAlgorithm)
	assert.Equal(t, "/test/test.pem", Actions.IDTokenSigningPrivateKeyFile)
	assert.EqualValues(t, 120, Actions.IDTokenExpirationTime)

	iniStr = `
  [actions]
	ID_TOKEN_SIGNING_ALGORITHM = EdDSA
  ID_TOKEN_SIGNING_PRIVATE_KEY_FILE = ./test/test.pem
  ID_TOKEN_EXPIRATION_TIME = 123
	`
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	require.NoError(t, loadActionsFrom(cfg))

	assert.EqualValues(t, "EdDSA", Actions.IDTokenSigningAlgorithm)
	assert.Equal(t, "/home/app/data/test/test.pem", Actions.IDTokenSigningPrivateKeyFile)
	assert.EqualValues(t, 123, Actions.IDTokenExpirationTime)

	iniStr = `
  [actions]
	ID_TOKEN_SIGNING_ALGORITHM = HS256
	`
	cfg, err = NewConfigProviderFromData(iniStr)
	require.NoError(t, err)
	err = loadActionsFrom(cfg)
	require.ErrorContains(t, err,
		fmt.Sprintf("invalid [actions] ID_TOKEN_SIGNING_ALGORITHM: %q", Actions.IDTokenSigningAlgorithm))
}
