// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parsePostgreSQLHostPort(t *testing.T) {
	tests := map[string]struct {
		HostPort string
		Host     string
		Port     string
	}{
		"host-port": {
			HostPort: "127.0.0.1:1234",
			Host:     "127.0.0.1",
			Port:     "1234",
		},
		"no-port": {
			HostPort: "127.0.0.1",
			Host:     "127.0.0.1",
			Port:     "5432",
		},
		"ipv6-port": {
			HostPort: "[::1]:1234",
			Host:     "::1",
			Port:     "1234",
		},
		"ipv6-no-port": {
			HostPort: "[::1]",
			Host:     "::1",
			Port:     "5432",
		},
		"unix-socket": {
			HostPort: "/tmp/pg.sock:1234",
			Host:     "/tmp/pg.sock",
			Port:     "1234",
		},
		"unix-socket-no-port": {
			HostPort: "/tmp/pg.sock",
			Host:     "/tmp/pg.sock",
			Port:     "5432",
		},
	}
	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			t.Log(test.HostPort)
			host, port := parsePostgreSQLHostPort(test.HostPort)
			assert.Equal(t, test.Host, host)
			assert.Equal(t, test.Port, port)
		})
	}
}

func Test_getPostgreSQLConnectionString(t *testing.T) {
	tests := []struct {
		Host    string
		User    string
		Passwd  string
		Name    string
		SSLMode string
		Output  string
	}{
		{
			Host:   "", // empty means default
			Output: "postgres://:@127.0.0.1:5432?sslmode=",
		},
		{
			Host:    "/tmp/pg.sock",
			User:    "testuser",
			Passwd:  "space space !#$%^^%^```-=?=",
			Name:    "gitea",
			SSLMode: "false",
			Output:  "postgres://testuser:space%20space%20%21%23$%25%5E%5E%25%5E%60%60%60-=%3F=@:5432/gitea?host=%2Ftmp%2Fpg.sock&sslmode=false",
		},
		{
			Host:    "/tmp/pg.sock:6432",
			User:    "testuser",
			Passwd:  "pass",
			Name:    "gitea",
			SSLMode: "false",
			Output:  "postgres://testuser:pass@:6432/gitea?host=%2Ftmp%2Fpg.sock&sslmode=false",
		},
		{
			Host:    "localhost",
			User:    "pgsqlusername",
			Passwd:  "I love Gitea!",
			Name:    "gitea",
			SSLMode: "true",
			Output:  "postgres://pgsqlusername:I%20love%20Gitea%21@localhost:5432/gitea?sslmode=true",
		},
		{
			Host:   "localhost:1234",
			User:   "user",
			Passwd: "pass",
			Name:   "gitea?param=1",
			Output: "postgres://user:pass@localhost:1234/gitea?param=1&sslmode=",
		},
		{
			// Multi-host with same ports
			Host:    "host1,host2,host3",
			User:    "user",
			Passwd:  "pass",
			Name:    "forgejo",
			SSLMode: "disable",
			Output:  "postgres://user:pass@host1:5432,host2:5432,host3:5432/forgejo?sslmode=disable",
		},
		{
			// Multi-host with different ports
			Host:    "host1:5432,host2:5433",
			User:    "user",
			Passwd:  "pass",
			Name:    "forgejo",
			SSLMode: "require",
			Output:  "postgres://user:pass@host1:5432,host2:5433/forgejo?sslmode=require",
		},
		{
			// Multi-host IPv6
			Host:    "[::1]:1234,[::2]:2345",
			User:    "user",
			Passwd:  "pass",
			Name:    "forgejo",
			SSLMode: "disable",
			Output:  "postgres://user:pass@[::1]:1234,[::2]:2345/forgejo?sslmode=disable",
		},
		{
			// Multi-host with spaces (should be trimmed)
			Host:    "host1:5432 , host2:5433 , host3",
			User:    "user",
			Passwd:  "pass",
			Name:    "forgejo",
			SSLMode: "verify-full",
			Output:  "postgres://user:pass@host1:5432,host2:5433,host3:5432/forgejo?sslmode=verify-full",
		},
		{
			// Multi-host with database parameters
			Host:    "host1,host2",
			User:    "user",
			Passwd:  "pass",
			Name:    "forgejo?connect_timeout=10",
			SSLMode: "disable",
			Output:  "postgres://user:pass@host1:5432,host2:5432/forgejo?connect_timeout=10&sslmode=disable",
		},
	}

	for _, test := range tests {
		connStr := getPostgreSQLConnectionString(test.Host, test.User, test.Passwd, test.Name, test.SSLMode)
		assert.Equal(t, test.Output, connStr)
	}
}

func getPostgreSQLEngineGroupConnectionStrings(primaryHost, replicaHosts, user, passwd, name, sslmode string) (string, []string) {
	// Determine the primary connection string.
	primary := primaryHost
	if strings.TrimSpace(primary) == "" {
		primary = "127.0.0.1:5432"
	}
	primaryConn := getPostgreSQLConnectionString(primary, user, passwd, name, sslmode)

	// Build the replica connection strings.
	replicaConns := []string{}
	if strings.TrimSpace(replicaHosts) != "" {
		// Split comma-separated replica host values
		for h := range strings.SplitSeq(replicaHosts, ",") {
			trimmed := strings.TrimSpace(h)
			if trimmed != "" {
				replicaConns = append(replicaConns,
					getPostgreSQLConnectionString(trimmed, user, passwd, name, sslmode))
			}
		}
	}

	return primaryConn, replicaConns
}

func Test_getPostgreSQLEngineGroupConnectionStrings(t *testing.T) {
	tests := []struct {
		primaryHost    string // primary host setting (e.g. "localhost" or "[::1]:1234")
		replicaHosts   string // comma-separated replica hosts (e.g. "replica1,replica2:2345")
		user           string
		passwd         string
		name           string
		sslmode        string
		outputPrimary  string
		outputReplicas []string
	}{
		{
			// No primary override (empty => default) and no replicas.
			primaryHost:    "",
			replicaHosts:   "",
			user:           "",
			passwd:         "",
			name:           "",
			sslmode:        "",
			outputPrimary:  "postgres://:@127.0.0.1:5432?sslmode=",
			outputReplicas: []string{},
		},
		{
			// Primary set and one replica.
			primaryHost:    "localhost",
			replicaHosts:   "replicahost",
			user:           "user",
			passwd:         "pass",
			name:           "gitea",
			sslmode:        "disable",
			outputPrimary:  "postgres://user:pass@localhost:5432/gitea?sslmode=disable",
			outputReplicas: []string{"postgres://user:pass@replicahost:5432/gitea?sslmode=disable"},
		},
		{
			// Primary with explicit port; multiple replicas (one without and one with an explicit port).
			primaryHost:   "localhost:5433",
			replicaHosts:  "replica1,replica2:5434",
			user:          "test",
			passwd:        "secret",
			name:          "db",
			sslmode:       "require",
			outputPrimary: "postgres://test:secret@localhost:5433/db?sslmode=require",
			outputReplicas: []string{
				"postgres://test:secret@replica1:5432/db?sslmode=require",
				"postgres://test:secret@replica2:5434/db?sslmode=require",
			},
		},
		{
			// IPv6 addresses for primary and replica.
			primaryHost:   "[::1]:1234",
			replicaHosts:  "[::2]:2345",
			user:          "ipv6",
			passwd:        "ipv6pass",
			name:          "ipv6db",
			sslmode:       "disable",
			outputPrimary: "postgres://ipv6:ipv6pass@[::1]:1234/ipv6db?sslmode=disable",
			outputReplicas: []string{
				"postgres://ipv6:ipv6pass@[::2]:2345/ipv6db?sslmode=disable",
			},
		},
	}

	for _, test := range tests {
		primary, replicas := getPostgreSQLEngineGroupConnectionStrings(
			test.primaryHost,
			test.replicaHosts,
			test.user,
			test.passwd,
			test.name,
			test.sslmode,
		)
		assert.Equal(t, test.outputPrimary, primary)
		assert.Equal(t, test.outputReplicas, replicas)
	}
}

func Test_loadDBSetting(t *testing.T) {
	defer test.MockProtect(&Database)()
	t.Run("Does not overwrite Passwd", func(t *testing.T) {
		expectedPassword := "already_set"

		cfg, _ := NewConfigProviderFromData(`
			[database]
			PASSWD="new password"
		`)

		Database.Passwd = expectedPassword
		loadDBSetting(cfg)

		assert.Equal(t, expectedPassword, Database.Passwd)
	})
	t.Run("uses PASSWD", func(t *testing.T) {
		expectedPassword := "testpassword"

		cfg, _ := NewConfigProviderFromData(fmt.Sprintf(`
			[database]
			PASSWD="%s"
		`, expectedPassword))

		Database.Passwd = ""
		loadDBSetting(cfg)

		assert.Equal(t, expectedPassword, Database.Passwd)
	})
	t.Run("Uses PASSWD_URI", func(t *testing.T) {
		expectedPassword := "testpassworduri"

		uri := filepath.Join(t.TempDir(), "db_passwd")
		require.NoError(t, os.WriteFile(uri, []byte(expectedPassword), 0o644))

		cfg, _ := NewConfigProviderFromData(fmt.Sprintf(`
			[database]
			PASSWD_URI="file:%s"
		`, uri))

		Database.Passwd = ""
		loadDBSetting(cfg)

		assert.Equal(t, expectedPassword, Database.Passwd)
	})
}
