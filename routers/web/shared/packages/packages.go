// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package packages

import (
	"errors"
	"fmt"
	"net/http"

	packages_model "forgejo.org/models/packages"
	repo_model "forgejo.org/models/repo"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	"forgejo.org/modules/util"
	"forgejo.org/modules/web"
	"forgejo.org/services/context"
	"forgejo.org/services/forms"
	cargo_service "forgejo.org/services/packages/cargo"
	cleanup_service "forgejo.org/services/packages/cleanup"
)

func SetPackagesContext(ctx *context.Context, owner *user_model.User) {
	pcrs, err := packages_model.GetCleanupRulesByOwner(ctx, owner.ID)
	if err != nil {
		ctx.ServerError("GetCleanupRulesByOwner", err)
		return
	}

	ctx.Data["CleanupRules"] = pcrs

	ctx.Data["CargoIndexExists"], err = repo_model.IsRepositoryModelExist(ctx, owner, cargo_service.IndexRepositoryName)
	if err != nil {
		ctx.ServerError("IsRepositoryModelExist", err)
		return
	}
}

func SetRuleAddContext(ctx *context.Context) {
	setRuleEditContext(ctx, nil)
}

func SetRuleEditContext(ctx *context.Context, owner *user_model.User) {
	pcr := getCleanupRuleByContext(ctx, owner)
	if pcr == nil {
		return
	}

	setRuleEditContext(ctx, pcr)
}

func setRuleEditContext(ctx *context.Context, pcr *packages_model.PackageCleanupRule) {
	ctx.Data["IsEditRule"] = pcr != nil

	if pcr == nil {
		pcr = &packages_model.PackageCleanupRule{}
	}
	ctx.Data["CleanupRule"] = pcr
	ctx.Data["AvailableTypes"] = packages_model.TypeList
}

func PerformRuleAddPost(ctx *context.Context, owner *user_model.User, redirectURL string, template base.TplName) {
	performRuleEditPost(ctx, owner, nil, redirectURL, template)
}

func PerformRuleEditPost(ctx *context.Context, owner *user_model.User, redirectURL string, template base.TplName) {
	pcr := getCleanupRuleByContext(ctx, owner)
	if pcr == nil {
		return
	}

	form := web.GetForm(ctx).(*forms.PackageCleanupRuleForm)

	if form.Action == "remove" {
		if err := packages_model.DeleteCleanupRuleByID(ctx, pcr.ID); err != nil {
			ctx.ServerError("DeleteCleanupRuleByID", err)
			return
		}

		ctx.Flash.Success(ctx.Tr("packages.owner.settings.cleanuprules.success.delete"))
		ctx.Redirect(redirectURL)
	} else {
		performRuleEditPost(ctx, owner, pcr, redirectURL, template)
	}
}

func performRuleEditPost(ctx *context.Context, owner *user_model.User, pcr *packages_model.PackageCleanupRule, redirectURL string, template base.TplName) {
	isEditRule := pcr != nil

	if pcr == nil {
		pcr = &packages_model.PackageCleanupRule{}
	}

	form := web.GetForm(ctx).(*forms.PackageCleanupRuleForm)

	pcr.Enabled = form.Enabled
	pcr.OwnerID = owner.ID
	pcr.KeepCount = form.KeepCount
	pcr.KeepPattern = form.KeepPattern
	pcr.RemoveDays = form.RemoveDays
	pcr.RemovePattern = form.RemovePattern
	pcr.MatchFullName = form.MatchFullName

	ctx.Data["IsEditRule"] = isEditRule
	ctx.Data["CleanupRule"] = pcr
	ctx.Data["AvailableTypes"] = packages_model.TypeList

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, template)
		return
	}

	if isEditRule {
		if err := packages_model.UpdateCleanupRule(ctx, pcr); err != nil {
			ctx.ServerError("UpdateCleanupRule", err)
			return
		}
	} else {
		pcr.Type = packages_model.Type(form.Type)

		if has, err := packages_model.HasOwnerCleanupRuleForPackageType(ctx, owner.ID, pcr.Type); err != nil {
			ctx.ServerError("HasOwnerCleanupRuleForPackageType", err)
			return
		} else if has {
			ctx.Data["Err_Type"] = true
			ctx.HTML(http.StatusOK, template)
			return
		}

		var err error
		if pcr, err = packages_model.InsertCleanupRule(ctx, pcr); err != nil {
			ctx.ServerError("InsertCleanupRule", err)
			return
		}
	}

	ctx.Flash.Success(ctx.Tr("packages.owner.settings.cleanuprules.success.update"))
	ctx.Redirect(fmt.Sprintf("%s/rules/%d", redirectURL, pcr.ID))
}

func SetRulePreviewContext(ctx *context.Context, owner *user_model.User) {
	pcr := getCleanupRuleByContext(ctx, owner)
	if pcr == nil {
		return
	}

	versionsToRemove, err := cleanup_service.GetCleanupTargets(ctx, pcr, false)
	if err != nil {
		ctx.ServerError("GetCleanupTargets", err)
		return
	}
	packageDescriptors := make([]*packages_model.PackageDescriptor, len(versionsToRemove))
	for i := range len(versionsToRemove) {
		packageDescriptors[i] = versionsToRemove[i].PackageDescriptor
	}

	ctx.Data["CleanupRule"] = pcr
	ctx.Data["VersionsToRemove"] = packageDescriptors
}

func getCleanupRuleByContext(ctx *context.Context, owner *user_model.User) *packages_model.PackageCleanupRule {
	id := ctx.FormInt64("id")
	if id == 0 {
		id = ctx.ParamsInt64("id")
	}

	pcr, err := packages_model.GetCleanupRuleByID(ctx, id)
	if err != nil {
		if err == packages_model.ErrPackageCleanupRuleNotExist {
			ctx.NotFound("", err)
		} else {
			ctx.ServerError("GetCleanupRuleByID", err)
		}
		return nil
	}

	if pcr != nil && pcr.OwnerID == owner.ID {
		return pcr
	}

	ctx.NotFound("", fmt.Errorf("PackageCleanupRule[%v] not associated to owner %v", id, owner))

	return nil
}

func InitializeCargoIndex(ctx *context.Context, owner *user_model.User) {
	err := cargo_service.InitializeIndexRepository(ctx, owner, owner)
	if err != nil {
		log.Error("InitializeIndexRepository failed: %v", err)
		ctx.Flash.Error(ctx.Tr("packages.owner.settings.cargo.initialize.error", err))
	} else {
		ctx.Flash.Success(ctx.Tr("packages.owner.settings.cargo.initialize.success"))
	}
}

func RebuildCargoIndex(ctx *context.Context, owner *user_model.User) {
	err := cargo_service.RebuildIndex(ctx, owner, owner)
	if err != nil {
		log.Error("RebuildIndex failed: %v", err)
		if errors.Is(err, util.ErrNotExist) {
			ctx.Flash.Error(ctx.Tr("packages.owner.settings.cargo.rebuild.no_index"))
		} else {
			ctx.Flash.Error(ctx.Tr("packages.owner.settings.cargo.rebuild.error", err))
		}
	} else {
		ctx.Flash.Success(ctx.Tr("packages.owner.settings.cargo.rebuild.success"))
	}
}
