// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"slices"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/templates"
	"forgejo.org/modules/util"
	"forgejo.org/modules/web"
	"forgejo.org/modules/web/middleware"
	auth_service "forgejo.org/services/auth"
	"forgejo.org/services/authz"
	"forgejo.org/services/context"

	"code.forgejo.org/go-chi/binding"
	"xorm.io/builder"
)

const (
	tplAuthorizedIntegrationList base.TplName = "user/settings/authorized_integrations"
)

var authorizedIntegrationUIs = []authorizedIntegrationUIImpl{
	actionsLocalUI{},
	genericUI{},
}

// Encapsulates the implementation of each authorized integration's UI.
type authorizedIntegrationUIImpl interface {
	// The URL path, and value of the UI field in the authorized integration.
	UIIdentifier() auth_model.AuthorizedIntegrationUI
	// When rendered in the "Add authorized integration" list, the Icon to use.
	Icon(size int) template.HTML
	// When rendered in the "Add authorized integration" list, the Label to use.
	Label(ctx *templates.Context) template.HTML
	// HTML template used when rendering this UI.
	editTemplate() base.TplName
	// When rendering editTemplate, populateTemplateContext will be invoked to allow the UI to perform backend data
	// fetches as needed to populate `ctx.Data`.  The current form will be available as `ctx.Data["Form"]`.
	populateTemplateContext(ctx *context.Context)
	// Form object used when rendering and processing this UI.
	form() authorizedIntegrationUIForm
	// If an error occurs, typically in [convertForm] when evaluating if the inputs provided by the user are sufficient
	// to create claim rules, populateError will be invoked.  If it returns [true] then the [editTemplate] will
	// re-render with the updated context, allowing each UI to handle form errors and display validation results to the
	// user.  If it returns false this indicates the error isn't recognized by the UI, and the error will be inspected
	// by the base form processing and may result in a form error or a server error.
	populateError(ctx *context.Context, err error) (handled bool)
}

// Contains all the form data going to the browser when loading an authorized integration, and returning to the server
// when validating and saving an authorized integration.
//
// Every UI-specific form must embed a [baseAuthorizedIntegrationForm] which contains information common to every
// authorized integration -- identifying information and permission information.  Forms must not conflict on field name
// binding with the fields in [baseAuthorizedIntegrationForm].
type authorizedIntegrationUIForm interface {
	// Check whether the form is empty.  When loading the new & edit pages, this is used to identify if the form data
	// needs to be loaded from the database objects or initialized for a new create.
	isEmpty() bool
	// Access the embedded baseAuthorizedIntegrationForm
	baseForm() *baseAuthorizedIntegrationForm
	// Populate the form from a persisted state, given the stored authorized integration's issuer & claim rules.
	populateForm(ctx *context.Context, issuer string, claimRules *auth_model.ClaimRules) error
	// Convert the form into the issuer and claim rules that will be saved with the authorized integration.
	convertForm(ctx *context.Context) (issuer string, claimRules *auth_model.ClaimRules, err error)
	// Initialize the form with values appropriate for a new authorized integration.
	initNew()
}

// Middleware that resolves the "ui" parameter into ctx.Data.  Access the resolved UI interface via [authorizedIntegrationUI].
func BindAuthorizedIntegrationUI(ctx *context.Context) {
	var ui authorizedIntegrationUIImpl
	uiString := ctx.Params(":ui")
	for _, check := range authorizedIntegrationUIs {
		if auth_model.AuthorizedIntegrationUI(uiString) == check.UIIdentifier() {
			ui = check
			break
		}
	}
	if ui == nil {
		ctx.NotFound("invalid UI", fmt.Errorf("invalid UI: %q is not a supported Authorized Integration UI", uiString))
		return
	}
	ctx.Data["AuthorizedIntegrationUI"] = ui
}

func authorizedIntegrationUI(ctx *context.Context) authorizedIntegrationUIImpl {
	return ctx.Data["AuthorizedIntegrationUI"].(authorizedIntegrationUIImpl)
}

// Middleware that acts like `web.Bind`, but uses the authorized integration UI's to work with a specific form type
// defined by the [AuthorizedIntegrationUI].
func DynamicBindAuthorizedIntegrationForm(ctx *context.Context) {
	ui := authorizedIntegrationUI(ctx)
	formObj := ui.form()
	data := middleware.GetContextData(ctx)
	binding.Bind(ctx.Req, formObj)
	web.SetForm(data, formObj)
}

func ListAuthorizedIntegrations(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.authorized_integrations")
	ctx.Data["PageIsSettingsAuthorizedIntegrations"] = true

	ais, err := db.Find[auth_model.AuthorizedIntegration](ctx,
		auth_model.ListAuthorizedIntegrationOptions{UserID: optional.Some(ctx.Doer.ID)})
	if err != nil {
		ctx.ServerError("ListAuthorizedIntegrations", err)
		return
	}
	ctx.Data["AuthorizedIntegrations"] = ais
	ctx.Data["UIs"] = authorizedIntegrationUIs

	ctx.HTML(http.StatusOK, tplAuthorizedIntegrationList)
}

func getAuthorizedIntegration(ctx *context.Context) *auth_model.AuthorizedIntegration {
	aiUIString := ctx.Params("ui")
	aiUI, err := auth_model.ParseAuthorizedIntegrationUI(aiUIString)
	if err != nil {
		ctx.NotFound("ParseAuthorizedIntegrationUI", err)
		return nil
	}

	aiID := ctx.ParamsInt64("id")
	ai, err := auth_model.GetAuthorizedIntegrationByUI(ctx, ctx.Doer.ID, aiUI, aiID)
	if errors.Is(err, util.ErrNotExist) {
		ctx.NotFound("GetAuthorizedIntegrationByUI", err)
		return nil
	} else if err != nil {
		ctx.ServerError("GetAuthorizedIntegrationByUI", err)
		return nil
	}

	return ai
}

func EditAuthorizedIntegration(ctx *context.Context) {
	ai := getAuthorizedIntegration(ctx)
	if ctx.Written() {
		return
	}

	form := web.GetForm(ctx).(authorizedIntegrationUIForm)
	if form.isEmpty() {
		repos, err := auth_model.GetRepositoriesAccessibleWithIntegration(ctx, ai.ID)
		if err != nil {
			ctx.ServerError("GetRepositoriesAccessibleWithIntegration", err)
			return
		}

		err = form.populateForm(ctx, ai.Issuer, ai.ClaimRules)
		if err != nil {
			ctx.ServerError("copyAuthorizedIntegrationToForm", err)
			return
		}
		err = form.baseForm().copyAuthorizedIntegrationToForm(ctx, ai, repos)
		if err != nil {
			ctx.ServerError("BaseForm().copyAuthorizedIntegrationToForm", err)
			return
		}
	}
	ctx.Data["Form"] = form

	EditAuthorizedIntegrationRenderCommon(ctx)
}

func EditAuthorizedIntegrationPost(ctx *context.Context) {
	form := web.GetForm(ctx).(authorizedIntegrationUIForm)
	ctx.Data["Form"] = form // make form available for template render on any error

	ai := getAuthorizedIntegration(ctx)
	if ctx.Written() {
		return
	}

	issuer, claimRules, err := form.convertForm(ctx)
	if err != nil {
		editAuthorizedIntegrationErrorHandler(ctx, err)
		return
	}
	ai.Issuer = issuer
	ai.ClaimRules = claimRules

	rr, err := form.baseForm().copyFormToAuthorizedIntegration(ctx, ai)
	if err != nil {
		editAuthorizedIntegrationErrorHandler(ctx, err)
		return
	}

	if err := auth_service.UpdateAuthorizedIntegration(ctx, ai, rr); err != nil {
		editAuthorizedIntegrationErrorHandler(ctx, err)
		return
	}

	ctx.Redirect(setting.AppSubURL + "/user/settings/authorized-integrations")
}

func NewAuthorizedIntegration(ctx *context.Context) {
	form := web.GetForm(ctx).(authorizedIntegrationUIForm)
	if form.isEmpty() {
		form.initNew()
		form.baseForm().InitNew()
	}
	ctx.Data["Form"] = form
	ctx.Data["IsNew"] = true

	EditAuthorizedIntegrationRenderCommon(ctx)
}

func NewAuthorizedIntegrationPost(ctx *context.Context) {
	form := web.GetForm(ctx).(authorizedIntegrationUIForm)
	ctx.Data["Form"] = form // make form available for template render on any error
	ctx.Data["IsNew"] = true

	ai := &auth_model.AuthorizedIntegration{
		UserID: ctx.Doer.ID,
		UI:     authorizedIntegrationUI(ctx).UIIdentifier(),
	}

	issuer, claimRules, err := form.convertForm(ctx)
	if err != nil {
		editAuthorizedIntegrationErrorHandler(ctx, err)
		return
	}
	ai.Issuer = issuer
	ai.ClaimRules = claimRules

	rr, err := form.baseForm().copyFormToAuthorizedIntegration(ctx, ai)
	if err != nil {
		editAuthorizedIntegrationErrorHandler(ctx, err)
		return
	}

	if err := auth_service.InsertAuthorizedIntegration(ctx, ai, rr); err != nil {
		editAuthorizedIntegrationErrorHandler(ctx, err)
		return
	}

	ctx.Flash.Success(ctx.Tr("settings.authorized_integration.create_success", ai.Name))
	ctx.Redirect(setting.AppSubURL + fmt.Sprintf("/user/settings/authorized-integrations/%s/%d", authorizedIntegrationUI(ctx).UIIdentifier(), ai.ID))
}

func editAuthorizedIntegrationErrorHandler(ctx *context.Context, err error) {
	if authorizedIntegrationUI(ctx).populateError(ctx, err) {
		EditAuthorizedIntegrationRenderCommon(ctx)
		return
	}

	// Note: auth_service.ErrInvalidClaimRules is not handled here.  If a UI created claim rules that the auth service
	// identified as invalid, it indicates a logic bug or unhandled case in the UI implementation, and will be treated
	// as a server error because the base implementation here doesn't know what UI fields would be responsible for these
	// invalid claim rules.  (Excepting the Generic UI, which handles this case itself.)
	var errMissingField *auth_service.MissingFieldError
	switch {
	case errors.As(err, &errMissingField):
		switch errMissingField.Field {
		case "Name":
			ctx.Data["Err_Name"] = true
			ctx.Flash.Error(ctx.Tr("settings.authorized_integration.name.required"), true)
		default:
			// Unrecognized field; fallback to server error handling.
			ctx.ServerError("UpdateAuthorizedIntegration", err)
			return
		}
		EditAuthorizedIntegrationRenderCommon(ctx)
		return
	case errors.Is(err, auth_service.ErrInvalidIssuer):
		// ErrInvalidIssuer is a little awkward if it reaches here.  Most UIs should prevent the user from entering
		// something that would cause auth service to find the OIDC issuer to be invalid, but if we've reached here that
		// hasn't happened.  Validating the issuer performs remote HTTP calls so it's possible that this is a transient
		// error, or the state of the remote has changed (used to be valid, isn't currently).  We'll flash the error
		// here as it might be useful to the user but we don't know how to highlight the relevant fields for the error.
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.issuer.invalid", err.Error()), true)
		EditAuthorizedIntegrationRenderCommon(ctx)
		return
	case errors.Is(err, auth_service.ErrInvalidClaimRules):
		ctx.Data["Err_ClaimRules"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.claim_rules.invalid", err.Error()), true)
		EditAuthorizedIntegrationRenderCommon(ctx)
		return
	case errors.Is(err, authz.ErrSpecifiedReposNone):
		ctx.Data["Err_SelectedRepo"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.specified_repos_none"), true)
		EditAuthorizedIntegrationRenderCommon(ctx)
		return
	case errors.Is(err, authz.ErrSpecifiedReposNoPublicOnly):
		ctx.Data["Err_SelectedRepo"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.specified_repos_and_public_only"), true)
		EditAuthorizedIntegrationRenderCommon(ctx)
		return
	case errors.Is(err, authz.ErrSpecifiedReposInvalidScope):
		ctx.Data["Err_SelectedRepo"] = true
		ctx.Data["Err_Scope"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.specified_repos_and_invalid_scope"), true)
		EditAuthorizedIntegrationRenderCommon(ctx)
		return
	}

	ctx.ServerError("UpdateAuthorizedIntegration", err)
}

func EditAuthorizedIntegrationRenderCommon(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("settings.authorized_integrations")
	ctx.Data["PageIsSettingsAuthorizedIntegrations"] = true

	categories := []string{
		"activitypub",
		"issue",
		"misc",
		"notification",
		"organization",
		"package",
		"repository",
		"user",
	}
	if ctx.Doer.IsAdmin {
		categories = append(categories, "admin")
	}
	slices.Sort(categories)
	ctx.Data["Categories"] = categories

	repoMultiSelect(ctx)
	if ctx.Written() {
		return
	}
	authorizedIntegrationUI(ctx).populateTemplateContext(ctx)
	if ctx.Written() {
		return
	}

	ctx.HTML(http.StatusOK, authorizedIntegrationUI(ctx).editTemplate())
}

func repoMultiSelect(ctx *context.Context) {
	form := ctx.Data["Form"].(authorizedIntegrationUIForm).baseForm()

	if form.AddSelectedRepo != "" {
		form.SelectedRepo = append(form.SelectedRepo, form.AddSelectedRepo)
	}
	if form.RemoveSelectedRepo != "" {
		form.SelectedRepo = slices.DeleteFunc(
			form.SelectedRepo,
			func(r string) bool { return r == form.RemoveSelectedRepo },
		)
	}

	selectedRepos, err := getSelectedRepos(ctx, form.SelectedRepo)
	if err != nil {
		ctx.Error(http.StatusBadRequest, "getSelectedRepos")
		return
	}
	ctx.Data["SelectedRepos"] = selectedRepos

	repoSearchText := form.RepoSearch

	page := 1
	// Pagination on the repo search has form submit buttons that send the `set_page` param.  It's then encoded into the
	// page in the hidden input `page` which we fall back to, if anything else causes a form get (eg. adding or removing
	// a repo).
	if form.SetPage > 0 {
		page = form.SetPage
	} else if form.Page > 0 {
		page = form.Page
	}
	pageSize := 10
	repoSearch := &repo_model.SearchRepoOptions{
		Actor:    ctx.Doer,
		Keyword:  repoSearchText,
		Private:  true,
		Archived: optional.Some(false),

		// Restrict repositories to those owned by, or collaborated with, by the user.  Repo-specific access tokens
		// could theoretically be created on any public repository as well, but there wouldn't be much point to that and
		// it would really balloon the search results to an impractical number of repos.
		OwnerID: ctx.Doer.ID,

		ListOptions: db.ListOptions{
			Page:     page,
			PageSize: pageSize,
		},
		OrderBy: db.SearchOrderByAlphabetically, // match sorting in getSelectedRepos for consistency
	}
	cond := repo_model.SearchRepositoryCondition(repoSearch)
	// Exclude all the repos that are currently in `form.SelectedRepo` from the search, by omitting them from the search
	// condition.  This prevents the UI from displaying the same repo on the left and right, and maintains the repo
	// search and page-size correctly.
	for _, selected := range selectedRepos {
		cond = cond.And(builder.Neq{"id": selected.ID})
	}
	repos, count, err := repo_model.SearchRepositoryByCondition(ctx, repoSearch, cond, false)
	if err != nil {
		log.Error("SearchRepository: %v", err)
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.Data["Repos"] = repos

	pager := context.NewPagination(int(count), pageSize, page, 3)
	pager.SetDefaultParams(ctx)
	ctx.Data["Page"] = pager
}

func DeleteAuthorizedIntegration(ctx *context.Context) {
	if err := auth_model.DeleteAuthorizedIntegrationByID(ctx, ctx.FormInt64("id"), ctx.Doer.ID); err != nil {
		ctx.Flash.Error("DeleteAuthorizedIntegrationByID: " + err.Error())
	} else {
		ctx.Flash.Success(ctx.Tr("settings.authorized_integration.deleted"))
	}
	ctx.JSONRedirect(setting.AppSubURL + "/user/settings/authorized-integrations")
}
