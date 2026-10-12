// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package lfs

import (
	"context"
	"fmt"
	"os"
	"testing"

	"forgejo.org/modules/git"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/util"
)

const testReposDir = "tests/repos/"

func TestMain(m *testing.M) {
	gitHomePath, err := os.MkdirTemp(os.TempDir(), "git-home")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Test failed: unable to create temp dir: %v", err)
		os.Exit(1)
	}

	setting.Git.HomePath = gitHomePath

	if err = git.InitFull(context.Background()); err != nil {
		util.RemoveAll(gitHomePath)
		_, _ = fmt.Fprintf(os.Stderr, "Test failed: failed to call git.InitFull: %v", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	util.RemoveAll(gitHomePath)
	os.Exit(exitCode)
}
