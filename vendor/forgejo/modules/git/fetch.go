// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later
package git

import (
	"errors"
	"fmt"
	"strings"

	"forgejo.org/modules/log"
	"forgejo.org/modules/util"
)

var ErrRemoteRefNotFound = errors.New("unable to find remote ref")

// Fetch executes git-fetch(1) for the repository, it will fetch the refspec
// from the remote into the repository.
//
// Valid remote URLs are denoted in section GIT URLS of git-fetch(1).
//
// The return value, on success, is the object ID of the remote reference. If
// no local reference is given in `refspec` then do not assume it's available in
// the `FETCH_HEAD` reference, it might have been overwritten by the time you
// read or use it.
func (repo *Repository) Fetch(remoteURL, refspec string) (string, error) {
	objectFormat, err := repo.GetObjectFormat()
	if err != nil {
		return "", err
	}

	cmd := NewCommand(repo.Ctx, "fetch")
	if SupportFetchPorcelain {
		cmd.AddArguments("--porcelain")
	} else if !strings.Contains(refspec, ":") {
		refspec += ":refs/tmp/" + util.CryptoRandomString(util.RandomStringLow)
	}

	cmd.AddArguments("--end-of-options").AddDynamicArguments(remoteURL, refspec)

	stdout, stderr, err := cmd.RunStdString(&RunOpts{Dir: repo.Path})
	if err != nil {
		if strings.HasPrefix(stderr, "fatal: couldn't find remote ref ") {
			return "", ErrRemoteRefNotFound
		}
		return "", err
	}

	_, localReference, ok := strings.Cut(refspec, ":")
	if !ok {
		localReference = "FETCH_HEAD"
	}

	// Happy path
	if SupportFetchPorcelain {
		// The porcelain format, per section OUTPUT of git-fetch(1), is expected to be:
		// <flag><space><old-object-id><space><new-object-id><space><local-reference>\n
		// flag is one character.
		if expectedLen := 1 + 1 + objectFormat.FullLength() + 1 + objectFormat.FullLength() + 1 + len(localReference) + 1; len(stdout) != expectedLen {
			return "", fmt.Errorf("output of git-fetch(1) is unexpected, we expected it to be %d bytes but it is %d bytes. stdout: %s", expectedLen, len(stdout), stdout)
		}

		// Extract the new objectID.
		newObjectID := stdout[1+1+objectFormat.FullLength()+1 : 1+1+objectFormat.FullLength()+1+objectFormat.FullLength()]
		return newObjectID, nil
	}

	defer func() {
		if err := repo.RemoveReference(localReference); err != nil {
			log.Error("Could not remove reference %q from repository %q: %v", localReference, repo.Path, err)
		}
	}()

	newObjectID, err := repo.ResolveReference(localReference)
	if err != nil {
		return "", err
	}

	return newObjectID, nil
}
