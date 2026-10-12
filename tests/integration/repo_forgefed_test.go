// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"net/url"
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/forgefed"
	"forgejo.org/models/unittest"
	"forgejo.org/models/user"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/routers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForgefedRepositoryCreateHostValid(t *testing.T) {
	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		// Arrange
		ctx := t.Context()

		// Act
		err := forgefed.CreateFederationHost(ctx, &forgefed.FederationHost{
			HostFqdn:   "forgejo.example.com",
			HostPort:   80,
			HostSchema: "http",
			NodeInfo: forgefed.NodeInfo{
				SoftwareName: "forgejo",
			},
		})

		// Assert
		require.NoError(t, err)
		unittest.AssertExistsAndLoadBean(t, &forgefed.FederationHost{HostFqdn: "forgejo.example.com"})
	})
}

func TestForgefedRepositoryCreateHostInvalid(t *testing.T) {
	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		// Arrange
		ctx := t.Context()

		// Act
		err := forgefed.CreateFederationHost(ctx, &forgefed.FederationHost{
			// invalid
		})

		// Assert
		require.Error(t, err)
		unittest.AssertNotExistsBean(t, &forgefed.FederationHost{HostFqdn: "forgejo.example.com"})
	})
}

func TestForgefedRepositoryCreateUserValid(t *testing.T) {
	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		// Arrange
		ctx := t.Context()

		err := forgefed.CreateFederationHost(ctx, &forgefed.FederationHost{
			HostFqdn:   "forgejo.example.com",
			HostPort:   80,
			HostSchema: "http",
			NodeInfo: forgefed.NodeInfo{
				SoftwareName: "forgejo",
			},
		})
		require.NoError(t, err)

		// Act
		err = user.CreateFederatedUser(ctx, &user.User{
			Name:  "@bob@forgejo.example.com",
			Email: "bob@forgejo.example.com",
		}, &user.FederatedUser{
			ExternalID:       "1",
			FederationHostID: 1,
			InboxPath:        "/inbox",
		})

		// Assert
		require.NoError(t, err)
		localUser := unittest.AssertExistsAndLoadBean(t, &user.User{Name: "@bob@forgejo.example.com", Email: "bob@forgejo.example.com"})
		unittest.AssertExistsAndLoadBean(t, &user.FederatedUser{UserID: localUser.ID, FederationHostID: 1})
	})
}

func TestForgefedRepositoryCreateUserInvalid(t *testing.T) {
	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		// Arrange
		ctx := t.Context()

		err := forgefed.CreateFederationHost(ctx, &forgefed.FederationHost{
			HostFqdn:   "forgejo.example.com",
			HostPort:   80,
			HostSchema: "http",
			NodeInfo: forgefed.NodeInfo{
				SoftwareName: "forgejo",
			},
		})
		require.NoError(t, err)

		// Act
		err = user.CreateFederatedUser(ctx, &user.User{
			Name:  "@bob@forgejo.example.com",
			Email: "bob@forgejo.example.com",
		}, &user.FederatedUser{
			// invalid
		})

		// Assert
		require.Error(t, err)
		unittest.AssertNotExistsBean(t, &user.User{Email: "bob@forgejo.example.com"})
		unittest.AssertNotExistsBean(t, &user.FederatedUser{FederationHostID: 1})
	})
}

func TestForgefedRepositoryFindHostsAndUsers(t *testing.T) {
	defer test.MockVariableValue(&setting.Federation.Enabled, true)()
	defer test.MockVariableValue(&setting.Federation.SignatureEnforced, false)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		// Arrange
		ctx := t.Context()

		err := forgefed.CreateFederationHost(ctx, &forgefed.FederationHost{
			HostFqdn:   "bob.example.com",
			HostPort:   80,
			HostSchema: "http",
			NodeInfo: forgefed.NodeInfo{
				SoftwareName: "forgejo",
			},
		})
		require.NoError(t, err)

		err = forgefed.CreateFederationHost(ctx, &forgefed.FederationHost{
			HostFqdn:   "alice.example.com",
			HostPort:   443,
			HostSchema: "https",
			NodeInfo: forgefed.NodeInfo{
				SoftwareName: "gitea",
			},
		})
		require.NoError(t, err)

		err = user.CreateFederatedUser(ctx, &user.User{
			Name:  "@bob@bob.example.com",
			Email: "bob@bob.example.com",
		}, &user.FederatedUser{
			ExternalID:       "1",
			FederationHostID: 1,
			InboxPath:        "/inbox",
		})
		require.NoError(t, err)

		err = user.CreateFederatedUser(ctx, &user.User{
			Name:  "@alice@alice.example.com",
			Email: "alice@alice.example.com",
		}, &user.FederatedUser{
			ExternalID:       "1",
			FederationHostID: 2,
			InboxPath:        "/inbox",
		})
		require.NoError(t, err)

		err = user.CreateFederatedUser(ctx, &user.User{
			Name:  "@eve@alice.example.com",
			Email: "eve@alice.example.com",
		}, &user.FederatedUser{
			ExternalID:       "2",
			FederationHostID: 2,
			InboxPath:        "/inbox",
		})
		require.NoError(t, err)

		// Act & Assert
		hosts, err := forgefed.FindFederationHosts(ctx, db.ListOptions{PageSize: 100, Page: 1})
		require.NoError(t, err)
		assert.Len(t, hosts, 2)
		hostFqdns := []string{hosts[0].HostFqdn, hosts[1].HostFqdn}
		assert.Contains(t, hostFqdns, "bob.example.com")
		assert.Contains(t, hostFqdns, "alice.example.com")

		count, err := forgefed.CountFederationHosts(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		users, err := user.FindFederatedUsers(ctx, db.ListOptions{PageSize: 100, Page: 1})
		require.NoError(t, err)
		assert.Len(t, users, 3) // Bob, Alice and Eve

		count, err = user.CountFederatedUsers(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		users, err = user.FindFederatedUsersByHostID(ctx, 1, db.ListOptions{PageSize: 100, Page: 1})
		require.NoError(t, err)
		assert.Len(t, users, 1) // Only Bob belongs to the host with ID 1
		assert.Equal(t, int64(1), users[0].FederationHostID)

		count, err = user.CountFederatedUsersByHostID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		users, err = user.FindFederatedUsersByHostID(ctx, 2, db.ListOptions{PageSize: 100, Page: 1})
		require.NoError(t, err)
		assert.Len(t, users, 2) // Alice and Eve belong to the host with ID 2
		assert.Equal(t, int64(2), users[0].FederationHostID)
		assert.Equal(t, int64(2), users[1].FederationHostID)

		count, err = user.CountFederatedUsersByHostID(ctx, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}
