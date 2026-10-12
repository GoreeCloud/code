// SPDX-License-Identifier: MIT

package setting

import (
	"os"
	"path/filepath"

	"forgejo.org/modules/log"
)

// Friendly Forge Format (F3) settings
var (
	F3 = struct {
		Enabled bool
		Path    string
	}{
		Enabled: false,
		Path:    "f3",
	}
)

func LoadF3Setting() {
	loadF3From(CfgProvider)
}

func loadF3From(rootCfg ConfigProvider) {
	if err := rootCfg.Section("f3").MapTo(&F3); err != nil {
		log.Fatal("Failed to map F3 settings: %v", err)
	}

	if !filepath.IsAbs(F3.Path) {
		F3.Path = filepath.Join(AppDataPath, F3.Path)
	} else {
		F3.Path = filepath.Clean(F3.Path)
	}

	if err := os.MkdirAll(F3.Path, os.ModePerm); err != nil {
		log.Fatal("Failed to create F3 path %s: %v", F3.Path, err)
	}
}
