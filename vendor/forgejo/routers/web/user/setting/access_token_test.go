// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/templates"
	"forgejo.org/modules/web"
	"forgejo.org/services/context"
	"forgejo.org/services/contexttest"
	"forgejo.org/services/forms"

	"code.forgejo.org/go-chi/binding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessTokenCreate(t *testing.T) {
	unittest.PrepareTestEnv(t)

	ctx, resp := contexttest.MockContext(t, "user/settings/applications/tokens/new",
		contexttest.MockContextOption{Render: templates.HTMLRenderer()})
	contexttest.LoadUser(t, ctx, 2)

	web.SetForm(ctx, &forms.NewAccessTokenGetForm{})
	AccessTokenCreate(ctx)

	assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
	assert.False(t, ctx.HasError(), "error: %s", ctx.GetErrMsg())

	// check empty-form GET ctx.Data values
	assert.Equal(t, "name", ctx.Data["Autofocus"])
	assert.Len(t, ctx.Data["Repos"], 10)
	assert.Empty(t, ctx.Data["SelectedRepos"])

	// check that repo in the search is rendered in the content
	assert.Contains(t, resp.Body.String(), "org17/big_test_private_4")
}

func TestAccessTokenCreatePost(t *testing.T) {
	unittest.PrepareTestEnv(t)

	post := func(t *testing.T, form *forms.NewAccessTokenPostForm) (*context.Context, *httptest.ResponseRecorder) {
		t.Helper()

		ctx, resp := contexttest.MockContext(t, "user/settings/applications/tokens/new",
			contexttest.MockContextOption{Render: templates.HTMLRenderer()})
		contexttest.LoadUser(t, ctx, 2)

		ctx.AppendContextValue(context.WebContextKey, ctx)
		binding.Bind(ctx.Req.WithContext(ctx), form)
		web.SetForm(ctx, form)
		AccessTokenCreatePost(ctx)

		return ctx, resp
	}

	render := func(t *testing.T, form *forms.NewAccessTokenPostForm) (*context.Context, string) {
		t.Helper()
		ctx, resp := post(t, form)
		assert.Equal(t, http.StatusOK, resp.Result().StatusCode)
		return ctx, resp.Body.String()
	}

	t.Run("retains form info on missing token name", func(t *testing.T) {
		ctx, body := render(t, &forms.NewAccessTokenPostForm{
			Name:         "", // absent
			Resource:     "repo-specific",
			Scope:        []string{"read:repository"},
			SelectedRepo: []string{"org17/big_test_private_4"},
		})
		require.Contains(t, body, "settings.token_nameform.require_error")

		assert.Equal(t, "repo-specific", ctx.Data["resource"])
		assert.Contains(t, ctx.Data["scope"], "read:repository")
		assert.True(t, slices.ContainsFunc(ctx.Data["SelectedRepos"].([]*repo_model.Repository), func(r *repo_model.Repository) bool {
			return r.OwnerName == "org17" && r.Name == "big_test_private_4"
		}), "SelectedRepos missing org17/big_test_private_4")
	})

	t.Run("retains form info on missing scopes", func(t *testing.T) {
		ctx, body := render(t, &forms.NewAccessTokenPostForm{
			Name:         "my new token",
			Resource:     "repo-specific",
			Scope:        []string{"", "", ""}, // absent
			SelectedRepo: []string{"org17/big_test_private_4"},
		})
		require.Contains(t, body, "settings.at_least_one_permission")

		assert.Equal(t, "my new token", ctx.Data["name"])
		assert.Equal(t, "repo-specific", ctx.Data["resource"])
		assert.True(t, slices.ContainsFunc(ctx.Data["SelectedRepos"].([]*repo_model.Repository), func(r *repo_model.Repository) bool {
			return r.OwnerName == "org17" && r.Name == "big_test_private_4"
		}), "SelectedRepos missing org17/big_test_private_4")
	})

	t.Run("retains form info on duplicate token name", func(t *testing.T) {
		ctx, body := render(t, &forms.NewAccessTokenPostForm{
			Name:         "Token A", // duplicate
			Resource:     "repo-specific",
			Scope:        []string{"read:repository"},
			SelectedRepo: []string{"org17/big_test_private_4"},
		})
		require.Contains(t, body, "settings.generate_token_name_duplicate")

		assert.Equal(t, "Token A", ctx.Data["name"])
		assert.Equal(t, "repo-specific", ctx.Data["resource"])
		assert.Contains(t, ctx.Data["scope"], "read:repository")
		assert.True(t, slices.ContainsFunc(ctx.Data["SelectedRepos"].([]*repo_model.Repository), func(r *repo_model.Repository) bool {
			return r.OwnerName == "org17" && r.Name == "big_test_private_4"
		}), "SelectedRepos missing org17/big_test_private_4")
	})

	t.Run("retains form info on ValidateAccessToken error", func(t *testing.T) {
		ctx, body := render(t, &forms.NewAccessTokenPostForm{
			Name:         "my new token",
			Resource:     "repo-specific",
			Scope:        []string{"read:admin"}, // not permitted for repo-specific
			SelectedRepo: []string{"org17/big_test_private_4"},
		})
		require.Contains(t, body, "access_token.error.specified_repos_and_invalid_scope")

		assert.Equal(t, "my new token", ctx.Data["name"])
		assert.Equal(t, "repo-specific", ctx.Data["resource"])
		assert.Contains(t, ctx.Data["scope"], "read:admin")
		assert.True(t, slices.ContainsFunc(ctx.Data["SelectedRepos"].([]*repo_model.Repository), func(r *repo_model.Repository) bool {
			return r.OwnerName == "org17" && r.Name == "big_test_private_4"
		}), "SelectedRepos missing org17/big_test_private_4")
	})

	t.Run("invalid repo selected", func(t *testing.T) {
		_, resp := post(t, &forms.NewAccessTokenPostForm{
			Name:         "my new token",
			Resource:     "repo-specific",
			Scope:        []string{"read:admin"}, // not permitted for repo-specific
			SelectedRepo: []string{"org17/big_test_private_4000000_does_not_exist"},
		})
		assert.Equal(t, http.StatusBadRequest, resp.Result().StatusCode)
		assert.Equal(t, "GetRepositoryByOwnerAndName\n", resp.Body.String())
	})

	t.Run("non-visible repo selected via IDOR", func(t *testing.T) {
		_, resp := post(t, &forms.NewAccessTokenPostForm{
			Name:         "my new token",
			Resource:     "repo-specific",
			Scope:        []string{"read:repository"},
			SelectedRepo: []string{"user30/empty"}, // private repo, user2 has no visibility
		})
		assert.Equal(t, http.StatusBadRequest, resp.Result().StatusCode)
		assert.Equal(t, "GetRepositoryByOwnerAndName\n", resp.Body.String()) // should be exact same response as "invalid repo selected" case
	})
}
