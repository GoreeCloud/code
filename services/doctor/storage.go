// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"errors"
	"io/fs"
	"strings"

	"forgejo.org/models/git"
	"forgejo.org/models/packages"
	"forgejo.org/models/repo"
	"forgejo.org/models/user"
	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	packages_module "forgejo.org/modules/packages"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/storage"
	"forgejo.org/modules/util"
)

type commonStorageCheckOptions struct {
	storer     storage.ObjectStorage
	isOrphaned func(path string, obj storage.Object, stat fs.FileInfo) (bool, error)
	name       string
}

func commonCheckStorage(logger log.Logger, autofix bool, opts *commonStorageCheckOptions) error {
	totalCount, orphanedCount := 0, 0
	totalSize, orphanedSize := int64(0), int64(0)

	var pathsToDelete []string
	if err := opts.storer.IterateObjects("", func(p string, obj storage.Object) error {
		defer obj.Close()

		totalCount++
		stat, err := obj.Stat()
		if err != nil {
			return err
		}
		totalSize += stat.Size()

		orphaned, err := opts.isOrphaned(p, obj, stat)
		if err != nil {
			return err
		}
		if orphaned {
			orphanedCount++
			orphanedSize += stat.Size()
			if autofix {
				pathsToDelete = append(pathsToDelete, p)
			}
		}
		return nil
	}); err != nil {
		logger.Error("Error whilst iterating %s storage: %v", opts.name, err)
		return err
	}

	if orphanedCount > 0 {
		if autofix {
			var deletedNum int
			for _, p := range pathsToDelete {
				if err := opts.storer.Delete(p); err != nil {
					log.Error("Error whilst deleting %s from %s storage: %v", p, opts.name, err)
				} else {
					deletedNum++
				}
			}
			logger.Info("Deleted %d/%d orphaned %s(s)", deletedNum, orphanedCount, opts.name)
		} else {
			logger.Warn("Found %d/%d (%s/%s) orphaned %s(s)", orphanedCount, totalCount, base.FileSize(orphanedSize), base.FileSize(totalSize), opts.name)
		}
	} else {
		logger.Info("Found %d (%s) %s(s)", totalCount, base.FileSize(totalSize), opts.name)
	}
	return nil
}

type CheckStorageOptions struct {
	All          bool
	Attachments  bool
	LFS          bool
	Avatars      bool
	RepoAvatars  bool
	RepoArchives bool
	Packages     bool
}

// CheckStorage will return a doctor check function to check the requested storage types for "orphaned" stored object/files and optionally delete them
func CheckStorage(opts *CheckStorageOptions) func(ctx context.Context, logger log.Logger, autofix bool) error {
	return func(ctx context.Context, logger log.Logger, autofix bool) error {
		if err := storage.Init(); err != nil {
			logger.Error("storage.Init failed: %v", err)
			return err
		}

		if opts.Attachments || opts.All {
			if err := commonCheckStorage(logger, autofix,
				&commonStorageCheckOptions{
					storer: storage.Attachments,
					isOrphaned: func(path string, obj storage.Object, stat fs.FileInfo) (bool, error) {
						exists, err := repo.ExistAttachmentsByUUID(ctx, stat.Name())
						return !exists, err
					},
					name: "attachment",
				}); err != nil {
				return err
			}
		}

		if opts.LFS || opts.All {
			if !setting.LFS.StartServer {
				logger.Info("LFS isn't enabled (skipped)")
				return nil
			}
			if err := commonCheckStorage(logger, autofix,
				&commonStorageCheckOptions{
					storer: storage.LFS,
					isOrphaned: func(path string, obj storage.Object, stat fs.FileInfo) (bool, error) {
						// The oid of an LFS stored object is the name but with all the path.Separators removed
						oid := strings.ReplaceAll(path, "/", "")
						exists, err := git.ExistsLFSObject(ctx, oid)
						return !exists, err
					},
					name: "LFS file",
				}); err != nil {
				return err
			}
		}

		if opts.Avatars || opts.All {
			if err := commonCheckStorage(logger, autofix,
				&commonStorageCheckOptions{
					storer: storage.Avatars,
					isOrphaned: func(path string, obj storage.Object, stat fs.FileInfo) (bool, error) {
						// The path is either just a hash, if the file is an original avatar uploaded by the user,
						// or of the form "resized/<size>/<hash>" if it is a resized version of the avatar.
						// In both cases, we retain the file if and only if the hash corresponds to an avatar
						// of an existing user.
						pathParts := strings.Split(path, "/")
						hash := pathParts[len(pathParts)-1]
						exists, err := user.ExistsWithAvatarAtStoragePath(ctx, hash)
						return !exists, err
					},
					name: "avatar",
				}); err != nil {
				return err
			}
		}

		if opts.RepoAvatars || opts.All {
			if err := commonCheckStorage(logger, autofix,
				&commonStorageCheckOptions{
					storer: storage.RepoAvatars,
					isOrphaned: func(path string, obj storage.Object, stat fs.FileInfo) (bool, error) {
						// See the comment above to explain the handling of original and resized avatars.
						pathParts := strings.Split(path, "/")
						hash := pathParts[len(pathParts)-1]
						exists, err := repo.ExistsWithAvatarAtStoragePath(ctx, hash)
						return !exists, err
					},
					name: "repo avatar",
				}); err != nil {
				return err
			}
		}

		if opts.RepoArchives || opts.All {
			if err := commonCheckStorage(logger, autofix,
				&commonStorageCheckOptions{
					storer: storage.RepoArchives,
					isOrphaned: func(path string, obj storage.Object, stat fs.FileInfo) (bool, error) {
						exists, err := repo.ExistsRepoArchiverWithStoragePath(ctx, path)
						if err == nil || errors.Is(err, util.ErrInvalidArgument) {
							// invalid arguments mean that the object is not a valid repo archiver and it should be removed
							return !exists, nil
						}
						return !exists, err
					},
					name: "repo archive",
				}); err != nil {
				return err
			}
		}

		if opts.Packages || opts.All {
			if !setting.Packages.Enabled {
				logger.Info("Packages isn't enabled (skipped)")
				return nil
			}
			if err := commonCheckStorage(logger, autofix,
				&commonStorageCheckOptions{
					storer: storage.Packages,
					isOrphaned: func(path string, obj storage.Object, stat fs.FileInfo) (bool, error) {
						key, err := packages_module.RelativePathToKey(path)
						if err != nil {
							// If there is an error here then the relative path does not match a valid package
							// Therefore it is orphaned by default
							return true, nil
						}

						exists, err := packages.ExistPackageBlobWithSHA(ctx, string(key))

						return !exists, err
					},
					name: "package blob",
				}); err != nil {
				return err
			}
		}

		return nil
	}
}

func init() {
	Register(&Check{
		Title:                      "Check if there are orphaned storage files",
		Name:                       "storages",
		IsDefault:                  false,
		Run:                        CheckStorage(&CheckStorageOptions{All: true}),
		AbortIfFailed:              false,
		SkipDatabaseInitialization: false,
		Priority:                   1,
	})

	Register(&Check{
		Title:                      "Check if there are orphaned attachments in storage",
		Name:                       "storage-attachments",
		IsDefault:                  false,
		Run:                        CheckStorage(&CheckStorageOptions{Attachments: true}),
		AbortIfFailed:              false,
		SkipDatabaseInitialization: false,
		Priority:                   1,
	})

	Register(&Check{
		Title:                      "Check if there are orphaned lfs files in storage",
		Name:                       "storage-lfs",
		IsDefault:                  false,
		Run:                        CheckStorage(&CheckStorageOptions{LFS: true}),
		AbortIfFailed:              false,
		SkipDatabaseInitialization: false,
		Priority:                   1,
	})

	Register(&Check{
		Title:                      "Check if there are orphaned avatars in storage",
		Name:                       "storage-avatars",
		IsDefault:                  false,
		Run:                        CheckStorage(&CheckStorageOptions{Avatars: true, RepoAvatars: true}),
		AbortIfFailed:              false,
		SkipDatabaseInitialization: false,
		Priority:                   1,
	})

	Register(&Check{
		Title:                      "Check if there are orphaned archives in storage",
		Name:                       "storage-archives",
		IsDefault:                  false,
		Run:                        CheckStorage(&CheckStorageOptions{RepoArchives: true}),
		AbortIfFailed:              false,
		SkipDatabaseInitialization: false,
		Priority:                   1,
	})

	Register(&Check{
		Title:                      "Check if there are orphaned package blobs in storage",
		Name:                       "storage-packages",
		IsDefault:                  false,
		Run:                        CheckStorage(&CheckStorageOptions{Packages: true}),
		AbortIfFailed:              false,
		SkipDatabaseInitialization: false,
		Priority:                   1,
	})
}
