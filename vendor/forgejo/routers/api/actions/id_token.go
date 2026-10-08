// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"forgejo.org/models/actions"
	"forgejo.org/modules/json"
	"forgejo.org/modules/log"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/timeutil"
	"forgejo.org/modules/web"
	web_types "forgejo.org/modules/web/types"
	actions_service "forgejo.org/services/actions"
	"forgejo.org/services/context"

	"github.com/golang-jwt/jwt/v5"
)

const idTokenRouteBase = "/_apis/pipelines/workflows/{run_id}/idtoken"

type idTokenContextKeyType struct{}

var idTokenContextKey = idTokenContextKeyType{}

type IDTokenContext struct {
	*context.Base

	Audience                 string
	AuthorizationTokenClaims *actions_service.AuthorizationTokenClaims
	IDTokenCustomClaims      *actions_service.IDTokenCustomClaims
}

func init() {
	web.RegisterResponseStatusProvider[*IDTokenContext](func(req *http.Request) web_types.ResponseStatusProvider {
		return req.Context().Value(idTokenContextKey).(*IDTokenContext)
	})
}

func IDTokenContexter() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			base, baseCleanUp := context.NewBaseContext(resp, req)
			defer baseCleanUp()

			ctx := &IDTokenContext{Base: base}
			ctx.AppendContextValue(idTokenContextKey, ctx)

			// action task call server api with Bearer ACTIONS_ID_TOKEN_REQUEST_TOKEN
			// we should verify the ACTIONS_ID_TOKEN_REQUEST_TOKEN
			authHeader := req.Header.Get("Authorization")
			if len(authHeader) == 0 || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				ctx.Error(http.StatusUnauthorized, "Bad authorization header")
				return
			}

			// Require using new act_runner that uses jwt to authenticate
			authorizationTokenClaims, err := actions_service.ParseAuthorizationTokenClaims(req)
			if err != nil {
				log.Error("Error runner api parsing authorization token: %v", err)
				ctx.Error(http.StatusInternalServerError, "Error runner api parsing authorization token")
				return
			}

			customClaims := &actions_service.IDTokenCustomClaims{}
			err = json.Unmarshal([]byte(authorizationTokenClaims.OIDCExtra), customClaims)
			if err != nil {
				log.Error("Error runner api parsing custom claims: %v", err)
				ctx.Error(http.StatusInternalServerError, "Error runner api parsing custom claims")
				return
			}

			task, err := actions.GetTaskByID(req.Context(), authorizationTokenClaims.TaskID)
			if err != nil {
				log.Error("Error runner api getting task by ID: %v", err)
				ctx.Error(http.StatusInternalServerError, "Error runner api getting task by ID")
				return
			}
			if task.Status != actions.StatusRunning {
				log.Error("Error runner api getting task: task is not running")
				ctx.Error(http.StatusInternalServerError, "Error runner api getting task: task is not running")
				return
			}
			err = task.LoadAttributes(req.Context())
			if err != nil {
				log.Error("Error runner api getting task attributes: %v", err)
				ctx.Error(http.StatusInternalServerError, "Error runner api getting task attributes")
				return
			}

			runID := ctx.ParamsInt64("run_id")
			if task.Job.RunID != runID {
				log.Error("Error runID not match" + fmt.Sprint(task.Job.RunID) + " " + fmt.Sprint(runID))
				ctx.Error(http.StatusBadRequest, "run-id does not match")
				return
			}

			generateIDTokenScp := fmt.Sprintf("generate_id_token:%d:%d", task.Job.RunID, task.Job.ID)
			scp := strings.Split(authorizationTokenClaims.Scp, " ")
			if !slices.Contains(scp, generateIDTokenScp) {
				ctx.Error(http.StatusForbidden, "missing scp generate_id_token")
				return
			}

			audience := req.URL.Query().Get("audience")
			if audience == "" {
				// Default to organization that owns the repo if no audience is provided
				audience = setting.AppURL + customClaims.RepositoryOwner
			}

			ctx.AuthorizationTokenClaims = authorizationTokenClaims
			ctx.IDTokenCustomClaims = customClaims
			ctx.Audience = audience
			next.ServeHTTP(ctx.Resp, ctx.Req)
		})
	}
}

func generateIDToken(ctx *IDTokenContext) {
	expirationDate := timeutil.TimeStampNow().Add(setting.Actions.IDTokenExpirationTime)

	var claims jwt.MapClaims
	inrec, _ := json.Marshal(ctx.IDTokenCustomClaims)
	err := json.Unmarshal(inrec, &claims)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "Error generating token")
		return
	}
	now := time.Now()

	claims["sub"] = ctx.AuthorizationTokenClaims.OIDCSub
	claims["aud"] = ctx.Audience
	claims["exp"] = jwt.NewNumericDate(expirationDate.AsTime())
	claims["iat"] = jwt.NewNumericDate(now)
	claims["nbf"] = jwt.NewNumericDate(now)
	claims["iss"] = strings.TrimSuffix(setting.AppURL, "/") + "/api/actions"

	signedToken, err := jwtSigningKey.JWT(claims)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "Error signing token")
		return
	}

	resp := IDTokenResponse{
		Value: signedToken,
	}

	ctx.JSON(http.StatusOK, resp)
}

type IDTokenResponse struct {
	Value string `json:"value"`
}
