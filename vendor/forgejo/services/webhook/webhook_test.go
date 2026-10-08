// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package webhook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	webhook_model "forgejo.org/models/webhook"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/test"
	webhook_module "forgejo.org/modules/webhook"
	"forgejo.org/services/convert"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func activateWebhook(t *testing.T, hookID int64) {
	t.Helper()
	updated, err := db.GetEngine(db.DefaultContext).ID(hookID).Cols("is_active").Update(webhook_model.Webhook{IsActive: true})
	assert.Equal(t, int64(1), updated)
	require.NoError(t, err)
}

func TestPrepareWebhooks(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	activateWebhook(t, 1)

	hookTasks := []*webhook_model.HookTask{
		{HookID: 1, EventType: webhook_module.HookEventPush},
	}
	for _, hookTask := range hookTasks {
		unittest.AssertNotExistsBean(t, hookTask)
	}
	require.NoError(t, PrepareWebhooks(db.DefaultContext, EventSource{Repository: repo}, webhook_module.HookEventPush, &api.PushPayload{Commits: []*api.PayloadCommit{{}}}))
	for _, hookTask := range hookTasks {
		unittest.AssertExistsAndLoadBean(t, hookTask)
	}
}

func eventType(p api.Payloader) webhook_module.HookEventType {
	switch p.(type) {
	case *api.CreatePayload:
		return webhook_module.HookEventCreate
	case *api.DeletePayload:
		return webhook_module.HookEventDelete
	case *api.PushPayload:
		return webhook_module.HookEventPush
	}
	panic(fmt.Sprintf("no event type for payload %T", p))
}

func TestPrepareWebhooksBranchFilterMatch(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// branch_filter: {master,feature*}
	w := unittest.AssertExistsAndLoadBean(t, &webhook_model.Webhook{ID: 4})
	activateWebhook(t, w.ID)

	for _, p := range []api.Payloader{
		&api.PushPayload{Ref: "refs/heads/feature/7791"},
		&api.CreatePayload{Ref: "refs/heads/feature/7791"}, // branch creation
		&api.DeletePayload{Ref: "refs/heads/feature/7791"}, // branch deletion
	} {
		t.Run(fmt.Sprintf("%T", p), func(t *testing.T) {
			db.DeleteBeans(db.DefaultContext, webhook_model.HookTask{HookID: w.ID})
			typ := eventType(p)
			require.NoError(t, PrepareWebhook(db.DefaultContext, w, typ, p))
			unittest.AssertExistsAndLoadBean(t, &webhook_model.HookTask{
				HookID:    w.ID,
				EventType: typ,
			})
		})
	}
}

func TestPrepareWebhooksBranchFilterNoMatch(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// branch_filter: {master,feature*}
	w := unittest.AssertExistsAndLoadBean(t, &webhook_model.Webhook{ID: 4})
	activateWebhook(t, w.ID)

	for _, p := range []api.Payloader{
		&api.PushPayload{Ref: "refs/heads/fix_weird_bug"},
		&api.CreatePayload{Ref: "refs/heads/fix_weird_bug"}, // branch creation
		&api.DeletePayload{Ref: "refs/heads/fix_weird_bug"}, // branch deletion
	} {
		t.Run(fmt.Sprintf("%T", p), func(t *testing.T) {
			db.DeleteBeans(db.DefaultContext, webhook_model.HookTask{HookID: w.ID})
			require.NoError(t, PrepareWebhook(db.DefaultContext, w, eventType(p), p))
			unittest.AssertNotExistsBean(t, &webhook_model.HookTask{HookID: w.ID})
		})
	}
}

func TestWebhookUserMail(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	defer test.MockVariableValue(&setting.Service.NoReplyAddress, "no-reply.com")()

	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
	assert.Equal(t, user.GetPlaceholderEmail(), convert.ToUser(db.DefaultContext, user, nil).Email)
	assert.Equal(t, user.Email, convert.ToUser(db.DefaultContext, user, user).Email)
}

func TestDeliverTestPayloadWithoutPushEvent(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	done := make(chan struct{}, 1)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/webhook", r.URL.Path)
		w.WriteHeader(200)
		done <- struct{}{}
	}))
	t.Cleanup(s.Close)

	hookEvent := webhook_module.HookEvent{ChooseEvents: true, HookEvents: webhook_module.HookEvents{Release: true}}
	hook := &webhook_model.Webhook{
		RepoID:      3,
		URL:         s.URL + "/webhook",
		ContentType: webhook_model.ContentTypeJSON,
		IsActive:    true,
		Type:        webhook_module.GITEA,
		HookEvent:   &hookEvent,
	}
	require.NoError(t, webhook_model.CreateWebhook(t.Context(), hook, ""))

	// if we deliver this webhook for a push event, nothing happens because the webhook isn't configured to run on those events
	require.NoError(t, PrepareWebhook(db.DefaultContext, hook, webhook_module.HookEventPush, &api.ReleasePayload{}))
	unittest.AssertNotExistsBean(t, &webhook_model.HookTask{
		HookID: hook.ID,
	})

	// but if we deliver it as a testing payload, the check on event types is bypassed
	// (so that webhooks can be tested regardless of the event types they are enabled for)
	// See https://codeberg.org/forgejo/forgejo/issues/7934
	require.NoError(t, PrepareTestWebhook(db.DefaultContext, hook, &api.ReleasePayload{}))

	unittest.AssertExistsAndLoadBean(t, &webhook_model.HookTask{
		HookID:    hook.ID,
		EventType: webhook_module.HookEventPush,
	})
}
