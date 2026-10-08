// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd

import (
	"context"
	"testing"

	"forgejo.org/models/auth"
	"forgejo.org/modules/test"
	"forgejo.org/services/auth/source/oauth2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestAddOauth(t *testing.T) {
	// Mock cli functions to do not exit on error
	defer test.MockVariableValue(&cli.OsExiter, func(code int) {})()

	// Test cases
	cases := []struct {
		args   []string
		source *auth.Source
		errMsg string
	}{
		// case 0
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 (via openidConnect) source full",
				"--provider", "openidConnect",
				"--key", "client id",
				"--secret", "client secret",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--use-custom-urls", "",
				"--custom-tenant-id", "tenant id",
				"--custom-auth-url", "https://example.com/auth",
				"--custom-token-url", "https://example.com/token",
				"--custom-profile-url", "https://example.com/profile",
				"--custom-email-url", "https://example.com/email",
				"--icon-url", "https://example.com/icon.svg",
				"--skip-local-2fa",
				"--scopes", "address",
				"--scopes", "email",
				"--scopes", "phone",
				"--scopes", "profile",
				"--attribute-ssh-public-key", "ssh_public_key",
				"--required-claim-name", "can_access",
				"--required-claim-value", "yes",
				"--group-claim-name", "groups",
				"--admin-group", "admin",
				"--restricted-group", "restricted",
				"--group-team-map", `{"org_a_team_1": {"organization-a": ["Team 1"]}, "org_a_all_teams": {"organization-a": ["Team 1", "Team 2", "Team 3"]}}`,
				"--group-team-map-removal",
				"--allow-username-change",
				"--quota-group-claim-name", "quota_groups",
				"--quota-group-map", `{"oauth_group_1": ["quota_group_1"], "oauth_group_2": ["quota_group_2"]}`,
				"--quota-group-map-removal",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 (via openidConnect) source full",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					ClientID:                      "client id",
					ClientSecret:                  "client secret",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					CustomURLMapping: &oauth2.CustomURLMapping{
						AuthURL:    "https://example.com/auth",
						TokenURL:   "https://example.com/token",
						ProfileURL: "https://example.com/profile",
						EmailURL:   "https://example.com/email",
						Tenant:     "tenant id",
					},
					IconURL:               "https://example.com/icon.svg",
					Scopes:                []string{"address", "email", "phone", "profile"},
					AttributeSSHPublicKey: "ssh_public_key",
					RequiredClaimName:     "can_access",
					RequiredClaimValue:    "yes",
					GroupClaimName:        "groups",
					AdminGroup:            "admin",
					GroupTeamMap:          `{"org_a_team_1": {"organization-a": ["Team 1"]}, "org_a_all_teams": {"organization-a": ["Team 1", "Team 2", "Team 3"]}}`,
					GroupTeamMapRemoval:   true,
					QuotaGroupClaimName:   "quota_groups",
					QuotaGroupMap:         `{"oauth_group_1": ["quota_group_1"], "oauth_group_2": ["quota_group_2"]}`,
					QuotaGroupMapRemoval:  true,
					RestrictedGroup:       "restricted",
					SkipLocalTwoFA:        true,
					AllowUsernameChange:   true,
				},
			},
		},
		// case 1
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 (via openidConnect) source min",
				"--provider", "openidConnect",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 (via openidConnect) source min",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					Scopes:                        []string{},
				},
			},
		},
		// case 2
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 (via openidConnect) source `--use-custom-urls` required for `--custom-*` flags",
				"--custom-tenant-id", "tenant id",
				"--custom-auth-url", "https://example.com/auth",
				"--custom-token-url", "https://example.com/token",
				"--custom-profile-url", "https://example.com/profile",
				"--custom-email-url", "https://example.com/email",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 (via openidConnect) source `--use-custom-urls` required for `--custom-*` flags",
				IsActive: true,
				Cfg: &oauth2.Source{
					Scopes: []string{},
				},
			},
		},
		// case 3
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 (via openidConnect) source `--scopes` aggregates multiple uses",
				"--provider", "openidConnect",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--scopes", "address",
				"--scopes", "email",
				"--scopes", "phone",
				"--scopes", "profile",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 (via openidConnect) source `--scopes` aggregates multiple uses",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					Scopes:                        []string{"address", "email", "phone", "profile"},
				},
			},
		},
		// case 4
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 (via openidConnect) source `--scopes` supports commas as separators",
				"--provider", "openidConnect",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--scopes", "address,email,phone,profile",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 (via openidConnect) source `--scopes` supports commas as separators",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					Scopes:                        []string{"address", "email", "phone", "profile"},
				},
			},
		},
		// case 5
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 (via openidConnect) source",
				"--provider", "openidConnect",
			},
			errMsg: "invalid Auto Discovery URL:  (this must be a valid URL starting with http:// or https://)",
		},
		// case 6
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 (via openidConnect) source",
				"--provider", "openidConnect",
				"--auto-discover-url", "example.com",
			},
			errMsg: "invalid Auto Discovery URL: example.com (this must be a valid URL starting with http:// or https://)",
		},
		// case 7
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 source with quota group claim name",
				"--provider", "openidConnect",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--quota-group-claim-name", "quota_groups",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 source with quota group claim name",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					Scopes:                        []string{},
					QuotaGroupClaimName:           "quota_groups",
				},
			},
		},
		// case 8
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 source with quota group map",
				"--provider", "openidConnect",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--quota-group-map", `{"oauth_group_1": ["quota_group_1"], "oauth_group_2": ["quota_group_2"]}`,
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 source with quota group map",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					Scopes:                        []string{},
					QuotaGroupMap:                 `{"oauth_group_1": ["quota_group_1"], "oauth_group_2": ["quota_group_2"]}`,
				},
			},
		},
		// case 9
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 source with quota group map removal",
				"--provider", "openidConnect",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--quota-group-map-removal",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 source with quota group map removal",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					Scopes:                        []string{},
					QuotaGroupMapRemoval:          true,
				},
			},
		},
		// case 10
		{
			args: []string{
				"oauth-test",
				"--name", "oauth2 source with all quota group fields",
				"--provider", "openidConnect",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--quota-group-claim-name", "quota_groups",
				"--quota-group-map", `{"developers": ["dev_quota"], "admins": ["admin_quota"]}`,
				"--quota-group-map-removal",
			},
			source: &auth.Source{
				Type:     auth.OAuth2,
				Name:     "oauth2 source with all quota group fields",
				IsActive: true,
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					Scopes:                        []string{},
					QuotaGroupClaimName:           "quota_groups",
					QuotaGroupMap:                 `{"developers": ["dev_quota"], "admins": ["admin_quota"]}`,
					QuotaGroupMapRemoval:          true,
				},
			},
		},
	}

	for n, c := range cases {
		// Mock functions.
		var createdAuthSource *auth.Source
		service := &authService{
			initDB: func(context.Context) error {
				return nil
			},
			createAuthSource: func(ctx context.Context, authSource *auth.Source) error {
				createdAuthSource = authSource
				return nil
			},
			updateAuthSource: func(ctx context.Context, authSource *auth.Source) error {
				assert.FailNow(t, "should not call updateAuthSource", "case: %d", n)
				return nil
			},
			getAuthSourceByID: func(ctx context.Context, id int64) (*auth.Source, error) {
				assert.FailNow(t, "should not call getAuthSourceByID", "case: %d", n)
				return nil, nil
			},
		}

		// Create a copy of command to test
		app := cli.Command{}
		app.Flags = microcmdAuthAddOauth().Flags
		app.Action = service.addOauth

		// Run it
		err := app.Run(t.Context(), c.args)
		if c.errMsg != "" {
			assert.EqualError(t, err, c.errMsg, "case %d: error should match", n)
		} else {
			require.NoError(t, err, "case %d: should have no errors", n)
			assert.Equal(t, c.source, createdAuthSource, "case %d: wrong authSource", n)
		}
	}
}

func TestUpdateOauth(t *testing.T) {
	// Mock cli functions to do not exit on error
	defer test.MockVariableValue(&cli.OsExiter, func(code int) {})()

	// Test cases
	cases := []struct {
		args               []string
		id                 int64
		existingAuthSource *auth.Source
		authSource         *auth.Source
		errMsg             string
	}{
		// case 0
		{
			args: []string{
				"oauth-test",
				"--id", "23",
				"--name", "oauth2 (via openidConnect) source full",
				"--provider", "openidConnect",
				"--key", "client id",
				"--secret", "client secret",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
				"--use-custom-urls", "",
				"--custom-tenant-id", "tenant id",
				"--custom-auth-url", "https://example.com/auth",
				"--custom-token-url", "https://example.com/token",
				"--custom-profile-url", "https://example.com/profile",
				"--custom-email-url", "https://example.com/email",
				"--icon-url", "https://example.com/icon.svg",
				"--skip-local-2fa",
				"--scopes", "address",
				"--scopes", "email",
				"--scopes", "phone",
				"--scopes", "profile",
				"--attribute-ssh-public-key", "ssh_public_key",
				"--required-claim-name", "can_access",
				"--required-claim-value", "yes",
				"--group-claim-name", "groups",
				"--admin-group", "admin",
				"--restricted-group", "restricted",
				"--group-team-map", `{"org_a_team_1": {"organization-a": ["Team 1"]}, "org_a_all_teams": {"organization-a": ["Team 1", "Team 2", "Team 3"]}}`,
				"--group-team-map-removal",
			},
			id: 23,
			existingAuthSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg:  &oauth2.Source{},
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Name: "oauth2 (via openidConnect) source full",
				Cfg: &oauth2.Source{
					Provider:                      "openidConnect",
					ClientID:                      "client id",
					ClientSecret:                  "client secret",
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					CustomURLMapping: &oauth2.CustomURLMapping{
						AuthURL:    "https://example.com/auth",
						TokenURL:   "https://example.com/token",
						ProfileURL: "https://example.com/profile",
						EmailURL:   "https://example.com/email",
						Tenant:     "tenant id",
					},
					IconURL:               "https://example.com/icon.svg",
					Scopes:                []string{"address", "email", "phone", "profile"},
					AttributeSSHPublicKey: "ssh_public_key",
					RequiredClaimName:     "can_access",
					RequiredClaimValue:    "yes",
					GroupClaimName:        "groups",
					AdminGroup:            "admin",
					GroupTeamMap:          `{"org_a_team_1": {"organization-a": ["Team 1"]}, "org_a_all_teams": {"organization-a": ["Team 1", "Team 2", "Team 3"]}}`,
					GroupTeamMapRemoval:   true,
					RestrictedGroup:       "restricted",
					// `--skip-local-2fa` is currently ignored.
					// SkipLocalTwoFA:        true,
				},
			},
		},
		// case 1
		{
			args: []string{
				"oauth-test",
				"--id", "1",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
				},
			},
		},
		// case 2
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--name", "oauth2 (via openidConnect) source full",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Name: "oauth2 (via openidConnect) source full",
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
				},
			},
		},
		// case 3
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--provider", "openidConnect",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					Provider:         "openidConnect",
					CustomURLMapping: &oauth2.CustomURLMapping{},
				},
			},
		},
		// case 4
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--key", "client id",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					ClientID:         "client id",
					CustomURLMapping: &oauth2.CustomURLMapping{},
				},
			},
		},
		// case 5
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--secret", "client secret",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					ClientSecret:     "client secret",
					CustomURLMapping: &oauth2.CustomURLMapping{},
				},
			},
		},
		// case 6
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--auto-discover-url", "https://example.com/.well-known/openid-configuration",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					OpenIDConnectAutoDiscoveryURL: "https://example.com/.well-known/openid-configuration",
					CustomURLMapping:              &oauth2.CustomURLMapping{},
				},
			},
		},
		// case 7
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--use-custom-urls", "",
				"--custom-tenant-id", "tenant id",
				"--custom-auth-url", "https://example.com/auth",
				"--custom-token-url", "https://example.com/token",
				"--custom-profile-url", "https://example.com/profile",
				"--custom-email-url", "https://example.com/email",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{
						AuthURL:    "https://example.com/auth",
						TokenURL:   "https://example.com/token",
						ProfileURL: "https://example.com/profile",
						EmailURL:   "https://example.com/email",
						Tenant:     "tenant id",
					},
				},
			},
		},
		// case 8
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--name", "oauth2 (via openidConnect) source `--use-custom-urls` required for `--custom-*` flags",
				"--custom-tenant-id", "tenant id",
				"--custom-auth-url", "https://example.com/auth",
				"--custom-token-url", "https://example.com/token",
				"--custom-profile-url", "https://example.com/profile",
				"--custom-email-url", "https://example.com/email",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Name: "oauth2 (via openidConnect) source `--use-custom-urls` required for `--custom-*` flags",
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
				},
			},
		},
		// case 9
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--icon-url", "https://example.com/icon.svg",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					IconURL:          "https://example.com/icon.svg",
				},
			},
		},
		// case 10
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--name", "oauth2 (via openidConnect) source `--skip-local-2fa` is currently ignored",
				"--skip-local-2fa",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Name: "oauth2 (via openidConnect) source `--skip-local-2fa` is currently ignored",
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					// `--skip-local-2fa` is currently ignored.
					// SkipLocalTwoFA: true,
				},
			},
		},
		// case 11
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--name", "oauth2 (via openidConnect) source `--scopes` aggregates multiple uses",
				"--scopes", "address",
				"--scopes", "email",
				"--scopes", "phone",
				"--scopes", "profile",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Name: "oauth2 (via openidConnect) source `--scopes` aggregates multiple uses",
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					Scopes:           []string{"address", "email", "phone", "profile"},
				},
			},
		},
		// case 12
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--name", "oauth2 (via openidConnect) source `--scopes` supports commas as separators",
				"--scopes", "address,email,phone,profile",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Name: "oauth2 (via openidConnect) source `--scopes` supports commas as separators",
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					Scopes:           []string{"address", "email", "phone", "profile"},
				},
			},
		},
		// case 13
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--attribute-ssh-public-key", "ssh_public_key",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:      &oauth2.CustomURLMapping{},
					AttributeSSHPublicKey: "ssh_public_key",
				},
			},
		},
		// case 14
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--required-claim-name", "can_access",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:  &oauth2.CustomURLMapping{},
					RequiredClaimName: "can_access",
				},
			},
		},
		// case 15
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--required-claim-value", "yes",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:   &oauth2.CustomURLMapping{},
					RequiredClaimValue: "yes",
				},
			},
		},
		// case 16
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--group-claim-name", "groups",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					GroupClaimName:   "groups",
				},
			},
		},
		// case 17
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--admin-group", "admin",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					AdminGroup:       "admin",
				},
			},
		},
		// case 18
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--restricted-group", "restricted",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					RestrictedGroup:  "restricted",
				},
			},
		},
		// case 19
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--group-team-map", `{"org_a_team_1": {"organization-a": ["Team 1"]}, "org_a_all_teams": {"organization-a": ["Team 1", "Team 2", "Team 3"]}}`,
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					GroupTeamMap:     `{"org_a_team_1": {"organization-a": ["Team 1"]}, "org_a_all_teams": {"organization-a": ["Team 1", "Team 2", "Team 3"]}}`,
				},
			},
		},
		// case 20
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--group-team-map-removal",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:    &oauth2.CustomURLMapping{},
					GroupTeamMapRemoval: true,
				},
			},
		},
		// case 21
		{
			args: []string{
				"oauth-test",
				"--id", "23",
				"--group-team-map-removal=false",
			},
			id: 23,
			existingAuthSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					GroupTeamMapRemoval: true,
				},
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:    &oauth2.CustomURLMapping{},
					GroupTeamMapRemoval: false,
				},
			},
		},
		// case 22
		{
			args: []string{
				"oauth-test",
			},
			errMsg: "--id flag is missing",
		},
		// case 23
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--quota-group-claim-name", "quota_groups",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:    &oauth2.CustomURLMapping{},
					QuotaGroupClaimName: "quota_groups",
				},
			},
		},
		// case 24
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--quota-group-map", `{"oauth_group_1": ["quota_group_1"], "oauth_group_2": ["quota_group_2"]}`,
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping: &oauth2.CustomURLMapping{},
					QuotaGroupMap:    `{"oauth_group_1": ["quota_group_1"], "oauth_group_2": ["quota_group_2"]}`,
				},
			},
		},
		// case 25
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--quota-group-map-removal",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:     &oauth2.CustomURLMapping{},
					QuotaGroupMapRemoval: true,
				},
			},
		},
		// case 26
		{
			args: []string{
				"oauth-test",
				"--id", "24",
				"--quota-group-map-removal=false",
			},
			id: 24,
			existingAuthSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					QuotaGroupMapRemoval: true,
				},
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:     &oauth2.CustomURLMapping{},
					QuotaGroupMapRemoval: false,
				},
			},
		},
		// case 27
		{
			args: []string{
				"oauth-test",
				"--id", "1",
				"--quota-group-claim-name", "quota_groups",
				"--quota-group-map", `{"developers": ["dev_quota"], "admins": ["admin_quota"]}`,
				"--quota-group-map-removal",
			},
			authSource: &auth.Source{
				Type: auth.OAuth2,
				Cfg: &oauth2.Source{
					CustomURLMapping:     &oauth2.CustomURLMapping{},
					QuotaGroupClaimName:  "quota_groups",
					QuotaGroupMap:        `{"developers": ["dev_quota"], "admins": ["admin_quota"]}`,
					QuotaGroupMapRemoval: true,
				},
			},
		},
	}

	for n, c := range cases {
		// Mock functions.
		var updatedAuthSource *auth.Source
		service := &authService{
			initDB: func(context.Context) error {
				return nil
			},
			createAuthSource: func(ctx context.Context, authSource *auth.Source) error {
				assert.FailNow(t, "should not call createAuthSource", "case: %d", n)
				return nil
			},
			updateAuthSource: func(ctx context.Context, authSource *auth.Source) error {
				updatedAuthSource = authSource
				return nil
			},
			getAuthSourceByID: func(ctx context.Context, id int64) (*auth.Source, error) {
				if c.id != 0 {
					assert.Equal(t, c.id, id, "case %d: wrong id", n)
				}
				if c.existingAuthSource != nil {
					return c.existingAuthSource, nil
				}
				return &auth.Source{
					Type: auth.OAuth2,
					Cfg:  &oauth2.Source{},
				}, nil
			},
		}

		// Create a copy of command to test
		app := cli.Command{}
		app.Flags = microcmdAuthUpdateOauth().Flags
		app.Action = service.updateOauth

		// Run it
		err := app.Run(t.Context(), c.args)
		if c.errMsg != "" {
			assert.EqualError(t, err, c.errMsg, "case %d: error should match", n)
		} else {
			require.NoError(t, err, "case %d: should have no errors", n)
			assert.Equal(t, c.authSource, updatedAuthSource, "case %d: wrong authSource", n)
		}
	}
}
