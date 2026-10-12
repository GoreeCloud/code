// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
package git

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"forgejo.org/modules/log"
	"forgejo.org/modules/util"
)

// CheckIfDiffDiffers returns if the diff of the newCommitID and
// oldCommitID with the merge base of the base branch has changed.
//
// Informally it checks if the following two diffs are exactly the same in their
// contents, thus ignoring different commit IDs, headers and messages:
// 1. git diff --merge-base baseReference newCommitID
// 2. git diff --merge-base baseReference oldCommitID
func (repo *Repository) CheckIfDiffDiffers(base, oldCommitID, newCommitID string, env []string) (hasChanged bool, err error) {
	cmd := NewCommand(repo.Ctx, "diff", "--name-only", "-z").AddDynamicArguments(newCommitID, oldCommitID, base)
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return false, fmt.Errorf("unable to open pipe for to run diff: %w", err)
	}

	stderr := new(bytes.Buffer)
	if err := cmd.Run(&RunOpts{
		Dir:    repo.Path,
		Stdout: stdoutWriter,
		Stderr: stderr,
		PipelineFunc: func(ctx context.Context, cancel context.CancelFunc) error {
			_ = stdoutWriter.Close()
			defer func() {
				_ = stdoutReader.Close()
			}()
			return util.IsEmptyReader(stdoutReader)
		},
	}); err != nil {
		if err == util.ErrNotEmpty {
			return true, nil
		}
		err = ConcatenateError(err, stderr.String())

		log.Error("Unable to run git diff on %s %s %s in %q: Error: %v",
			newCommitID, oldCommitID, base,
			repo.Path,
			err)

		return false, fmt.Errorf("Unable to run git diff --name-only -z %s %s %s: %w", newCommitID, oldCommitID, base, err)
	}

	return false, nil
}
