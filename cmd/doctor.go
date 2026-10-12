// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"fmt"
	"image"
	golog "log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"forgejo.org/models/db"
	git_model "forgejo.org/models/git"
	"forgejo.org/models/gitea_migrations"
	migrate_base "forgejo.org/models/gitea_migrations/base"
	repo_model "forgejo.org/models/repo"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/avatarstore"
	"forgejo.org/modules/container"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/storage"
	"forgejo.org/services/doctor"

	"github.com/urfave/cli/v3"
	"xorm.io/builder"
)

// CmdDoctor represents the available doctor sub-command.
func cmdDoctor() *cli.Command {
	return &cli.Command{
		Name:        "doctor",
		Usage:       "Diagnose and optionally fix problems, convert or re-create database tables",
		Description: "A command to diagnose problems with the current Forgejo instance according to the given configuration. Some problems can optionally be fixed by modifying the database or data storage.",

		Commands: []*cli.Command{
			cmdDoctorCheck(),
			cmdRecreateTable(),
			cmdDoctorConvert(),
			cmdAvatarStripExif(),
			cmdCleanupCommitStatuses(),
			cmdResizeAvatars(),
		},
	}
}

func cmdDoctorCheck() *cli.Command {
	return &cli.Command{
		Name:        "check",
		Usage:       "Diagnose and optionally fix problems",
		Description: "A command to diagnose problems with the current Forgejo instance according to the given configuration. Some problems can optionally be fixed by modifying the database or data storage.",
		Before:      noDanglingArgs,
		Action:      runDoctorCheck,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "list",
				Usage: "List the available checks",
			},
			&cli.BoolFlag{
				Name:  "default",
				Usage: "Run the default checks (if neither --run or --all is set, this is the default behaviour)",
			},
			&cli.StringSliceFlag{
				Name:  "run",
				Usage: "Run the provided checks - (if --default is set, the default checks will also run)",
			},
			&cli.BoolFlag{
				Name:  "all",
				Usage: "Run all the available checks",
			},
			&cli.BoolFlag{
				Name:  "fix",
				Usage: "Automatically fix what we can",
			},
			&cli.StringFlag{
				Name:  "log-file",
				Usage: `Name of the log file (no verbose log output by default). Set to "-" to output to stdout`,
			},
			&cli.BoolFlag{
				Name:    "color",
				Aliases: []string{"H"},
				Usage:   "Use color for outputted information",
			},
		},
	}
}

func cmdRecreateTable() *cli.Command {
	return &cli.Command{
		Name:      "recreate-table",
		Usage:     "Recreate tables from XORM definitions and copy the data.",
		ArgsUsage: "[TABLE]... : (TABLEs to recreate - leave blank for all)",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Print SQL commands sent",
			},
		},
		Description: `The database definitions Forgejo uses change across versions, sometimes changing default values and leaving old unused columns.

This command will cause Xorm to recreate tables, copying over the data and deleting the old table.

You should back-up your database before doing this and ensure that your database is up-to-date first.`,
		Action: runRecreateTable,
	}
}

func cmdAvatarStripExif() *cli.Command {
	return &cli.Command{
		Name:  "avatar-strip-exif",
		Usage: "Strip EXIF metadata from all images in the avatar storage [unsupported]",
		Description: `Stripping EXIF metadata is not currently supported. The capability was
available in previous Forgejo releases, but has been removed. This command
may be re-enabled in the future if the capability can be supported again.`,
		Before: noDanglingArgs,
		Action: runAvatarStripExif,
	}
}

func cmdResizeAvatars() *cli.Command {
	return &cli.Command{
		Name:  "avatar-resize",
		Usage: "Generate resized versions of user or repository avatars",
		Description: `Forgejo serves small versions of avatars for inclusion in the web UI.

Those rescaled versions are computed on-demand and cached in the avatar storage.

This command pre-computes rescaled versions of avatars ahead of time.`,
		Before: noDanglingArgs,
		Action: runAvatarResize,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "user",
				Usage: "Resize the user avatars",
			},
			&cli.BoolFlag{
				Name:  "repository",
				Usage: "Resize the repository avatars",
			},
		},
	}
}

func cmdCleanupCommitStatuses() *cli.Command {
	return &cli.Command{
		Name:  "cleanup-commit-status",
		Usage: "Cleanup extra records in commit_status table",
		Description: `Forgejo suffered from a bug which caused the creation of more entries in the
"commit_status" table than necessary. This operation removes the redundant
data caused by the bug. Removing this data is almost always safe.
These redundant records can be accessed by users through the API, making it
possible, but unlikely, that removing it could have an impact to
integrating services (API: /repos/{owner}/{repo}/commits/{ref}/statuses).

It is safe to run while Forgejo is online.

On very large Forgejo instances, the performance of operation will improve
if the buffer-size option is used with large values. Approximately 130 MB of
memory is required for every 100,000 records in the buffer.

Bug reference: https://codeberg.org/forgejo/forgejo/issues/10671
`,

		Before: multipleBefore(noDanglingArgs, PrepareConsoleLoggerLevel(log.INFO)),
		Action: runCleanupCommitStatus,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "Show process details",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Report statistics from the operation but do not modify the database",
			},
			&cli.IntFlag{
				Name:  "buffer-size",
				Usage: "Record count per query while iterating records; larger values are typically faster but use more memory",
				// See IterateByKeyset's documentation for performance notes which led to the choice of the default
				// buffer size for this operation.
				Value: 100000,
			},
			&cli.IntFlag{
				Name:  "delete-chunk-size",
				Usage: "Number of records to delete per DELETE query",
				Value: 1000,
			},
		},
	}
}

func runRecreateTable(stdCtx context.Context, ctx *cli.Command) error {
	stdCtx, cancel := installSignals(stdCtx)
	defer cancel()

	// Redirect the default golog to here
	golog.SetFlags(0)
	golog.SetPrefix("")
	golog.SetOutput(log.LoggerToWriter(log.GetLogger(log.DEFAULT).Info))

	debug := ctx.Bool("debug")
	setting.MustInstalled()
	setting.LoadDBSetting()

	if debug {
		setting.InitSQLLoggersForCli(log.DEBUG)
	} else {
		setting.InitSQLLoggersForCli(log.INFO)
	}

	setting.Database.LogSQL = debug
	if err := db.InitEngine(stdCtx); err != nil {
		fmt.Println(err)
		fmt.Println("Check if you are using the right config file. You can use a --config directive to specify one.")
		return nil
	}

	args := ctx.Args()
	names := make([]string, 0, ctx.NArg())
	for i := range ctx.NArg() {
		names = append(names, args.Get(i))
	}

	beans, err := db.NamesToBean(names...)
	if err != nil {
		return err
	}
	recreateTables := migrate_base.RecreateTables(beans...)

	return db.InitEngineWithMigration(stdCtx, func(x db.Engine) error {
		engine, err := db.GetMasterEngine(x)
		if err != nil {
			return err
		}

		if err := gitea_migrations.EnsureUpToDate(engine); err != nil {
			return err
		}

		return recreateTables(engine)
	})
}

func setupDoctorDefaultLogger(ctx *cli.Command, colorize bool) {
	// Silence the default loggers
	setupConsoleLogger(log.FATAL, log.CanColorStderr, os.Stderr)

	logFile := ctx.String("log-file")
	switch logFile {
	case "":
		return // if no doctor log-file is set, do not show any log from default logger
	case "-":
		setupConsoleLogger(log.TRACE, colorize, os.Stdout)
	default:
		logFile, _ = filepath.Abs(logFile)
		writeMode := log.WriterMode{Level: log.TRACE, WriterOption: log.WriterFileOption{FileName: logFile}}
		writer, err := log.NewEventWriter("console-to-file", "file", writeMode)
		if err != nil {
			log.FallbackErrorf("unable to create file log writer: %v", err)
			return
		}
		log.GetManager().GetLogger(log.DEFAULT).ReplaceAllWriters(writer)
	}
}

func runDoctorCheck(stdCtx context.Context, ctx *cli.Command) error {
	stdCtx, cancel := installSignals(stdCtx)
	defer cancel()

	colorize := log.CanColorStdout
	if ctx.IsSet("color") {
		colorize = ctx.Bool("color")
	}

	setupDoctorDefaultLogger(ctx, colorize)

	// Finally redirect the default golang's log to here
	golog.SetFlags(0)
	golog.SetPrefix("")
	golog.SetOutput(log.LoggerToWriter(log.GetLogger(log.DEFAULT).Info))

	if ctx.IsSet("list") {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		_, _ = w.Write([]byte("Default\tName\tTitle\n"))
		doctor.SortChecks(doctor.Checks)
		for _, check := range doctor.Checks {
			if check.IsDefault {
				_, _ = w.Write([]byte{'*'})
			}
			_, _ = w.Write([]byte{'\t'})
			_, _ = w.Write([]byte(check.Name))
			_, _ = w.Write([]byte{'\t'})
			_, _ = w.Write([]byte(check.Title))
			_, _ = w.Write([]byte{'\n'})
		}
		return w.Flush()
	}

	var checks []*doctor.Check
	if ctx.Bool("all") {
		checks = make([]*doctor.Check, len(doctor.Checks))
		copy(checks, doctor.Checks)
	} else if ctx.IsSet("run") {
		addDefault := ctx.Bool("default")
		runNamesSet := container.SetOf(ctx.StringSlice("run")...)
		for _, check := range doctor.Checks {
			if (addDefault && check.IsDefault) || runNamesSet.Contains(check.Name) {
				checks = append(checks, check)
				runNamesSet.Remove(check.Name)
			}
		}
		if len(runNamesSet) > 0 {
			return fmt.Errorf("unknown checks: %q", strings.Join(runNamesSet.Values(), ","))
		}
	} else {
		for _, check := range doctor.Checks {
			if check.IsDefault {
				checks = append(checks, check)
			}
		}
	}
	return doctor.RunChecks(stdCtx, colorize, ctx.Bool("fix"), checks)
}

func runAvatarStripExif(ctx context.Context, c *cli.Command) error {
	log.Warn("avatar-strip-exif is not currently supported.")
	return nil
}

func precomputeResizedAvatars(imgStorage storage.ObjectStorage, imgPath string, maxOriginSize int64) error {
	// Load the avatar
	avatarBytes, err := imgStorage.Open(imgPath)
	if err != nil {
		return err
	}
	meta, err := avatarBytes.Stat()
	if err != nil {
		return err
	}
	// If the avatar is small enough, don't compute resized versions for it.
	// This makes it possible to preserve animated avatars when they are small enough.
	if meta.Size() < maxOriginSize {
		return nil
	}
	img, _, err := image.Decode(avatarBytes)
	if err != nil {
		return err
	}
	return avatarstore.PrecomputeResizedAvatars(imgStorage, img, imgPath)
}

func runAvatarResize(ctx context.Context, c *cli.Command) error {
	ctx, cancel := installSignals(ctx)
	defer cancel()

	if err := initDB(ctx); err != nil {
		return err
	}

	if err := storage.Init(); err != nil {
		return err
	}

	runUser := c.Bool("user")
	runRepo := c.Bool("repository")
	return RunAvatarResize(ctx, runUser, runRepo)
}

func RunAvatarResize(ctx context.Context, runUser, runRepo bool) error {
	if !runUser && !runRepo {
		return fmt.Errorf("at least one of --user or --repository should be provided")
	}

	if runUser {
		log.Info("Resizing user avatars")
		if err := db.Iterate(
			ctx,
			builder.Neq{"avatar": ""},
			func(ctx context.Context, user *user_model.User) error {
				return precomputeResizedAvatars(storage.Avatars, user.Avatar, setting.Avatar.MaxOriginSize)
			},
		); err != nil {
			return err
		}
	}

	if runRepo {
		log.Info("Resizing repository avatars")
		if err := db.Iterate(
			ctx,
			builder.Neq{"avatar": ""},
			func(ctx context.Context, repo *repo_model.Repository) error {
				return precomputeResizedAvatars(storage.RepoAvatars, repo.Avatar, setting.Avatar.MaxOriginSize)
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func runCleanupCommitStatus(ctx context.Context, cli *cli.Command) error {
	ctx, cancel := installSignals(ctx)
	defer cancel()

	if err := initDB(ctx); err != nil {
		return err
	}

	bufferSize := cli.Int("buffer-size")
	deleteChunkSize := cli.Int("delete-chunk-size")
	dryRun := cli.Bool("dry-run")
	log.Debug("bufferSize = %d, deleteChunkSize = %d, dryRun = %v", bufferSize, deleteChunkSize, dryRun)

	return git_model.CleanupCommitStatus(ctx, bufferSize, deleteChunkSize, dryRun)
}
