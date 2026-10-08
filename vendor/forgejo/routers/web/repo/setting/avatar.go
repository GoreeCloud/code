// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"fmt"

	"forgejo.org/modules/web"
	"forgejo.org/services/context"
	"forgejo.org/services/forms"
	repo_service "forgejo.org/services/repository"
)

// UpdateAvatarSetting update repo's avatar
func UpdateAvatarSetting(ctx *context.Context, form forms.AvatarForm) error {
	data, err := forms.ReadAvatar(form.Avatar, ctx.Locale)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	if err := repo_service.UploadAvatar(ctx, ctx.Repo.Repository, data); err != nil {
		return fmt.Errorf("UploadAvatar: %w", err)
	}
	return nil
}

// SettingsAvatar save new POSTed repository avatar
func SettingsAvatar(ctx *context.Context) {
	form := web.GetForm(ctx).(*forms.AvatarForm)
	form.Source = forms.AvatarLocal
	if err := UpdateAvatarSetting(ctx, *form); err != nil {
		ctx.Flash.Error(err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("repo.settings.update_avatar_success"))
	}
	ctx.Redirect(ctx.Repo.RepoLink + "/settings")
}

// SettingsDeleteAvatar delete repository avatar
func SettingsDeleteAvatar(ctx *context.Context) {
	if err := repo_service.DeleteAvatar(ctx, ctx.Repo.Repository); err != nil {
		ctx.Flash.Error(fmt.Sprintf("DeleteAvatar: %v", err))
	}
	ctx.JSONRedirect(ctx.Repo.RepoLink + "/settings")
}
