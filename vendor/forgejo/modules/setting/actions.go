// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forgejo.org/modules/jwtx"
)

// Actions settings
var (
	Actions = struct {
		Enabled                      bool
		LogStorage                   *Storage          // how the created logs should be stored
		LogRetentionDays             int64             `ini:"LOG_RETENTION_DAYS"`
		LogCompression               logCompression    `ini:"LOG_COMPRESSION"`
		ArtifactStorage              *Storage          // how the created artifacts should be stored
		ArtifactRetentionDays        int64             `ini:"ARTIFACT_RETENTION_DAYS"`
		DefaultActionsURL            defaultActionsURL `ini:"DEFAULT_ACTIONS_URL"`
		ZombieTaskTimeout            time.Duration     `ini:"ZOMBIE_TASK_TIMEOUT"`
		EndlessTaskTimeout           time.Duration     `ini:"ENDLESS_TASK_TIMEOUT"`
		AbandonedJobTimeout          time.Duration     `ini:"ABANDONED_JOB_TIMEOUT"`
		SkipWorkflowStrings          []string          `ini:"SKIP_WORKFLOW_STRINGS"`
		LimitDispatchInputs          int64             `ini:"LIMIT_DISPATCH_INPUTS"`
		ConcurrencyGroupQueueEnabled bool              `ini:"CONCURRENCY_GROUP_QUEUE_ENABLED"`
		IDTokenSigningAlgorithm      idTokenAlgorithm  `ini:"ID_TOKEN_SIGNING_ALGORITHM"`
		IDTokenSigningPrivateKeyFile string            `ini:"ID_TOKEN_SIGNING_PRIVATE_KEY_FILE"`
		IDTokenExpirationTime        int64             `ini:"ID_TOKEN_EXPIRATION_TIME"`
	}{
		Enabled:                      true,
		DefaultActionsURL:            defaultActionsURLForgejo,
		SkipWorkflowStrings:          []string{"[skip ci]", "[ci skip]", "[no ci]", "[skip actions]", "[actions skip]"},
		LimitDispatchInputs:          100,
		ConcurrencyGroupQueueEnabled: true,
		IDTokenSigningAlgorithm:      "RS256",
		IDTokenSigningPrivateKeyFile: "actions_id_token/private.pem",
		IDTokenExpirationTime:        3600,
	}
)

type defaultActionsURL string

func (url defaultActionsURL) URL() string {
	switch url {
	case defaultActionsURLGitHub:
		return "https://github.com"
	case defaultActionsURLSelf:
		return strings.TrimSuffix(AppURL, "/")
	default:
		return string(url)
	}
}

const (
	defaultActionsURLForgejo = "https://data.forgejo.org"
	defaultActionsURLGitHub  = "github" // https://github.com
	defaultActionsURLSelf    = "self"   // the root URL of the self-hosted instance
)

type logCompression string

func (c logCompression) IsValid() bool {
	return c.IsNone() || c.IsZstd()
}

func (c logCompression) IsNone() bool {
	return strings.ToLower(string(c)) == "none"
}

func (c logCompression) IsZstd() bool {
	return c == "" || strings.ToLower(string(c)) == "zstd"
}

type idTokenAlgorithm string

func (c idTokenAlgorithm) IsValid() bool {
	// Empty string implies RS256
	return jwtx.IsValidAsymmetricAlgorithm(string(c)) || string(c) == ""
}

func loadActionsFrom(rootCfg ConfigProvider) error {
	sec := rootCfg.Section("actions")
	err := sec.MapTo(&Actions)
	if err != nil {
		return fmt.Errorf("failed to map Actions settings: %v", err)
	}

	// don't support to read configuration from [actions]
	Actions.LogStorage, err = getStorage(rootCfg, "actions_log", "", nil)
	if err != nil {
		return err
	}
	// default to 1 year
	if Actions.LogRetentionDays <= 0 {
		Actions.LogRetentionDays = 365
	}

	actionsSec, _ := rootCfg.GetSection("actions.artifacts")

	Actions.ArtifactStorage, err = getStorage(rootCfg, "actions_artifacts", "", actionsSec)
	if err != nil {
		return err
	}

	// default to 90 days in Github Actions
	if Actions.ArtifactRetentionDays <= 0 {
		Actions.ArtifactRetentionDays = 90
	}

	Actions.ZombieTaskTimeout = sec.Key("ZOMBIE_TASK_TIMEOUT").MustDuration(10 * time.Minute)
	Actions.EndlessTaskTimeout = sec.Key("ENDLESS_TASK_TIMEOUT").MustDuration(3 * time.Hour)
	Actions.AbandonedJobTimeout = sec.Key("ABANDONED_JOB_TIMEOUT").MustDuration(24 * time.Hour)

	if !Actions.LogCompression.IsValid() {
		return fmt.Errorf("invalid [actions] LOG_COMPRESSION: %q", Actions.LogCompression)
	}

	if !Actions.IDTokenSigningAlgorithm.IsValid() {
		return fmt.Errorf("invalid [actions] ID_TOKEN_SIGNING_ALGORITHM: %q", Actions.IDTokenSigningAlgorithm)
	}

	if !filepath.IsAbs(Actions.IDTokenSigningPrivateKeyFile) {
		Actions.IDTokenSigningPrivateKeyFile = filepath.Join(AppDataPath, Actions.IDTokenSigningPrivateKeyFile)
	}

	return nil
}
