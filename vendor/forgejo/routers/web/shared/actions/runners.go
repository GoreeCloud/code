// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/models/db"
	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/util"
	"forgejo.org/modules/web"
	shared_user "forgejo.org/routers/web/shared/user"
	"forgejo.org/services/context"
	"forgejo.org/services/forms"

	gouuid "github.com/google/uuid"
)

const (
	tplAdminRunnerCreate  base.TplName = "admin/runners/create"
	tplAdminRunnerDetails base.TplName = "admin/runners/details"
	tplAdminRunnerEdit    base.TplName = "admin/runners/edit"
	tplAdminRunnerSetup   base.TplName = "admin/runners/setup"
	tplAdminRunners       base.TplName = "admin/actions"
	tplOrgRunnerCreate    base.TplName = "org/settings/runners_create"
	tplOrgRunnerDetails   base.TplName = "org/settings/runners_details"
	tplOrgRunnerEdit      base.TplName = "org/settings/runners_edit"
	tplOrgRunnerSetup     base.TplName = "org/settings/runners_setup"
	tplOrgRunners         base.TplName = "org/settings/actions"
	tplRepoRunnerCreate   base.TplName = "repo/settings/runner_create"
	tplRepoRunnerDetails  base.TplName = "repo/settings/runner_details"
	tplRepoRunnerEdit     base.TplName = "repo/settings/runner_edit"
	tplRepoRunnerSetup    base.TplName = "repo/settings/runner_setup"
	tplRepoRunners        base.TplName = "repo/settings/actions"
	tplUserRunnerCreate   base.TplName = "user/settings/runner_create"
	tplUserRunnerDetails  base.TplName = "user/settings/runner_details"
	tplUserRunnerEdit     base.TplName = "user/settings/runner_edit"
	tplUserRunnerSetup    base.TplName = "user/settings/runner_setup"
	tplUserRunners        base.TplName = "user/settings/actions"
)

type runnersCtx struct {
	OwnerID               int64
	RepoID                int64
	IsRepo                bool
	IsOrg                 bool
	IsAdmin               bool
	IsUser                bool
	RunnerCreateTemplate  base.TplName
	RunnerDetailsTemplate base.TplName
	RunnerEditTemplate    base.TplName
	RunnerSetupTemplate   base.TplName
	RunnersTemplate       base.TplName
	RedirectLink          string
}

func getRunnersCtx(ctx *context.Context) (*runnersCtx, error) {
	if ctx.Data["PageIsRepoSettings"] == true {
		return &runnersCtx{
			RepoID:                ctx.Repo.Repository.ID,
			OwnerID:               0,
			IsRepo:                true,
			RunnerCreateTemplate:  tplRepoRunnerCreate,
			RunnerDetailsTemplate: tplRepoRunnerDetails,
			RunnerEditTemplate:    tplRepoRunnerEdit,
			RunnerSetupTemplate:   tplRepoRunnerSetup,
			RunnersTemplate:       tplRepoRunners,
			RedirectLink:          ctx.Repo.RepoLink + "/settings/actions/runners/",
		}, nil
	}

	if ctx.Data["PageIsOrgSettings"] == true {
		err := shared_user.LoadHeaderCount(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not load project and package counts: %w", err)
		}
		return &runnersCtx{
			RepoID:                0,
			OwnerID:               ctx.Org.Organization.ID,
			IsOrg:                 true,
			RunnerCreateTemplate:  tplOrgRunnerCreate,
			RunnerDetailsTemplate: tplOrgRunnerDetails,
			RunnerEditTemplate:    tplOrgRunnerEdit,
			RunnerSetupTemplate:   tplOrgRunnerSetup,
			RunnersTemplate:       tplOrgRunners,
			RedirectLink:          ctx.Org.OrgLink + "/settings/actions/runners/",
		}, nil
	}

	if ctx.Data["PageIsAdmin"] == true {
		return &runnersCtx{
			RepoID:                0,
			OwnerID:               0,
			IsAdmin:               true,
			RunnerCreateTemplate:  tplAdminRunnerCreate,
			RunnerDetailsTemplate: tplAdminRunnerDetails,
			RunnerEditTemplate:    tplAdminRunnerEdit,
			RunnerSetupTemplate:   tplAdminRunnerSetup,
			RunnersTemplate:       tplAdminRunners,
			RedirectLink:          setting.AppSubURL + "/admin/actions/runners/",
		}, nil
	}

	if ctx.Data["PageIsUserSettings"] == true {
		return &runnersCtx{
			OwnerID:               ctx.Doer.ID,
			RepoID:                0,
			IsUser:                true,
			RunnerCreateTemplate:  tplUserRunnerCreate,
			RunnerDetailsTemplate: tplUserRunnerDetails,
			RunnerEditTemplate:    tplUserRunnerEdit,
			RunnerSetupTemplate:   tplUserRunnerSetup,
			RunnersTemplate:       tplUserRunners,
			RedirectLink:          setting.AppSubURL + "/user/settings/actions/runners/",
		}, nil
	}

	return nil, errors.New("unable to set Runners context")
}

// RunnersList renders the list of runners.
func RunnersList(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	page := max(ctx.FormInt("page"), 1)

	opts := actions_model.FindRunnerOptions{
		ListOptions: db.ListOptions{
			Page:     page,
			PageSize: 100,
		},
		WithVisible: true,
		Sort:        ctx.Req.URL.Query().Get("sort"),
		Filter:      ctx.Req.URL.Query().Get("q"),
	}
	if rCtx.IsRepo {
		opts.RepoID = rCtx.RepoID
	} else if rCtx.IsOrg || rCtx.IsUser {
		opts.OwnerID = rCtx.OwnerID
	}

	runners, count, err := db.FindAndCount[actions_model.ActionRunner](ctx, opts)
	if err != nil {
		ctx.ServerError("CountRunners", err)
		return
	}

	if err := actions_model.RunnerList(runners).LoadAttributes(ctx); err != nil {
		ctx.ServerError("LoadAttributes", err)
		return
	}

	// ownid=0,repo_id=0,means this token is used for global
	ownerID := optional.None[int64]()
	if opts.OwnerID != 0 {
		ownerID = optional.Some(opts.OwnerID)
	}
	repoID := optional.None[int64]()
	if opts.RepoID != 0 {
		repoID = optional.Some(opts.RepoID)
	}

	var token *actions_model.ActionRunnerToken
	token, err = actions_model.GetLatestRunnerToken(ctx, ownerID, repoID)
	if errors.Is(err, util.ErrNotExist) || (token != nil && !token.IsActive) {
		token, err = actions_model.NewRunnerToken(ctx, ownerID, repoID)
		if err != nil {
			ctx.ServerError("CreateRunnerToken", err)
			return
		}
	} else if err != nil {
		ctx.ServerError("GetLatestRunnerToken", err)
		return
	}

	ctx.Data["PageIsSharedSettingsRunners"] = true
	ctx.Data["Title"] = ctx.Tr("actions.actions")
	ctx.Data["PageType"] = "runners"
	ctx.Data["Keyword"] = opts.Filter
	ctx.Data["Runners"] = runners
	ctx.Data["Total"] = count
	ctx.Data["RegistrationToken"] = token.Token
	ctx.Data["RunnerOwnerID"] = opts.OwnerID
	ctx.Data["RunnerRepoID"] = opts.RepoID
	ctx.Data["SortType"] = opts.Sort
	ctx.Data["RunnersListLink"] = rCtx.RedirectLink

	pager := context.NewPagination(int(count), opts.PageSize, opts.Page, 5)

	ctx.Data["Page"] = pager
	ctx.HTML(http.StatusOK, rCtx.RunnersTemplate)
}

// RunnerDetails displays detail information about each runner. The page is purely informational and visible to everyone
// who is allowed to use a runner.
func RunnerDetails(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	runnerID := ctx.ParamsInt64(":runnerid")
	page := max(ctx.FormInt("page"), 1)

	runner, err := actions_model.GetVisibleRunnerByID(ctx, runnerID, rCtx.OwnerID, rCtx.RepoID)
	if errors.Is(err, util.ErrNotExist) {
		ctx.NotFound("GetVisibleRunnerByID", err)
		return
	} else if err != nil {
		ctx.ServerError("GetVisibleRunnerByID", err)
		return
	}
	if err := runner.LoadAttributes(ctx); err != nil {
		ctx.ServerError("LoadAttributes", err)
		return
	}

	opts := actions_model.FindTaskOptions{
		ListOptions: db.ListOptions{
			Page:     page,
			PageSize: 30,
		},
		RunnerID: runner.ID,
		OwnerID:  rCtx.OwnerID,
		RepoID:   rCtx.RepoID,
	}

	tasks, count, err := db.FindAndCount[actions_model.ActionTask](ctx, opts)
	if err != nil {
		ctx.ServerError("CountTasks", err)
		return
	}

	if err = actions_model.TaskList(tasks).LoadAttributes(ctx); err != nil {
		ctx.ServerError("TasksLoadAttributes", err)
		return
	}

	ctx.Data["PageIsSharedSettingsRunners"] = true
	ctx.Data["RunnerOwnerID"] = rCtx.OwnerID
	ctx.Data["RunnerRepoID"] = rCtx.RepoID
	ctx.Data["Title"] = ctx.Tr("actions.runners.runner_details.page_title", runner.Name)
	ctx.Data["Runner"] = runner
	ctx.Data["Tasks"] = tasks
	ctx.Data["IsRepo"] = rCtx.IsRepo
	ctx.Data["IsOrg"] = rCtx.IsOrg
	ctx.Data["IsAdmin"] = rCtx.IsAdmin
	ctx.Data["IsUser"] = rCtx.IsUser
	pager := context.NewPagination(int(count), opts.PageSize, opts.Page, 5)
	ctx.Data["Page"] = pager
	ctx.Data["RunnersListLink"] = rCtx.RedirectLink

	ctx.HTML(http.StatusOK, rCtx.RunnerDetailsTemplate)
}

// RunnerCreate displays a form for creating a new runner.
func RunnerCreate(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	ctx.Data["PageIsSharedSettingsRunners"] = true
	ctx.Data["Title"] = ctx.Tr("actions.runners.create_runner.page_title")
	ctx.Data["RunnersListLink"] = rCtx.RedirectLink
	ctx.HTML(http.StatusOK, rCtx.RunnerCreateTemplate)
}

// RunnerCreatePost handles the form submitted by RunnerCreate.
func RunnerCreatePost(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	form := web.GetForm(ctx).(*forms.CreateRunnerForm)

	runner := actions_model.ActionRunner{
		UUID:        gouuid.New().String(),
		Name:        form.RunnerName,
		OwnerID:     rCtx.OwnerID,
		RepoID:      rCtx.RepoID,
		Description: form.RunnerDescription,
		Ephemeral:   false,
	}
	runner.GenerateToken()

	ctx.Data["PageIsSharedSettingsRunners"] = true
	ctx.Data["Title"] = ctx.Tr("actions.runners.runner_setup.page_title", runner.Name)
	ctx.Data["AppURL"] = setting.AppURL
	ctx.Data["Runner"] = runner
	ctx.Data["RunnerOwnerID"] = rCtx.OwnerID
	ctx.Data["RunnerRepoID"] = rCtx.RepoID
	ctx.Data["RunnersListLink"] = rCtx.RedirectLink

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, rCtx.RunnerCreateTemplate)
		return
	}

	err = actions_model.CreateRunner(ctx, &runner)
	if err != nil {
		ctx.ServerError("CreateRunner", err)
		return
	}

	ctx.HTML(http.StatusOK, rCtx.RunnerSetupTemplate)
}

// RunnerEdit displays a form to modify the given runner.
func RunnerEdit(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	runner, err := actions_model.GetVisibleRunnerByID(ctx, ctx.ParamsInt64(":runnerid"), rCtx.OwnerID, rCtx.RepoID)
	if errors.Is(err, util.ErrNotExist) {
		ctx.NotFound("GetVisibleRunnerByID", err)
		return
	} else if err != nil {
		ctx.ServerError("GetVisibleRunnerByID", err)
		return
	}
	if err := runner.LoadAttributes(ctx); err != nil {
		ctx.ServerError("LoadAttributes", err)
		return
	}
	if !runner.Editable(rCtx.OwnerID, rCtx.RepoID) {
		err = errors.New("no permission to edit this runner")
		ctx.NotFound("RunnerDetails", err)
		return
	}

	ctx.Data["PageIsSharedSettingsRunners"] = true
	ctx.Data["Title"] = ctx.Tr("actions.runners.edit_runner.page_title", runner.Name)
	ctx.Data["Runner"] = runner
	ctx.Data["RunnerOwnerID"] = rCtx.OwnerID
	ctx.Data["RunnerRepoID"] = rCtx.RepoID
	ctx.Data["RunnersListLink"] = rCtx.RedirectLink
	ctx.HTML(http.StatusOK, rCtx.RunnerEditTemplate)
}

// RunnerEditPost handles the form submitted by RunnerEdit.
func RunnerEditPost(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	ctx.Data["RunnersListLink"] = rCtx.RedirectLink

	runnerID := ctx.ParamsInt64(":runnerid")
	redirectURL := rCtx.RedirectLink + url.PathEscape(ctx.Params(":runnerid"))

	runner, err := actions_model.GetVisibleRunnerByID(ctx, runnerID, rCtx.OwnerID, rCtx.RepoID)
	if errors.Is(err, util.ErrNotExist) {
		ctx.NotFound("GetVisibleRunnerByID", err)
		return
	} else if err != nil {
		ctx.ServerError("GetVisibleRunnerByID", err)
		return
	}
	if !runner.Editable(rCtx.OwnerID, rCtx.RepoID) {
		ctx.NotFound("RunnerEditPost.Editable", util.NewPermissionDeniedErrorf("no permission to edit this runner"))
		return
	}

	ctx.Data["PageIsSharedSettingsRunners"] = true
	ctx.Data["Title"] = ctx.Tr("actions.runners.runner_setup.page_title", runner.Name)
	ctx.Data["AppURL"] = setting.AppURL
	ctx.Data["Runner"] = runner
	ctx.Data["RunnerOwnerID"] = rCtx.OwnerID
	ctx.Data["RunnerRepoID"] = rCtx.RepoID
	ctx.Data["RunnersListLink"] = rCtx.RedirectLink

	form := web.GetForm(ctx).(*forms.EditRunnerForm)
	runner.Name = form.RunnerName
	runner.Description = form.RunnerDescription

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, rCtx.RunnerEditTemplate)
		return
	}

	if !form.RegenerateToken {
		err = actions_model.UpdateRunner(ctx, runner, "name", "description")
		if err != nil {
			log.Warn("RunnerEditPost.UpdateRunner failed: %v, url: %s", err, ctx.Req.URL)
			ctx.Flash.Warning(ctx.Tr("actions.runners.update_runner.failed"))
			ctx.Redirect(redirectURL)
			return
		}

		log.Debug("RunnerEditPost success: %s", ctx.Req.URL)

		ctx.Flash.Success(ctx.Tr("actions.runners.update_runner.success"))
		ctx.Redirect(redirectURL)
		return
	}

	runner.GenerateToken()
	err = actions_model.UpdateRunner(ctx, runner, "name", "description", "token_hash", "token_salt")
	if err != nil {
		log.Warn("RunnerEditPost.UpdateRunner failed: %v, url: %s", err, ctx.Req.URL)
		ctx.Flash.Warning(ctx.Tr("actions.runners.update_runner.failed"))
		ctx.Redirect(redirectURL)
		return
	}

	ctx.HTML(http.StatusOK, rCtx.RunnerSetupTemplate)
}

// RunnerResetRegistrationToken resets the runner registration token.
func RunnerResetRegistrationToken(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	optOwnerID := optional.None[int64]()
	if rCtx.OwnerID != 0 {
		optOwnerID = optional.Some(rCtx.OwnerID)
	}
	optRepoID := optional.None[int64]()
	if rCtx.RepoID != 0 {
		optRepoID = optional.Some(rCtx.RepoID)
	}

	_, err = actions_model.NewRunnerToken(ctx, optOwnerID, optRepoID)
	if err != nil {
		ctx.ServerError("ResetRunnerRegistrationToken", err)
		return
	}

	ctx.Flash.Success(ctx.Tr("actions.runners.reset_registration_token.success"))
	ctx.Redirect(rCtx.RedirectLink)
}

// RunnerDeletePost handles the request for deleting a particular runner.
func RunnerDeletePost(ctx *context.Context) {
	rCtx, err := getRunnersCtx(ctx)
	if err != nil {
		ctx.ServerError("getRunnersCtx", err)
		return
	}

	runner, err := actions_model.GetRunnerByID(ctx, ctx.ParamsInt64(":runnerid"))
	if err != nil {
		ctx.ServerError("GetRunnerByID", err)
		return
	}

	if !runner.Editable(rCtx.OwnerID, rCtx.RepoID) {
		ctx.NotFound("Editable", util.NewPermissionDeniedErrorf("no permission to edit this runner"))
		return
	}

	if err := actions_model.DeleteRunner(ctx, runner); err != nil {
		log.Warn("DeleteRunnerPost.UpdateRunner failed: %v, url: %s", err, ctx.Req.URL)

		ctx.Flash.Warning(ctx.Tr("actions.runners.delete_runner.failed"))

		ctx.JSONRedirect(rCtx.RedirectLink)
		return
	}

	log.Info("DeleteRunnerPost success: %s", ctx.Req.URL)

	ctx.Flash.Success(ctx.Tr("actions.runners.delete_runner.success"))

	ctx.JSONRedirect(rCtx.RedirectLink)
}

func RedirectToDefaultSetting(ctx *context.Context) {
	ctx.Redirect(ctx.Repo.RepoLink + "/settings/actions/runners")
}
