// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import (
	"fmt"
	"strings"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/models/db"
	access_model "forgejo.org/models/perm/access"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/services/context"
)

type baseAuthorizedIntegrationForm struct {
	// Top data in UI, descriptive information about the Authorized Integration:
	Name        string
	Description string
	Audience    string

	// // Middle data in UI, how JWTs are validated by this Authorized Integration:
	// Issuer     string // Future: Issuer is likely to be replaced with more-specific fields on non-generic UIs
	// ClaimRules string // Future: ClaimRules is only required when aiUI == "generic"

	// Bottom data in the UI, what authorization is permitted by this Authorized Integration:
	Resource     string   // all, public-only, repo-specific
	SelectedRepo []string // slice of ownername/reponame for repo-specific
	ScopeAll     bool
	Scope        []string

	// Values used for repo-specific repository multi-select UI, not stored in Authorized Integration:
	RepoSearch         string
	AddSelectedRepo    string // add a repo to SelectedRepo
	RemoveSelectedRepo string // remove a repo from SelectedRepo
	Page               int    // repo search page
	SetPage            int    // repo search buttons
}

func (f *baseAuthorizedIntegrationForm) isEmpty() bool {
	return f.Name == "" && f.Description == "" && f.Audience == "" &&
		f.Resource == "" && f.SelectedRepo == nil && f.Scope == nil
}

func (f *baseAuthorizedIntegrationForm) copyAuthorizedIntegrationToForm(ctx *context.Context, ai *auth_model.AuthorizedIntegration, rr []*auth_model.AuthorizedIntegResourceRepo) error {
	f.Name = ai.Name
	f.Description = ai.Description
	f.Audience = ai.Audience

	if ai.ResourceAllRepos {
		publicOnly, err := ai.Scope.PublicOnly()
		if err != nil {
			return err
		}
		if publicOnly {
			f.Resource = "public-only"
		} else {
			f.Resource = "all"
		}
	} else {
		f.Resource = "repo-specific"
	}

	f.Scope = ai.Scope.StringSlice()
	scopeAll, err := ai.Scope.HasScope(auth_model.AccessTokenScopeAll)
	if err != nil {
		return err
	}
	f.ScopeAll = scopeAll

	f.SelectedRepo = []string{}
	if len(rr) != 0 {
		repoIDs := make([]int64, len(rr))
		for i, r := range rr {
			repoIDs[i] = r.RepoID
		}
		repos, err := db.GetByIDs(ctx, "id", repoIDs, &repo_model.Repository{})
		if err != nil {
			return err
		}
		for _, r := range rr {
			repo := repos[r.RepoID]
			// Repos associated with an authorized integration should already be visible to the owner, but it's possible
			// that access has changed, such as a removed collaborator on a repo -- don't provide info on that repo if
			// so.
			permission, err := access_model.GetUserRepoPermission(ctx, repo, ctx.Doer)
			if err != nil {
				return err
			}
			if permission.HasAccess() {
				f.SelectedRepo = append(f.SelectedRepo, fmt.Sprintf("%s/%s", repo.OwnerName, repo.Name))
			}
		}
	}

	return nil
}

func (f *baseAuthorizedIntegrationForm) copyFormToAuthorizedIntegration(ctx *context.Context, ai *auth_model.AuthorizedIntegration) ([]*auth_model.AuthorizedIntegResourceRepo, error) {
	ai.Name = f.Name
	ai.Description = f.Description

	scopeRaw := strings.Join(f.Scope, ",")
	var resourceRepos []*auth_model.AuthorizedIntegResourceRepo
	switch f.Resource {
	case "all":
		ai.ResourceAllRepos = true
	case "public-only":
		ai.ResourceAllRepos = true
		scopeRaw = fmt.Sprintf("%s,%s", scopeRaw, auth_model.AccessTokenScopePublicOnly)
	case "repo-specific":
		ai.ResourceAllRepos = false
		selectedRepos, err := getSelectedRepos(ctx, f.SelectedRepo)
		if err != nil {
			return nil, err
		}
		for _, repo := range selectedRepos {
			resourceRepos = append(resourceRepos, &auth_model.AuthorizedIntegResourceRepo{RepoID: repo.ID})
		}
	}

	scope, err := auth_model.AccessTokenScope(scopeRaw).Normalize()
	if err != nil {
		return nil, err
	}
	ai.Scope = scope

	return resourceRepos, nil
}

func (f *baseAuthorizedIntegrationForm) InitNew() {
	f.Resource = "all"
}
