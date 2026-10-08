// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/activitypub"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/routers"
	"forgejo.org/services/contexttest"
	"forgejo.org/services/federation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Flow of this test is documented at: https://codeberg.org/forgejo-contrib/federation/src/branch/main/doc/user-activity-following.md
func TestActivityPubPersonInboxFollow(t *testing.T) {
	defer test.MockVariableValue(&setting.Federation.Enabled, true)()
	defer test.MockVariableValue(&setting.Federation.SignatureEnforced, false)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

	federation.Init()

	mock := test.NewFederationServerMock()
	federatedSrv := mock.DistantServer(t)
	defer federatedSrv.Close()

	onApplicationRun(t, func(t *testing.T, localUrl *url.URL) {
		defer test.MockVariableValue(&setting.AppURL, localUrl.String())()

		distantURL := federatedSrv.URL
		distantUser15URL := fmt.Sprintf("%s/api/v1/activitypub/user-id/15", distantURL)

		localUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
		localUser2URL := localUrl.JoinPath("/api/v1/activitypub/user-id/2").String()
		localUser2Inbox := localUrl.JoinPath("/api/v1/activitypub/user-id/2/inbox").String()

		ctx, _ := contexttest.MockAPIContext(t, localUser2Inbox)

		// distant follows local
		followActivity := fmt.Appendf(nil,
			`{"type":"Follow",`+
				`"actor":"%s",`+
				`"object":"%s"}`,
			distantUser15URL,
			localUser2URL,
		)
		cf, err := activitypub.NewClientFactoryWithTimeout(60 * time.Second)
		require.NoError(t, err)
		c, err := cf.WithKeysDirect(ctx, mock.ApActor.PrivKey,
			mock.ApActor.KeyID(federatedSrv.URL))
		require.NoError(t, err)
		resp, err := c.Post(followActivity, localUser2Inbox)
		require.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		// local follow exists
		distantFederatedUser := unittest.AssertExistsAndLoadBean(t, &user_model.FederatedUser{ExternalID: "15"})
		unittest.AssertExistsAndLoadBean(t,
			&user_model.FederatedUserFollower{
				FollowedUserID:  localUser.ID,
				FollowingUserID: distantFederatedUser.UserID,
			},
		)

		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: distantFederatedUser.UserID})
		assert.Equal(t, user_model.UserTypeActivityPubUser, user.Type)
		assert.True(t, user.ProhibitLogin)
		assert.Empty(t, user.Passwd)
		assert.Empty(t, user.PasswdHashAlgo)
		assert.Empty(t, user.Salt)

		// distant is informed about accepting follow
		assert.Contains(t, mock.LastPost, "\"type\":\"Accept\"")

		// distant undoes follow
		undoFollowActivity := fmt.Appendf(nil,
			`{"type":"Undo",`+
				`"actor":"%s",`+
				`"object":{"type":"Follow",`+
				`"actor":"%s",`+
				`"object":"%s"}}`,
			distantUser15URL,
			distantUser15URL,
			localUser2URL,
		)
		c, err = cf.WithKeysDirect(ctx, mock.ApActor.PrivKey,
			mock.ApActor.KeyID(federatedSrv.URL))
		require.NoError(t, err)
		resp, err = c.Post(undoFollowActivity, localUser2Inbox)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// local follow removed
		unittest.AssertNotExistsBean(t,
			&user_model.FederatedUserFollower{
				FollowedUserID:  localUser.ID,
				FollowingUserID: distantFederatedUser.UserID,
			},
		)
	})
}
