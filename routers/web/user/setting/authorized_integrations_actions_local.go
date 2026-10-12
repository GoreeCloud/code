// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	access_model "forgejo.org/models/perm/access"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unit"
	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/modules/svg"
	"forgejo.org/modules/templates"
	auth_service "forgejo.org/services/auth"
	"forgejo.org/services/context"

	"github.com/gobwas/glob"
)

var (
	_ authorizedIntegrationUIImpl = actionsLocalUI{}
	_ authorizedIntegrationUIForm = &actionsLocalAuthorizedIntegrationForm{}

	errInvalidWorkflowFileGlob = errors.New("invalid workflow file glob")
	errInvalidGitRefGlob       = errors.New("invalid git ref glob")
)

type actionsLocalUI struct{}

func (actionsLocalUI) UIIdentifier() auth_model.AuthorizedIntegrationUI {
	return auth_model.AuthorizedIntegrationUIForgejoActionsLocal
}

func (actionsLocalUI) Icon(size int) template.HTML {
	return svg.RenderHTML("gitea-forgejo", size, "img")
}

func (actionsLocalUI) Label(ctx *templates.Context) template.HTML {
	return ctx.Locale.Tr("settings.authorized_integration.ui.forgejo_actions_local")
}

func (actionsLocalUI) editTemplate() base.TplName {
	return "user/settings/authorized_integrations/actions_local/view"
}

func (actionsLocalUI) populateTemplateContext(ctx *context.Context) {
	// Varient of the base template's repoSingleSelect, except it supports a single-select, and filters repositories to
	// only those that enable Action unit.
	form := ctx.Data["Form"].(*actionsLocalAuthorizedIntegrationForm)

	if form.SourceRepo != "" {
		selectedRepos, err := getSelectedRepos(ctx, []string{form.SourceRepo})
		if err != nil {
			ctx.Error(http.StatusBadRequest, "getSelectedRepos")
			return
		} else if len(selectedRepos) != 1 {
			ctx.Error(http.StatusInternalServerError, "unexpected selectedRepos len")
			return
		}
		ctx.Data["SourceRepo"] = selectedRepos[0]
	} else {
		repoSearchText := form.ActionRepoSearch

		page := 1
		// Pagination on the repo search has form submit buttons that send the `set_page` param.  It's then encoded into the
		// page in the hidden input `page` which we fall back to, if anything else causes a form get (eg. adding or removing
		// a repo).
		if form.ActionSetPage > 0 {
			page = form.ActionSetPage
		} else if form.ActionPage > 0 {
			page = form.ActionPage
		}
		pageSize := 10
		repoSearch := &repo_model.SearchRepoOptions{
			Actor:       ctx.Doer,
			Keyword:     repoSearchText,
			Private:     true,
			Archived:    optional.Some(false),
			EnabledUnit: optional.Some(unit.TypeActions),

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
		repos, count, err := repo_model.SearchRepositoryByCondition(ctx, repoSearch, cond, false)
		if err != nil {
			log.Error("SearchRepository: %v", err)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		ctx.Data["ActionsRepos"] = repos

		pager := context.NewPagination(int(count), pageSize, page, 3)
		pager.SetDefaultParams(ctx)
		ctx.Data["ActionsPage"] = pager
	}
}

func (actionsLocalUI) form() authorizedIntegrationUIForm {
	return &actionsLocalAuthorizedIntegrationForm{}
}

func (actionsLocalUI) populateError(ctx *context.Context, err error) (handled bool) {
	var errMissingField *auth_service.MissingFieldError
	switch {
	case errors.As(err, &errMissingField):
		switch errMissingField.Field {
		case "SourceRepo":
			ctx.Data["Err_SourceRepo"] = true
			ctx.Flash.Error(ctx.Tr("settings.authorized_integration.forgejo_actions_local.repo.required"), true)
			return true
		default:
			// Unrecognized field; fallback to server error handling.
			return false
		}
	case errors.Is(err, errInvalidWorkflowFileGlob):
		ctx.Data["Err_WorkflowFile"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.forgejo_actions_local.workflow_file.error", err), true)
		return true
	case errors.Is(err, errInvalidGitRefGlob):
		ctx.Data["Err_GitRef"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.forgejo_actions_local.git_ref.error", err), true)
		return true
	}
	return false
}

type actionsLocalAuthorizedIntegrationForm struct {
	baseAuthorizedIntegrationForm
	SourceRepo   string // formatted as ownername/reponame
	WorkflowFile string
	GitRef       string
	Event        []string

	// Values used for source repo search & selection, not stored in Authorized Integration.  Must avoid conflicting
	// with similar form values in baseAuthorizedIntegrationForm.
	ActionRepoSearch string
	ActionPage       int // repo search page
	ActionSetPage    int // repo search buttons
}

func (g *actionsLocalAuthorizedIntegrationForm) baseForm() *baseAuthorizedIntegrationForm {
	return &g.baseAuthorizedIntegrationForm
}

func (g *actionsLocalAuthorizedIntegrationForm) isEmpty() bool {
	return g.baseAuthorizedIntegrationForm.isEmpty() && g.SourceRepo == "" && g.WorkflowFile == "" && g.GitRef == ""
}

func (g *actionsLocalAuthorizedIntegrationForm) populateForm(ctx *context.Context, issuer string, claimRules *auth_model.ClaimRules) error {
	var err error
	var repositoryID, repositoryOwnerID int64

	for _, cr := range claimRules.Rules {
		switch {
		case cr.Claim == "repository_id" && cr.Comparison == auth_model.ClaimEqual:
			repositoryID, err = strconv.ParseInt(cr.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("unexpected claim rule value on claim %q: %#v", cr.Claim, err)
			}
		case cr.Claim == "repository_owner_id" && cr.Comparison == auth_model.ClaimEqual:
			repositoryOwnerID, err = strconv.ParseInt(cr.Value, 10, 64)
			if err != nil {
				return fmt.Errorf("unexpected claim rule value on claim %q: %#v", cr.Claim, err)
			}
		case cr.Claim == "workflow" && cr.Comparison == auth_model.ClaimGlob:
			g.WorkflowFile = cr.Value
		case cr.Claim == "ref" && cr.Comparison == auth_model.ClaimGlob:
			g.GitRef = cr.Value
		case cr.Claim == "event_name" && cr.Comparison == auth_model.ClaimIn:
			g.Event = cr.Values
		default:
			return fmt.Errorf("unexpected claim rule: %#v", cr)
		}
	}

	if repositoryID == 0 && repositoryOwnerID == 0 {
		g.SourceRepo = ""
	} else {
		repo, err := repo_model.GetRepositoryByID(ctx, repositoryID)
		if err != nil {
			return fmt.Errorf("unable to load repo: %w", err)
		}
		permission, err := access_model.GetUserRepoPermission(ctx, repo, ctx.Doer)
		if err != nil {
			return err
		} else if !permission.HasAccess() || repo.OwnerID != repositoryOwnerID {
			return fmt.Errorf("repository could not be loaded")
		}
		g.SourceRepo = fmt.Sprintf("%s/%s", repo.OwnerName, repo.Name)
	}

	return nil
}

func (g *actionsLocalAuthorizedIntegrationForm) convertForm(ctx *context.Context) (issuer string, claimRules *auth_model.ClaimRules, err error) {
	issuer = "urn:forgejo:authorized-integrations:actions"

	rules := []auth_model.ClaimRule{}

	if g.SourceRepo == "" {
		return "", nil, &auth_service.MissingFieldError{Field: "SourceRepo"}
	}
	selectedRepos, err := getSelectedRepos(ctx, []string{g.SourceRepo})
	if err != nil {
		return "", nil, err
	} else if len(selectedRepos) != 1 {
		return "", nil, fmt.Errorf("expected only one repo, but received %d", len(selectedRepos))
	}
	repo := selectedRepos[0]
	// We'll translate the selected repo into two claim rules -- matching the repo's owner, and the repository,
	// based upon their immutable IDs (not their changeable names).  This is a little bit inflexible as it means
	// that the authorized integration will stop accepting JWTs from the repository if the owner is changed (for
	// example, the repository is transferred), which we could permit if we just matched based upon the repository
	// ID.  Matching on both fields is a bit more security conservative though; a repository transfer could be part
	// of stealing control of a repo from one admin to another, and it seems a little safer to reduce the repo's
	// access in that case and require the authorized integration owner to take some action to reestablish trust.
	rules = append(rules,
		auth_model.ClaimRule{
			Claim:      "repository_id",
			Comparison: auth_model.ClaimEqual,
			Value:      fmt.Sprintf("%d", repo.ID),
		},
		auth_model.ClaimRule{
			Claim:      "repository_owner_id",
			Comparison: auth_model.ClaimEqual,
			Value:      fmt.Sprintf("%d", repo.OwnerID),
		})

	if g.WorkflowFile != "" {
		_, err := glob.Compile(g.WorkflowFile)
		if err != nil {
			return "", nil, fmt.Errorf("%w: %w", errInvalidWorkflowFileGlob, err)
		}
		rules = append(rules, auth_model.ClaimRule{
			Claim:      "workflow",
			Comparison: auth_model.ClaimGlob,
			Value:      g.WorkflowFile,
		})
	}
	if g.GitRef != "" {
		_, err := glob.Compile(g.GitRef)
		if err != nil {
			return "", nil, fmt.Errorf("%w: %w", errInvalidGitRefGlob, err)
		}
		rules = append(rules, auth_model.ClaimRule{
			Claim:      "ref",
			Comparison: auth_model.ClaimGlob,
			Value:      g.GitRef,
		})
	}
	if len(g.Event) != 0 {
		rules = append(rules, auth_model.ClaimRule{
			Claim:      "event_name",
			Comparison: auth_model.ClaimIn,
			Values:     g.Event,
		})
	}

	// Safety check -- authorized integrations do support having empty claim rules, but it should never be the case for
	// the Actions Local UI to create this situation:
	if len(rules) == 0 {
		return "", nil, fmt.Errorf("unexpected: Actions Local UI didn't define any claim rules")
	}

	claimRules = &auth_model.ClaimRules{Rules: rules}

	return issuer, claimRules, nil
}

func (g *actionsLocalAuthorizedIntegrationForm) initNew() {
}
