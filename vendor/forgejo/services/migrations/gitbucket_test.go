// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package migrations

import (
	"os"
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/services/migrations/allowlist"

	"github.com/stretchr/testify/require"
)

func TestGitbucketDownloaderCreation(t *testing.T) {
	defer test.MockVariableValueWithReset(&setting.Migrations.AllowLocalNetworks, true, func() { require.NoError(t, allowlist.Init()) })()

	token := os.Getenv("GITHUB_READ_TOKEN")
	fixturePath := "./testdata/github/full_download"
	server := unittest.NewMockWebServer(t, "https://api.github.com", fixturePath, false)
	defer server.Close()

	downloader := NewGitBucketDownloader(t.Context(), server.URL, "", "", token, "forgejo", "test_repo")
	err := downloader.RefreshRate()
	require.NoError(t, err)
}
