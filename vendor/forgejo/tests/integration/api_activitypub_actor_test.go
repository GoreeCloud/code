// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"forgejo.org/modules/forgefed"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/routers"
	"forgejo.org/services/contexttest"
	"forgejo.org/services/federation"
	"forgejo.org/tests"

	ap "github.com/go-ap/activitypub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestActivityPubActor(t *testing.T) {
	defer test.MockVariableValue(&setting.Federation.Enabled, true)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()
	defer tests.PrepareTestEnv(t)()

	req := NewRequest(t, "GET", "/api/v1/activitypub/actor")
	resp := MakeRequest(t, req, http.StatusOK)
	assert.Contains(t, resp.Body.String(), "@context")

	var actor ap.Actor
	err := actor.UnmarshalJSON(resp.Body.Bytes())
	require.NoError(t, err)

	assert.Equal(t, ap.ApplicationType, actor.Type)
	assert.Equal(t, "ghost", actor.PreferredUsername.String())
	keyID := actor.GetID().String()
	assert.Regexp(t, "activitypub/actor$", keyID)
	assert.Regexp(t, "activitypub/actor/inbox$", actor.Inbox.GetID().String())
	assert.Regexp(t, "activitypub/actor/outbox$", actor.Outbox.GetID().String())

	pubKey := actor.PublicKey
	assert.NotNil(t, pubKey)
	publicKeyID := keyID + "#main-key"
	assert.Equal(t, pubKey.ID.String(), publicKeyID)

	pubKeyPem := pubKey.PublicKeyPem
	assert.NotNil(t, pubKeyPem)
	assert.Regexp(t, "^-----BEGIN PUBLIC KEY-----", pubKeyPem)

	t.Run("ActorOutboxEmpty", func(t *testing.T) {
		req := NewRequest(t, "GET", actor.Outbox.GetID().String())
		resp := MakeRequest(t, req, http.StatusOK)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		jsonResp, err := fastjson.ParseBytes(body)
		require.NoError(t, err)

		outbox := ap.JSONUnmarshalToItem(jsonResp)
		require.NoError(t, err)

		assert.Equal(t, ap.OrderedCollectionType, outbox.GetType())
		outboxCollection, ok := outbox.(*ap.OrderedCollection)
		require.True(t, ok)

		assert.Equal(t, uint(0), outboxCollection.TotalItems)
		assert.Nil(t, outboxCollection.First)
		assert.Nil(t, outboxCollection.Last)
	})
}

func TestActorNewFromKeyId(t *testing.T) {
	defer test.MockVariableValue(&setting.Federation.Enabled, true)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		ctx, _ := contexttest.MockAPIContext(t, "/api/v1/activitypub/actor")
		sut, err := federation.NewActorIDFromKeyID(ctx.Base, fmt.Sprintf("%sapi/v1/activitypub/actor#main-key", u))
		require.NoError(t, err)

		port, err := strconv.ParseUint(u.Port(), 10, 16)
		require.NoError(t, err)

		assert.Equal(t, forgefed.ActorID{
			ID:                 "actor",
			HostSchema:         "http",
			Path:               "api/v1/activitypub",
			Host:               setting.Domain,
			HostPort:           uint16(port),
			UnvalidatedInput:   fmt.Sprintf("http://%s:%d/api/v1/activitypub/actor", setting.Domain, port),
			IsPortSupplemented: false,
		}, sut)
	})
}
