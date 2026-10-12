// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"

	asymkey_model "forgejo.org/models/asymkey"

	"github.com/urfave/cli/v3"
)

var microcmdRegenKeys = &cli.Command{
	Name:   "keys",
	Usage:  "Regenerate authorized_keys file",
	Before: noDanglingArgs,
	Action: runRegenerateKeys,
}

func runRegenerateKeys(ctx context.Context, c *cli.Command) error {
	ctx, cancel := installSignals(ctx)
	defer cancel()

	if err := initDB(ctx); err != nil {
		return err
	}
	return asymkey_model.RewriteAllPublicKeys(ctx)
}
