// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"errors"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/services/auth/source/pam"

	"github.com/urfave/cli/v3"
)

func pamCLIFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "Application Name",
		},
		&cli.StringFlag{
			Name:  "service-name",
			Value: "PLAIN",
			Usage: "PAM service name",
		},
		&cli.StringFlag{
			Name:  "email-domain",
			Value: "",
			Usage: "PAM email domain",
		},
		&cli.BoolFlag{
			Name:  "skip-local-2fa",
			Usage: "Skip 2FA to log on.",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "active",
			Usage: "This Authentication Source is Activated.",
			Value: true,
		},
	}
}

func microcmdAuthAddPAM() *cli.Command {
	return &cli.Command{
		Name:   "add-pam",
		Usage:  "Add new PAM authentication source",
		Before: noDanglingArgs,
		Action: newAuthService().addPAM,
		Flags:  pamCLIFlags(),
	}
}

func microcmdAuthUpdatePAM() *cli.Command {
	return &cli.Command{
		Name:   "update-pam",
		Usage:  "Update existing PAM authentication source",
		Before: noDanglingArgs,
		Action: newAuthService().updatePAM,
		Flags:  append(pamCLIFlags()[:1], append([]cli.Flag{idFlag()}, pamCLIFlags()[1:]...)...),
	}
}

func parsePAMConfig(_ context.Context, c *cli.Command) *pam.Source {
	return &pam.Source{
		ServiceName:    c.String("service-name"),
		EmailDomain:    c.String("email-domain"),
		SkipLocalTwoFA: c.Bool("skip-local-2fa"),
	}
}

func (a *authService) addPAM(ctx context.Context, c *cli.Command) error {
	ctx, cancel := installSignals(ctx)
	defer cancel()

	if err := a.initDB(ctx); err != nil {
		return err
	}

	if !c.IsSet("name") || len(c.String("name")) == 0 {
		return errors.New("name must be set")
	}
	if !c.IsSet("service-name") || len(c.String("service-name")) == 0 {
		return errors.New("service-name must be set")
	}
	active := true
	if c.IsSet("active") {
		active = c.Bool("active")
	}

	config := parsePAMConfig(ctx, c)

	return a.createAuthSource(ctx, &auth_model.Source{
		Type:     auth_model.PAM,
		Name:     c.String("name"),
		IsActive: active,
		Cfg:      config,
	})
}

func (a *authService) updatePAM(ctx context.Context, c *cli.Command) error {
	if !c.IsSet("id") {
		return errors.New("--id flag is missing")
	}

	ctx, cancel := installSignals(ctx)
	defer cancel()

	if err := a.initDB(ctx); err != nil {
		return err
	}

	source, err := a.getAuthSource(ctx, c, auth_model.PAM)
	if err != nil {
		return err
	}

	pamConfig := source.Cfg.(*pam.Source)

	if c.IsSet("name") {
		source.Name = c.String("name")
	}

	if c.IsSet("service-name") {
		pamConfig.ServiceName = c.String("service-name")
	}

	if c.IsSet("email-domain") {
		pamConfig.EmailDomain = c.String("email-domain")
	}

	if c.IsSet("skip-local-2fa") {
		pamConfig.SkipLocalTwoFA = c.Bool("skip-local-2fa")
	}

	if c.IsSet("active") {
		source.IsActive = c.Bool("active")
	}

	source.Cfg = pamConfig

	return a.updateAuthSource(ctx, source)
}
