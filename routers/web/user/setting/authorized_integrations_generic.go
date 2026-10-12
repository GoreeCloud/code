// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package setting

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"

	auth_model "forgejo.org/models/auth"
	"forgejo.org/modules/base"
	"forgejo.org/modules/json"
	"forgejo.org/modules/svg"
	"forgejo.org/modules/templates"
	auth_service "forgejo.org/services/auth"
	"forgejo.org/services/context"
)

var (
	_ authorizedIntegrationUIImpl = genericUI{}
	_ authorizedIntegrationUIForm = &genericAuthorizedIntegrationForm{}
)

type genericUI struct{}

func (genericUI) UIIdentifier() auth_model.AuthorizedIntegrationUI {
	return auth_model.AuthorizedIntegrationUIGeneric
}

func (genericUI) Icon(size int) template.HTML {
	return svg.RenderHTML("octicon-cloud", size, "img")
}

func (genericUI) Label(ctx *templates.Context) template.HTML {
	return ctx.Locale.Tr("settings.authorized_integration.ui.generic")
}

func (genericUI) editTemplate() base.TplName {
	return "user/settings/authorized_integrations/generic/view"
}

func (genericUI) populateTemplateContext(ctx *context.Context) {
}

func (genericUI) form() authorizedIntegrationUIForm {
	return &genericAuthorizedIntegrationForm{}
}

func (genericUI) populateError(ctx *context.Context, err error) (handled bool) {
	switch {
	case errors.Is(err, auth_service.ErrInvalidIssuer):
		ctx.Data["Err_Issuer"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.issuer.invalid", err.Error()), true)
		return true
	case errors.Is(err, auth_service.ErrInvalidClaimRules):
		ctx.Data["Err_ClaimRules"] = true
		ctx.Flash.Error(ctx.Tr("settings.authorized_integration.claim_rules.invalid", err.Error()), true)
		return true
	}
	return false
}

type genericAuthorizedIntegrationForm struct {
	baseAuthorizedIntegrationForm
	Issuer     string
	ClaimRules string
}

func (g *genericAuthorizedIntegrationForm) baseForm() *baseAuthorizedIntegrationForm {
	return &g.baseAuthorizedIntegrationForm
}

func (g *genericAuthorizedIntegrationForm) isEmpty() bool {
	return g.baseAuthorizedIntegrationForm.isEmpty() && g.Issuer == "" && g.ClaimRules == ""
}

func (g *genericAuthorizedIntegrationForm) populateForm(ctx *context.Context, issuer string, claimRules *auth_model.ClaimRules) error {
	g.Issuer = issuer
	claimRulesJSON, err := json.MarshalIndent(claimRules, "", "  ")
	if err != nil {
		return err
	}
	g.ClaimRules = string(claimRulesJSON)
	return nil
}

func (g *genericAuthorizedIntegrationForm) convertForm(ctx *context.Context) (issuer string, claimRules *auth_model.ClaimRules, err error) {
	issuer = g.Issuer

	reader := bytes.NewReader([]byte(g.ClaimRules))
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields() // prevent typo fields from being ignored to make errors easier to identify
	if err := decoder.Decode(&claimRules); err != nil {
		return "", nil, fmt.Errorf("%w: %w", auth_service.ErrInvalidClaimRules, err)
	}
	// json.Decoder doesn't guarantee that all of the reader is consumed, which can lead to weird situations
	// where the UI appears to work correctly if extra content is in the form field, but it won't be parsed,
	// misleading users.  Detect if anything other than io.EOF comes out of further decodings:
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return "", nil, fmt.Errorf("%w: unexpected trailing content: %s", auth_service.ErrInvalidClaimRules, extra)
		}
		return "", nil, fmt.Errorf("%w: error after JSON value: %w", auth_service.ErrInvalidClaimRules, err)
	}

	return issuer, claimRules, nil
}

func (g *genericAuthorizedIntegrationForm) initNew() {
	g.ClaimRules = string("{\n  \"rules\":[]\n}")
}
