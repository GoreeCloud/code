// Copyright 2019 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"testing"

	"forgejo.org/models/auth"
	"forgejo.org/modules/test"
	"forgejo.org/services/auth/source/pam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestPamService(t *testing.T) {
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
				"pam-test",
				"--name", "Pam Service",
				"--service-name", "myservice",
			},
			source: &auth.Source{
				Type:     auth.PAM,
				Name:     "Pam Service",
				IsActive: true,
				Cfg: &pam.Source{
					ServiceName:    "myservice",
					EmailDomain:    "",
					SkipLocalTwoFA: true,
				},
			},
		},
		// case 1
		{
			args: []string{
				"pam-test",
				"--name", "Pam Service",
				"--service-name", "myservice",
				"--email-domain", "testdomain.org",
				"--skip-local-2fa",
			},
			source: &auth.Source{
				Type:     auth.PAM,
				Name:     "Pam Service",
				IsActive: true,
				Cfg: &pam.Source{
					ServiceName:    "myservice",
					EmailDomain:    "testdomain.org",
					SkipLocalTwoFA: true,
				},
			},
		},
		// case 2
		{
			args: []string{
				"pam-test",
				"--service-name", "myservice",
				"--email-domain", "testdomain.org",
				"--skip-local-2fa", "false",
				"--active", "true",
			},
			errMsg: "name must be set",
		},
		// case 3
		{
			args: []string{
				"pam-test",
				"--name", "Pam Service",
				"--email-domain", "testdomain.org",
				"--skip-local-2fa", "false",
				"--active", "true",
			},
			errMsg: "service-name must be set",
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
		app.Flags = microcmdAuthAddPAM().Flags
		app.Action = service.addPAM

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

func TestUpdatePAM(t *testing.T) {
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
				"pam-test",
				"--id", "23",
				"--name", "PAM Service",
				"--service-name", "myservice",
			},
			id: 23,
			existingAuthSource: &auth.Source{
				Type:     auth.PAM,
				IsActive: true,
				Cfg:      &pam.Source{},
			},
			authSource: &auth.Source{
				Type:     auth.PAM,
				Name:     "PAM Service",
				IsActive: true,
				Cfg: &pam.Source{
					ServiceName: "myservice",
				},
			},
		},
		// case 1
		{
			args: []string{
				"pam-test",
				"--id", "1",
			},
			authSource: &auth.Source{
				Type: auth.PAM,
				Cfg:  &pam.Source{},
			},
		},
		// case 2
		{
			args: []string{
				"pam-test",
				"--id", "1",
				"--name", "pam service",
			},
			authSource: &auth.Source{
				Type: auth.PAM,
				Name: "pam service",
				Cfg:  &pam.Source{},
			},
		},
		// case 3
		{
			args: []string{
				"pam-test",
				"--id", "1",
				"--active=false",
			},
			existingAuthSource: &auth.Source{
				Type:     auth.PAM,
				IsActive: true,
				Cfg:      &pam.Source{},
			},
			authSource: &auth.Source{
				Type:     auth.PAM,
				IsActive: false,
				Cfg:      &pam.Source{},
			},
		},
		// case 4
		{
			args: []string{
				"pam-test",
				"--id", "1",
				"--service-name", "myservice",
			},
			authSource: &auth.Source{
				Type: auth.PAM,
				Cfg: &pam.Source{
					ServiceName: "myservice",
				},
			},
		},
		// case 5
		{
			args: []string{
				"pam-test",
				"--id", "1",
				"--skip-local-2fa=false",
			},
			authSource: &auth.Source{
				Type: auth.PAM,
				Cfg: &pam.Source{
					SkipLocalTwoFA: false,
				},
			},
		},
		// case 6
		{
			args: []string{
				"pam-test",
				"--id", "1",
				"--email-domain", "testdomain.org",
			},
			authSource: &auth.Source{
				Type: auth.PAM,
				Cfg: &pam.Source{
					EmailDomain: "testdomain.org",
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
					Type: auth.PAM,
					Cfg:  &pam.Source{},
				}, nil
			},
		}

		// Create a copy of command to test
		app := cli.Command{}
		app.Flags = microcmdAuthUpdatePAM().Flags
		app.Action = service.updatePAM

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
