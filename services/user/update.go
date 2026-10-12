// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"context"
	"fmt"

	"forgejo.org/models"
	auth_model "forgejo.org/models/auth"
	user_model "forgejo.org/models/user"
	password_module "forgejo.org/modules/auth/password"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/structs"
	"forgejo.org/services/mailer"
)

type UpdateOptions struct {
	KeepEmailPrivate             optional.Option[bool]
	FullName                     optional.Option[string]
	Website                      optional.Option[string]
	Location                     optional.Option[string]
	Description                  optional.Option[string]
	Pronouns                     optional.Option[string]
	AllowGitHook                 optional.Option[bool]
	AllowImportLocal             optional.Option[bool]
	MaxRepoCreation              optional.Option[int]
	IsRestricted                 optional.Option[bool]
	Visibility                   optional.Option[structs.VisibleType]
	KeepActivityPrivate          optional.Option[bool]
	Language                     optional.Option[string]
	Theme                        optional.Option[string]
	DiffViewStyle                optional.Option[string]
	AllowCreateOrganization      optional.Option[bool]
	IsActive                     optional.Option[bool]
	IsAdmin                      optional.Option[bool]
	EmailNotificationsPreference optional.Option[string]
	SetLastLogin                 bool
	RepoAdminChangeTeamAccess    optional.Option[bool]
	EnableRepoUnitHints          optional.Option[bool]
	KeepPronounsPrivate          optional.Option[bool]
}

func UpdateUser(ctx context.Context, u *user_model.User, opts *UpdateOptions) error {
	cols := make([]string, 0, 20)

	if has, value := opts.KeepEmailPrivate.Get(); has {
		u.KeepEmailPrivate = value
		cols = append(cols, "keep_email_private")
	}
	if has, value := opts.FullName.Get(); has {
		u.FullName = value
		cols = append(cols, "full_name")
	}
	if has, value := opts.Pronouns.Get(); has {
		u.Pronouns = value
		cols = append(cols, "pronouns")
	}
	if has, value := opts.Website.Get(); has {
		u.Website = value
		cols = append(cols, "website")
	}
	if has, value := opts.Location.Get(); has {
		u.Location = value
		cols = append(cols, "location")
	}
	if has, value := opts.Description.Get(); has {
		u.Description = value
		cols = append(cols, "description")
	}
	if has, value := opts.Language.Get(); has {
		u.Language = value
		cols = append(cols, "language")
	}
	if has, value := opts.Theme.Get(); has {
		u.Theme = value
		cols = append(cols, "theme")
	}
	if has, value := opts.DiffViewStyle.Get(); has {
		u.DiffViewStyle = value
		cols = append(cols, "diff_view_style")
	}
	if has, value := opts.EnableRepoUnitHints.Get(); has {
		u.EnableRepoUnitHints = value
		cols = append(cols, "enable_repo_unit_hints")
	}
	if has, value := opts.KeepPronounsPrivate.Get(); has {
		u.KeepPronounsPrivate = value
		cols = append(cols, "keep_pronouns_private")
	}
	if has, value := opts.AllowGitHook.Get(); has {
		u.AllowGitHook = value
		cols = append(cols, "allow_git_hook")
	}
	if has, value := opts.AllowImportLocal.Get(); has {
		u.AllowImportLocal = value
		cols = append(cols, "allow_import_local")
	}
	if has, value := opts.MaxRepoCreation.Get(); has {
		u.MaxRepoCreation = value
		cols = append(cols, "max_repo_creation")
	}
	if has, value := opts.IsActive.Get(); has {
		u.IsActive = value
		cols = append(cols, "is_active")
	}
	if has, value := opts.IsRestricted.Get(); has {
		u.IsRestricted = value
		cols = append(cols, "is_restricted")
	}
	if has, value := opts.IsAdmin.Get(); has {
		if !value && user_model.IsLastAdminUser(ctx, u) {
			return models.ErrDeleteLastAdminUser{UID: u.ID}
		}
		u.IsAdmin = value
		cols = append(cols, "is_admin")
	}
	if has, value := opts.Visibility.Get(); has {
		if !u.IsOrganization() && !setting.Service.AllowedUserVisibilityModesSlice.IsAllowedVisibility(value) {
			return fmt.Errorf("visibility mode not allowed: %s", value.String())
		}
		u.Visibility = value
		cols = append(cols, "visibility")
	}
	if has, value := opts.KeepActivityPrivate.Get(); has {
		u.KeepActivityPrivate = value
		cols = append(cols, "keep_activity_private")
	}
	if has, value := opts.AllowCreateOrganization.Get(); has {
		u.AllowCreateOrganization = value
		cols = append(cols, "allow_create_organization")
	}
	if has, value := opts.RepoAdminChangeTeamAccess.Get(); has {
		u.RepoAdminChangeTeamAccess = value
		cols = append(cols, "repo_admin_change_team_access")
	}
	if has, value := opts.EmailNotificationsPreference.Get(); has {
		u.EmailNotificationsPreference = value
		cols = append(cols, "email_notifications_preference")
	}
	if opts.SetLastLogin {
		u.SetLastLogin()
		cols = append(cols, "last_login_unix")
	}

	return user_model.UpdateUserCols(ctx, u, cols...)
}

type UpdateAuthOptions struct {
	LoginSource        optional.Option[int64]
	LoginName          optional.Option[string]
	Password           optional.Option[string]
	MustChangePassword optional.Option[bool]
	ProhibitLogin      optional.Option[bool]
}

func UpdateAuth(ctx context.Context, u *user_model.User, opts *UpdateAuthOptions) error {
	if has, value := opts.LoginSource.Get(); has {
		source, err := auth_model.GetSourceByID(ctx, value)
		if err != nil {
			return err
		}

		u.LoginType = source.Type
		u.LoginSource = source.ID
	}
	if has, value := opts.LoginName.Get(); has {
		u.LoginName = value
	}

	if has, value := opts.Password.Get(); has && (u.IsLocal() || u.IsOAuth2()) {
		password := value

		if len(password) < setting.MinPasswordLength {
			return password_module.ErrMinLength
		}
		if !password_module.IsComplexEnough(password) {
			return password_module.ErrComplexity
		}
		if err := password_module.IsPwned(ctx, password); err != nil {
			return err
		}

		if err := u.SetPassword(password); err != nil {
			return err
		}
	}

	if has, value := opts.MustChangePassword.Get(); has {
		u.MustChangePassword = value
	}
	if has, value := opts.ProhibitLogin.Get(); has {
		u.ProhibitLogin = value
	}

	if err := user_model.UpdateUserCols(ctx, u, "login_type", "login_source", "login_name", "passwd", "passwd_hash_algo", "salt", "must_change_password", "prohibit_login"); err != nil {
		return err
	}

	if opts.Password.Has() {
		return mailer.SendPasswordChange(u)
	}

	return nil
}
