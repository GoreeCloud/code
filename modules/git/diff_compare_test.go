// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
package git

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckIfDiffDiffers(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitRepository(t.Context(), tmpDir, false, Sha1ObjectFormat.Name())
	require.NoError(t, err)

	gitRepo, err := openRepositoryWithDefaultContext(tmpDir)
	require.NoError(t, err)
	defer gitRepo.Close()

	require.NoError(t, NewCommand(t.Context(), "switch", "-c", "main").Run(&RunOpts{Dir: tmpDir}))

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("aaa"), 0o600))
	require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
	require.NoError(t, NewCommand(t.Context(), "commit", "-m", "initial commit").Run(&RunOpts{Dir: tmpDir}))

	t.Run("Simple fast-forward", func(t *testing.T) {
		// Check that A--B--C, where A is the base branch.

		t.Run("Different diff", func(t *testing.T) {
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "a-1", "main").Run(&RunOpts{Dir: tmpDir}))

			// B commit
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "a-2").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("ccc"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main", "a-1", "a-2", nil)
			require.NoError(t, err)
			assert.True(t, changed)
		})

		t.Run("Same diff", func(t *testing.T) {
			// Because C is a empty commit, the diff does not differ relative to the
			// base branch.

			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "a-3", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "a-4").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "--allow-empty", "-m", "No changes to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main", "a-3", "a-4", nil)
			require.NoError(t, err)
			assert.False(t, changed)
		})
	})

	t.Run("Merge-base is base reference", func(t *testing.T) {
		//   B
		//  /
		// A
		//  \
		//   C
		t.Run("Different diff", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "b-1", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changed to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "b-2", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("ccc"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changed to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main", "b-1", "b-2", nil)
			require.NoError(t, err)
			assert.True(t, changed)
		})

		t.Run("Same diff", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "b-3", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changed to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "b-4", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changed to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main", "b-3", "b-4", nil)
			require.NoError(t, err)
			assert.False(t, changed)
		})
	})

	t.Run("Merge-base is different", func(t *testing.T) {
		//   B
		//  /
		// A--D
		//     \
		//      C
		// Where D is the base reference.

		// D commit, where A is `main`.
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "main-D", "main").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "FUNFACT"), []byte("Smithy was the runner up to be Forgejo's name"), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "FUNFACT").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Forgejo history").Run(&RunOpts{Dir: tmpDir}))

		t.Run("Different diff", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "c-1", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "c-2", "main-D").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "FUNFACT"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "FUNFACT").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the funfact").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main-D", "c-1", "c-2", nil)
			require.NoError(t, err)
			assert.True(t, changed)
		})

		t.Run("Same diff", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "c-3", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "c-4", "main-D").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main-D", "c-3", "c-4", nil)
			require.NoError(t, err)
			assert.False(t, changed)
		})
	})

	t.Run("Merge commit", func(t *testing.T) {
		//   B
		//  /
		// A - D
		//      \
		//       C
		//
		// From B, it merges D where E is the merge commit :
		//   B---E
		//  /   /
		// A---D
		//      \
		//       C

		t.Run("Different diff", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "d-1", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))
			// E commit
			require.NoError(t, NewCommand(t.Context(), "merge", "--no-ff", "main-D").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "d-2", "main-D").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "FUNFACT"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "FUNFACT").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the funfact").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main-D", "d-1", "d-2", nil)
			require.NoError(t, err)
			assert.True(t, changed)
		})

		t.Run("Same diff", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "d-3", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))
			// Merges D.
			require.NoError(t, NewCommand(t.Context(), "merge", "--no-ff", "main-D").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "d-4", "main-D").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main-D", "d-3", "d-4", nil)
			require.NoError(t, err)
			assert.False(t, changed)
		})
	})

	t.Run("Non-typical rebase", func(t *testing.T) {
		//   B
		//  /
		// A--D
		//     \
		//      C
		// Where D is the base reference.
		// B was rebased onto D, which produced C.
		// B and D made the same change to same file.

		// D commit.
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "main-D-2", "main").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "FUNFACT"), []byte("Smithy was the runner up to be Forgejo's name"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("ccc"), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "FUNFACT", "README").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Forgejo history").Run(&RunOpts{Dir: tmpDir}))

		// B commit
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "e-1", "main").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "CONTACT"), []byte("@example.com"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("ccc"), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "README", "CONTACT").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the contact and README").Run(&RunOpts{Dir: tmpDir}))

		// C commit
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "e-2").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "rebase", "main-D-2").Run(&RunOpts{Dir: tmpDir}))

		// The diff changed, because it no longer shows the change made to `README`.
		changed, err := gitRepo.CheckIfDiffDiffers("main-D-2", "e-1", "e-2", nil)
		require.NoError(t, err)
		assert.False(t, changed) // This should be true.
	})

	t.Run("Directory", func(t *testing.T) {
		//   B
		//  /
		// A
		//  \
		//   C
		t.Run("Same directory", func(t *testing.T) {
			t.Run("Different diff", func(t *testing.T) {
				// B commit
				require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-1", "main").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "docs"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs", "README"), []byte("bbb"), 0o600))
				require.NoError(t, NewCommand(t.Context(), "add", "docs/README").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changed to the README").Run(&RunOpts{Dir: tmpDir}))

				// C commit
				require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-2", "main").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "docs"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs", "README"), []byte("ccc"), 0o600))
				require.NoError(t, NewCommand(t.Context(), "add", "docs/README").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changed to the README").Run(&RunOpts{Dir: tmpDir}))

				changed, err := gitRepo.CheckIfDiffDiffers("main", "f-1", "f-2", nil)
				require.NoError(t, err)
				assert.True(t, changed)
			})

			t.Run("Same diff", func(t *testing.T) {
				// B commit
				require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-3", "main").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "docs"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs", "README"), []byte("bbb"), 0o600))
				require.NoError(t, NewCommand(t.Context(), "add", "docs/README").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changed to the README").Run(&RunOpts{Dir: tmpDir}))

				// C commit
				require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-4", "main").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "docs"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs", "README"), []byte("bbb"), 0o600))
				require.NoError(t, NewCommand(t.Context(), "add", "docs/README").Run(&RunOpts{Dir: tmpDir}))
				require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

				changed, err := gitRepo.CheckIfDiffDiffers("main", "f-3", "f-4", nil)
				require.NoError(t, err)
				assert.False(t, changed)
			})
		})

		t.Run("Different directory", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-5", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "docs"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs", "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "docs/README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-6", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "documentation"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "documentation", "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "documentation/README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main", "f-5", "f-6", nil)
			require.NoError(t, err)
			assert.True(t, changed)
		})

		t.Run("Directory and file", func(t *testing.T) {
			// B commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-7", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "docs"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs", "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "docs/README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			// C commit
			require.NoError(t, NewCommand(t.Context(), "switch", "-c", "f-8", "main").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "README"), []byte("bbb"), 0o600))
			require.NoError(t, NewCommand(t.Context(), "add", "README").Run(&RunOpts{Dir: tmpDir}))
			require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Changes to the README").Run(&RunOpts{Dir: tmpDir}))

			changed, err := gitRepo.CheckIfDiffDiffers("main", "f-7", "f-8", nil)
			require.NoError(t, err)
			assert.True(t, changed)
		})
	})

	t.Run("Rebase", func(t *testing.T) {
		//   B
		//  /
		// A--D
		//     \
		//      C
		// Where D is the base reference.
		// B was rebased onto D, which produced C.
		// B and D made different (non conflicting) changes to same file.

		// A commit
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "main-3", "main").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "REBASE"), bytes.Repeat([]byte{'b', 'b', 'b', '\n'}, 100), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "REBASE").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Rebasing is fun").Run(&RunOpts{Dir: tmpDir}))

		// B commit
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "g-1", "main-3").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "REBASE"), append(bytes.Repeat([]byte{'b', 'b', 'b', '\n'}, 100), 'a', 'a', 'a'), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "REBASE").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Rebasing is fun").Run(&RunOpts{Dir: tmpDir}))

		// D commit
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "g-2", "main-3").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "REBASE"), append([]byte{'a', 'a', 'a'}, bytes.Repeat([]byte{'b', 'b', 'b', '\n'}, 99)...), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "REBASE").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Rebasing is fun").Run(&RunOpts{Dir: tmpDir}))

		// C commit
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "g-3", "g-1").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "rebase", "g-2").Run(&RunOpts{Dir: tmpDir}))

		changed, err := gitRepo.CheckIfDiffDiffers("g-2", "g-1", "g-3", nil)
		require.NoError(t, err)
		assert.True(t, changed)
	})

	t.Run("Rebasing change not shown", func(t *testing.T) {
		//   B
		//  /
		// A--D
		//     \
		//      C
		// Where D is the base reference.
		// B was rebased onto D, which produced C.
		// B and D made different (non conflicting) changes to same file.

		// A commit
		require.NoError(t, NewCommand(t.Context(), "switch", "--orphan", "main-4").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "A"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "A", "a"), bytes.Repeat([]byte{'A', 'A', 'A', '\n'}, 100), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "A/a").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Just wondering").Run(&RunOpts{Dir: tmpDir}))

		// B commit
		// Changes last line.
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "h-1").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "A", "a"), append(bytes.Repeat([]byte{'A', 'A', 'A', '\n'}, 99), 'B', 'B', 'B', '\n'), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "A/a").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Just wondering").Run(&RunOpts{Dir: tmpDir}))

		// D commit
		// Changes first line.
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "h-2", "main-4").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "A", "a"), append([]byte{'B', 'B', 'B', '\n'}, bytes.Repeat([]byte{'A', 'A', 'A', '\n'}, 99)...), 0o600))
		require.NoError(t, NewCommand(t.Context(), "add", "A/a").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "commit", "-m", "Just wondering").Run(&RunOpts{Dir: tmpDir}))

		// C commit
		require.NoError(t, NewCommand(t.Context(), "switch", "-c", "h-3").Run(&RunOpts{Dir: tmpDir}))
		require.NoError(t, NewCommand(t.Context(), "rebase", "h-2").Run(&RunOpts{Dir: tmpDir}))

		changed, err := gitRepo.CheckIfDiffDiffers("h-2", "h-1", "h-3", nil)
		require.NoError(t, err)

		assert.False(t, changed)
	})
}
