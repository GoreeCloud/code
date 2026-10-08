// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/http"
	"os"
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/modules/util"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestJson(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/manifest.json")
	resp := session.MakeRequest(t, req, http.StatusOK)

	data := make(map[string]any)
	DecodeJSON(t, resp, &data)

	assert.Equal(t, setting.AppName, data["name"])
	assert.Equal(t, setting.AppName, data["short_name"])
	assert.Equal(t, setting.AppURL, data["start_url"])
	assert.NotContains(t, data, "display")
}

func TestManifestJsonStandalone(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	defer test.MockVariableValue(&setting.PWA.Standalone, true)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/manifest.json")
	resp := session.MakeRequest(t, req, http.StatusOK)

	data := make(map[string]any)
	DecodeJSON(t, resp, &data)

	assert.Equal(t, setting.AppName, data["name"])
	assert.Equal(t, setting.AppName, data["short_name"])
	assert.Equal(t, setting.AppURL, data["start_url"])
	assert.Contains(t, data, "display")
	assert.Equal(t, "standalone", data["display"])
}

func TestManifestJsonCustomFile(t *testing.T) {
	require.NoError(t, os.MkdirAll(util.FilePathJoinAbs(setting.CustomPath, "public"), 0o777))
	manifestPath := util.FilePathJoinAbs(setting.CustomPath, "public/manifest.json")
	file, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_RDWR, 0o777)
	require.NoError(t, err)
	_, err = file.Write([]byte(`{"name":"MyCustomJson"}`))
	require.NoError(t, err)
	require.NoError(t, file.Close())
	defer os.Remove(manifestPath)

	defer tests.PrepareTestEnv(t)()

	session := loginUser(t, "user2")
	req := NewRequest(t, "GET", "/manifest.json")
	resp := session.MakeRequest(t, req, http.StatusOK)

	data := make(map[string]any)
	DecodeJSON(t, resp, &data)

	assert.Equal(t, "MyCustomJson", data["name"])
	assert.NotContains(t, data, "short_name")
	assert.NotContains(t, data, "start_url")
	assert.NotContains(t, data, "display")
}
