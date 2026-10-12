// Copyright 2015 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeEntry_Path(t *testing.T) {
	repo, err := openRepositoryWithDefaultContext(filepath.Join(testReposDir, "templates_repo"))
	require.NoError(t, err)
	defer repo.Close()

	tests := []struct {
		name string // description of this test case
		path string
	}{
		{
			name: "Top level dir",
			path: ".forgejo",
		},
		{
			name: "File in subdir",
			path: ".forgejo/default_merge_message/MERGE_TEMPLATE.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := repo.GetTree("HEAD^{tree}")
			require.NoError(t, err)

			te, err := tree.GetTreeEntryByPath(tt.path)
			require.NoError(t, err)

			got, gotErr := te.Path()
			require.NoError(t, gotErr, "Path() failed: %v", gotErr)
			assert.Equal(t, tt.path, got, "Path() = %v, want %v", got, tt.path)
		})
	}
}
