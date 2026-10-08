// Copyright 2018 The Gitea Authors. All rights reserved.
// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// Package cmd provides subcommands to the gitea binary - such as "web" or
// "admin".
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"forgejo.org/models/db"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/util"

	"github.com/urfave/cli/v3"
)

// argsSet checks that all the required arguments are set. args is a list of
// arguments that must be set in the passed Context.
func argsSet(c *cli.Command, args ...string) error {
	for _, a := range args {
		if !c.IsSet(a) {
			return errors.New(a + " is not set")
		}

		if s, ok := c.Value(a).(string); ok {
			if util.IsEmptyString(s) {
				return errors.New(a + " is required")
			}
		}
	}
	return nil
}

// When a CLI command is intended to be used only with flags and no other arbitrary args, noDanglingArgs will validate
// the end-user's usage.
func noDanglingArgs(ctx context.Context, c *cli.Command) (context.Context, error) {
	if c.Args().Len() != 0 {
		args := c.Args().Slice()
		if slices.Contains(args, "false") {
			println("Hint: boolean false must be specified as a single arg, eg. '--restricted=false', not '--restricted false'")
		}
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(c.Args().Slice(), ", "))
	}
	return nil, nil
}

// confirm waits for user input which confirms an action
func confirm() (bool, error) {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, errors.New(response + " isn't a correct confirmation string")
	}
}

func initDB(ctx context.Context) error {
	setting.MustInstalled()
	setting.LoadDBSetting()
	setting.InitSQLLoggersForCli(log.INFO)

	if setting.Database.Type == "" {
		log.Fatal(`Database settings are missing from the configuration file: %q.
Ensure you are running in the correct environment or set the correct configuration file with -c.
If this is the intended configuration file complete the [database] section.`, setting.CustomConf)
	}
	if err := db.InitEngine(ctx); err != nil {
		return fmt.Errorf("unable to initialize the database using the configuration in %q. Error: %w", setting.CustomConf, err)
	}
	return nil
}

// installSignals returns a context that's cancelled on the SIGINT and SIGTERM signals or if the passed ctx is cancelled.
func installSignals(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
}

func setupConsoleLogger(level log.Level, colorize bool, out io.Writer) {
	if out != os.Stdout && out != os.Stderr {
		panic("setupConsoleLogger can only be used with os.Stdout or os.Stderr")
	}

	writeMode := log.WriterMode{
		Level:        level,
		Colorize:     colorize,
		WriterOption: log.WriterConsoleOption{Stderr: out == os.Stderr},
	}
	writer := log.NewEventWriterConsole("console-default", writeMode)
	log.GetManager().GetLogger(log.DEFAULT).ReplaceAllWriters(writer)
}

func globalBool(c *cli.Command, name string) bool {
	for _, ctx := range c.Lineage() {
		if ctx.Bool(name) {
			return true
		}
	}
	return false
}

// PrepareConsoleLoggerLevel by default, use INFO level for console logger, but some sub-commands (for git/ssh protocol) shouldn't output any log to stdout.
// Any log appears in git stdout pipe will break the git protocol, eg: client can't push and hangs forever.
func PrepareConsoleLoggerLevel(defaultLevel log.Level) func(ctx context.Context, cli *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cli *cli.Command) (context.Context, error) {
		level := defaultLevel
		if globalBool(cli, "quiet") {
			level = log.FATAL
		}
		if globalBool(cli, "debug") || globalBool(cli, "verbose") {
			level = log.TRACE
		}
		log.SetConsoleLogger(log.DEFAULT, "console-default", level)
		return ctx, nil
	}
}

func multipleBefore(beforeFuncs ...cli.BeforeFunc) cli.BeforeFunc {
	return func(ctx context.Context, cli *cli.Command) (context.Context, error) {
		for _, beforeFunc := range beforeFuncs {
			bctx, err := beforeFunc(ctx, cli)
			if err != nil {
				return bctx, err
			}
			if bctx != nil {
				ctx = bctx
			}
		}
		return ctx, nil
	}
}
