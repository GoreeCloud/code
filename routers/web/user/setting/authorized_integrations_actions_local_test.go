// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import (
	"errors"
	"testing"

	auth_model "forgejo.org/models/auth"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/templates"
	auth_service "forgejo.org/services/auth"
	"forgejo.org/services/context"
	"forgejo.org/services/contexttest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeLocalContext(t *testing.T) *context.Context {
	ctx, _ := contexttest.MockContext(t, "user/settings/authorized-integrations/forgejo-actions-local/new",
		contexttest.MockContextOption{Render: templates.HTMLRenderer()})
	ctx.Doer = unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	return ctx
}

func TestLocalPopulateTemplateContext(t *testing.T) {
	ui := actionsLocalUI{}

	t.Run("search for repos", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{}

		ctx := makeLocalContext(t)
		ctx.Data["Form"] = form
		ui.populateTemplateContext(ctx)
		require.False(t, ctx.Written()) // no error written to ctx

		actionsRepos := ctx.Data["ActionsRepos"].(repo_model.RepositoryList)
		require.Len(t, actionsRepos, 3)
		assert.Equal(t, "repo1", actionsRepos[0].Name)
		assert.Equal(t, "test_action_run_search", actionsRepos[1].Name)
		assert.Equal(t, "test_workflows", actionsRepos[2].Name)

		pager := ctx.Data["ActionsPage"].(*context.Pagination)
		assert.Equal(t, 3, pager.Paginater.Total())
	})

	t.Run("has repo in form", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo: "user2/repo1",
		}

		ctx := makeLocalContext(t)
		ctx.Data["Form"] = form
		ui.populateTemplateContext(ctx)
		require.False(t, ctx.Written()) // no error written to ctx

		repo := ctx.Data["SourceRepo"].(*repo_model.Repository)
		assert.EqualValues(t, 1, repo.ID)
	})
}

func TestLocalPopulateError(t *testing.T) {
	ui := actionsLocalUI{}

	t.Run("unrecognized error", func(t *testing.T) {
		ctx := makeLocalContext(t)
		assert.False(t, ui.populateError(ctx, errors.New("some other error")))
	})

	t.Run("unrecognized field error", func(t *testing.T) {
		ctx := makeLocalContext(t)
		assert.False(t, ui.populateError(ctx, &auth_service.MissingFieldError{Field: "Description"}))
	})

	t.Run("workflow field error", func(t *testing.T) {
		ctx := makeLocalContext(t)
		assert.True(t, ui.populateError(ctx, errInvalidWorkflowFileGlob))
		assert.True(t, ctx.Data["Err_WorkflowFile"].(bool))
	})

	t.Run("git ref field error", func(t *testing.T) {
		ctx := makeLocalContext(t)
		assert.True(t, ui.populateError(ctx, errInvalidGitRefGlob))
		assert.True(t, ctx.Data["Err_GitRef"].(bool))
	})
}

func TestLocalPopulateForm(t *testing.T) {
	form := &actionsLocalAuthorizedIntegrationForm{}
	issuer := "urn:forgejo:authorized-integrations:actions"

	t.Run("fully populated claim rules", func(t *testing.T) {
		cr := &auth_model.ClaimRules{
			Rules: []auth_model.ClaimRule{
				{Claim: "repository_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "repository_owner_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "workflow", Comparison: auth_model.ClaimGlob, Value: ".forgejo/workflows/*.yml"},
				{Claim: "ref", Comparison: auth_model.ClaimGlob, Value: "refs/tags/v*"},
				{Claim: "event_name", Comparison: auth_model.ClaimIn, Values: []string{"push"}},
			},
		}
		require.NoError(t, form.populateForm(makeLocalContext(t), issuer, cr))
		assert.Equal(t, "user2/repo2", form.SourceRepo)
		assert.Equal(t, ".forgejo/workflows/*.yml", form.WorkflowFile)
		assert.Equal(t, "refs/tags/v*", form.GitRef)
		assert.Equal(t, []string{"push"}, form.Event)
	})

	t.Run("mismatched repository owner", func(t *testing.T) {
		cr := &auth_model.ClaimRules{
			Rules: []auth_model.ClaimRule{
				{Claim: "repository_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "repository_owner_id", Comparison: auth_model.ClaimEqual, Value: "200"},
			},
		}
		require.ErrorContains(t, form.populateForm(makeLocalContext(t), issuer, cr), "repository could not be loaded")
	})

	t.Run("repo with no visibility", func(t *testing.T) {
		cr := &auth_model.ClaimRules{
			Rules: []auth_model.ClaimRule{
				{Claim: "repository_id", Comparison: auth_model.ClaimEqual, Value: "7"},
				{Claim: "repository_owner_id", Comparison: auth_model.ClaimEqual, Value: "10"},
			},
		}
		require.ErrorContains(t, form.populateForm(makeLocalContext(t), issuer, cr), "repository could not be loaded")
	})
}

func TestLocalConvertForm(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{}
		_, _, err := form.convertForm(makeLocalContext(t))
		require.ErrorContains(t, err, "missing field SourceRepo")
	})

	t.Run("repo with no visibilty", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo: "user10/repo7",
		}
		_, _, err := form.convertForm(makeLocalContext(t))
		require.ErrorContains(t, err, "one or more of the repositories couldn't be found by owner & name")
	})

	t.Run("valid repo", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo: "user2/repo2",
		}
		issuer, claimRules, err := form.convertForm(makeLocalContext(t))
		require.NoError(t, err)
		assert.Equal(t, "urn:forgejo:authorized-integrations:actions", issuer)
		assert.Equal(t, &auth_model.ClaimRules{
			Rules: []auth_model.ClaimRule{
				{Claim: "repository_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "repository_owner_id", Comparison: auth_model.ClaimEqual, Value: "2"},
			},
		}, claimRules)
	})

	t.Run("valid workflow file", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo:   "user2/repo2",
			WorkflowFile: ".forgejo/workflows/test-*.yml",
		}
		issuer, claimRules, err := form.convertForm(makeLocalContext(t))
		require.NoError(t, err)
		assert.Equal(t, "urn:forgejo:authorized-integrations:actions", issuer)
		assert.Equal(t, &auth_model.ClaimRules{
			Rules: []auth_model.ClaimRule{
				{Claim: "repository_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "repository_owner_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "workflow", Comparison: auth_model.ClaimGlob, Value: ".forgejo/workflows/test-*.yml"},
			},
		}, claimRules)
	})

	t.Run("invalid workflow file", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo:   "user2/repo2",
			WorkflowFile: ".forgejo/workflows/test-*[",
		}
		_, _, err := form.convertForm(makeLocalContext(t))
		require.ErrorContains(t, err, "invalid workflow file glob: unexpected end of input")
	})

	t.Run("valid git ref", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo: "user2/repo2",
			GitRef:     "refs/tags/v*",
		}
		issuer, claimRules, err := form.convertForm(makeLocalContext(t))
		require.NoError(t, err)
		assert.Equal(t, "urn:forgejo:authorized-integrations:actions", issuer)
		assert.Equal(t, &auth_model.ClaimRules{
			Rules: []auth_model.ClaimRule{
				{Claim: "repository_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "repository_owner_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "ref", Comparison: auth_model.ClaimGlob, Value: "refs/tags/v*"},
			},
		}, claimRules)
	})

	t.Run("invalid git ref", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo: "user2/repo2",
			GitRef:     "refs/[",
		}
		_, _, err := form.convertForm(makeLocalContext(t))
		require.ErrorContains(t, err, "invalid git ref glob: unexpected end of input")
	})

	t.Run("valid event", func(t *testing.T) {
		form := &actionsLocalAuthorizedIntegrationForm{
			SourceRepo: "user2/repo2",
			Event:      []string{"push", "pull_request"},
		}
		issuer, claimRules, err := form.convertForm(makeLocalContext(t))
		require.NoError(t, err)
		assert.Equal(t, "urn:forgejo:authorized-integrations:actions", issuer)
		assert.Equal(t, &auth_model.ClaimRules{
			Rules: []auth_model.ClaimRule{
				{Claim: "repository_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "repository_owner_id", Comparison: auth_model.ClaimEqual, Value: "2"},
				{Claim: "event_name", Comparison: auth_model.ClaimIn, Values: []string{"push", "pull_request"}},
			},
		}, claimRules)
	})
}
