// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	actions_model "forgejo.org/models/actions"
	"forgejo.org/modules/json"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"

	"github.com/golang-jwt/jwt/v5"
)

type AuthorizationTokenClaims struct {
	jwt.RegisteredClaims
	Scp       string `json:"scp"`
	TaskID    int64
	RunID     int64
	JobID     int64
	Ac        string `json:"ac"`
	OIDCExtra string `json:"oidc_extra,omitempty"`
	OIDCSub   string `json:"oidc_sub,omitempty"`
}

type IDTokenCustomClaims struct {
	Actor           string `json:"actor"`
	BaseRef         string `json:"base_ref"`
	EventName       string `json:"event_name"`
	HeadRef         string `json:"head_ref"`
	Ref             string `json:"ref"`
	RefProtected    string `json:"ref_protected"`
	RefType         string `json:"ref_type"`
	Repository      string `json:"repository"`
	RepositoryOwner string `json:"repository_owner"`
	RunAttempt      string `json:"run_attempt"`
	RunID           string `json:"run_id"`
	RunNumber       string `json:"run_number"`
	Sha             string `json:"sha"`
	Workflow        string `json:"workflow"`
	WorkflowRef     string `json:"workflow_ref"`
}

type actionsCacheScope struct {
	Scope      string
	Permission actionsCachePermission
}

type actionsCachePermission int

const (
	actionsCachePermissionRead = 1 << iota
	actionsCachePermissionWrite
)

func CreateAuthorizationToken(task *actions_model.ActionTask, gitGtx map[string]any, enableOpenIDConnect bool) (string, error) {
	now := time.Now()
	taskID := task.ID
	runID := task.Job.RunID
	jobID := task.Job.ID

	ac, err := json.Marshal(&[]actionsCacheScope{
		{
			Scope:      "",
			Permission: actionsCachePermissionWrite,
		},
	})
	if err != nil {
		return "", err
	}

	runIDJobID := fmt.Sprintf("%d:%d", runID, jobID)
	claims := AuthorizationTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
		Scp:    fmt.Sprintf("Actions.Results:%s", runIDJobID),
		Ac:     string(ac),
		TaskID: taskID,
		RunID:  runID,
		JobID:  jobID,
	}

	// Only populate OIDC information if the task has OIDC enabled.
	if enableOpenIDConnect {
		oidcExtra, err := generateOIDCExtra(gitGtx)
		if err != nil {
			return "", err
		}

		claims.OIDCExtra = oidcExtra
		claims.OIDCSub = generateOIDCSub(gitGtx)
		claims.Scp = fmt.Sprintf("%s generate_id_token:%s", claims.Scp, runIDJobID)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(setting.GetGeneralTokenSigningSecret())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateOIDCExtra(gitCtx map[string]any) (string, error) {
	ctxVal := func(key string) string {
		val, ok := gitCtx[key]
		if !ok {
			return ""
		}
		return fmt.Sprint(val)
	}

	claims := IDTokenCustomClaims{
		Actor:           ctxVal("actor"),
		BaseRef:         ctxVal("base_ref"),
		EventName:       ctxVal("event_name"),
		HeadRef:         ctxVal("head_ref"),
		Ref:             ctxVal("ref"),
		RefProtected:    ctxVal("ref_protected"),
		RefType:         ctxVal("ref_type"),
		Repository:      ctxVal("repository"),
		RepositoryOwner: ctxVal("repository_owner"),
		RunAttempt:      ctxVal("run_attempt"),
		RunID:           ctxVal("run_id"),
		RunNumber:       ctxVal("run_number"),
		Sha:             ctxVal("sha"),
		Workflow:        ctxVal("workflow"),
		WorkflowRef:     ctxVal("workflow_ref"),
	}

	ret, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	return string(ret), nil
}

func generateOIDCSub(gitCtx map[string]any) string {
	switch gitCtx["event_name"] {
	case "pull_request":
		return fmt.Sprintf("repo:%s:pull_request", gitCtx["repository"])
	default:
		return fmt.Sprintf("repo:%s:ref:%s", gitCtx["repository"], gitCtx["ref"])
	}
}

func ParseAuthorizationToken(req *http.Request) (int64, error) {
	token, err := parseTokenFromHeader(req)
	if err != nil {
		return 0, err
	}

	if token == "" {
		return 0, nil
	}

	return TokenToTaskID(token)
}

// TokenToTaskID returns the TaskID associated with the provided JWT token
func TokenToTaskID(token string) (int64, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &AuthorizationTokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return setting.GetGeneralTokenSigningSecret(), nil
	})
	if err != nil {
		return 0, err
	}

	c, ok := parsedToken.Claims.(*AuthorizationTokenClaims)
	if !parsedToken.Valid || !ok {
		return 0, errors.New("invalid token claim")
	}

	return c.TaskID, nil
}

func ParseAuthorizationTokenClaims(req *http.Request) (*AuthorizationTokenClaims, error) {
	token, err := parseTokenFromHeader(req)
	if err != nil {
		return nil, err
	}

	claims, err := decodeTokenClaims(token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func parseTokenFromHeader(req *http.Request) (string, error) {
	h := req.Header.Get("Authorization")
	if h == "" {
		return "", nil
	}

	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		log.Error("split token failed: %s", h)
		return "", errors.New("split token failed")
	}

	return parts[1], nil
}

func decodeTokenClaims(token string) (*AuthorizationTokenClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &AuthorizationTokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return setting.GetGeneralTokenSigningSecret(), nil
	})
	if err != nil {
		return nil, err
	}

	c, ok := parsedToken.Claims.(*AuthorizationTokenClaims)
	if !parsedToken.Valid || !ok {
		return nil, errors.New("invalid token claim")
	}

	return c, nil
}
