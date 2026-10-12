// Copyright 2024 The Forgejo Authors.
// SPDX-License-Identifier: GPLv3-or-later

package setting

import (
	"fmt"
	"testing"

	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingF3(t *testing.T) {
	restoreF3 := test.MockProtect(&F3)
	restoreAppDataPath := test.MockProtect(&AppDataPath)
	restore := func() {
		restoreF3()
		restoreAppDataPath()
	}
	defer restore()

	t.Run("enabled", func(t *testing.T) {
		restore()
		cfg, err := NewConfigProviderFromData(`
[f3]
ENABLED = true
`)
		require.NoError(t, err)
		loadF3From(cfg)
		assert.True(t, F3.Enabled)
		assert.DirExists(t, F3.Path)
	})

	t.Run("disabled by default", func(t *testing.T) {
		restore()
		cfg, err := NewConfigProviderFromData(`
[f3]
`)
		require.NoError(t, err)
		loadF3From(cfg)
		assert.False(t, F3.Enabled)
	})

	t.Run("default f3 path", func(t *testing.T) {
		restore()
		cfg, err := NewConfigProviderFromData(`
[f3]
ENABLED = true
`)
		require.NoError(t, err)
		AppDataPath = t.TempDir()
		loadF3From(cfg)
		assert.Equal(t, AppDataPath+"/f3", F3.Path)
		assert.DirExists(t, F3.Path)
	})

	t.Run("absolute f3 path", func(t *testing.T) {
		restore()
		other := t.TempDir()
		cfg, err := NewConfigProviderFromData(fmt.Sprintf(`
[f3]
ENABLED = true
PATH = %[1]s
`, other))
		require.NoError(t, err)
		AppDataPath = t.TempDir()
		loadF3From(cfg)
		assert.Equal(t, other, F3.Path)
		assert.DirExists(t, F3.Path)
	})
}
