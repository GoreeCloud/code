// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"testing"

	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitConfig(t *testing.T) {
	defer test.MockProtect(&Git)()
	defer test.MockProtect(&GitConfig)()

	cfg, err := NewConfigProviderFromData(`
[git.config]
a.b = 1
`)
	require.NoError(t, err)
	loadGitFrom(cfg)
	assert.Equal(t, "1", GitConfig.Options["a.b"])
	assert.Equal(t, "histogram", GitConfig.Options["diff.algorithm"])

	cfg, err = NewConfigProviderFromData(`
[git.config]
diff.algorithm = other
`)
	require.NoError(t, err)
	loadGitFrom(cfg)
	assert.Equal(t, "other", GitConfig.Options["diff.algorithm"])

	t.Run("Fsck options", func(t *testing.T) {
		cfg, err := NewConfigProviderFromData(`
[git.config]
receive.fsckObjects = false
fetch.fsck.zeroPaddedFilemode = warn
fsck.missingSpaceBeforeDate = error
`)
		require.NoError(t, err)
		loadGitFrom(cfg)
		assert.Equal(t, "false", GitConfig.Options["receive.fsckobjects"])
		assert.Equal(t, "true", GitConfig.Options["transfer.fsckobjects"])
		assert.NotContains(t, GitConfig.Options, "fetch.fsckobjects")

		assert.Equal(t, "ignore", GitConfig.Options["fsck.badtimezone"])
		assert.Equal(t, "error", GitConfig.Options["fsck.missingspacebeforedate"])
		assert.Equal(t, "ignore", GitConfig.Options["fsck.zeropaddedfilemode"])

		assert.Equal(t, "ignore", GitConfig.Options["receive.fsck.badtimezone"])
		assert.Equal(t, "ignore", GitConfig.Options["receive.fsck.missingspacebeforedate"])
		assert.Equal(t, "ignore", GitConfig.Options["receive.fsck.zeropaddedfilemode"])

		assert.Equal(t, "ignore", GitConfig.Options["fetch.fsck.badtimezone"])
		assert.Equal(t, "ignore", GitConfig.Options["fetch.fsck.missingspacebeforedate"])
		assert.Equal(t, "warn", GitConfig.Options["fetch.fsck.zeropaddedfilemode"])
	})
}

func TestGitReflog(t *testing.T) {
	defer test.MockProtect(&Git)()
	defer test.MockProtect(&GitConfig)()

	// default reflog config.
	cfg, err := NewConfigProviderFromData(``)
	require.NoError(t, err)
	loadGitFrom(cfg)

	assert.Equal(t, "true", GitConfig.GetOption("core.logAllRefUpdates"))
	assert.Equal(t, "90", GitConfig.GetOption("gc.reflogExpire"))
}
