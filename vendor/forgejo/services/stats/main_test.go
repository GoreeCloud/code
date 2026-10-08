// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package stats

import (
	"fmt"
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
)

func TestMain(m *testing.M) {
	cfg, err := setting.NewConfigProviderFromData(`
[queue.stats_recalc]
TYPE = channel
`)
	if err != nil {
		panic(fmt.Sprintf("NewConfigProviderFromData: %v\n", err))
	}
	defer test.MockVariableValue(&setting.CfgProvider, cfg)()
	if err := Init(); err != nil {
		panic(fmt.Sprintf("stats.Init: %v\n", err))
	}

	m.Run()
}
