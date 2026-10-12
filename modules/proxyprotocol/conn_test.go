// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package proxyprotocol_test

import (
	"io"
	"net"
	"testing"
	"time"

	"forgejo.org/modules/proxyprotocol"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var v2Header = []byte{0xd, 0xa, 0xd, 0xa, 0x0, 0xd, 0xa, 0x51, 0x55, 0x49, 0x54, 0xa}

func testConnection(t *testing.T, input []byte) *proxyprotocol.Conn {
	local, remote := net.Pipe()
	conn := proxyprotocol.NewConn(remote, 10*time.Second)

	go func(t *testing.T, conn net.Conn) {
		_, err := conn.Write(input)
		require.NoError(t, err)

		err = conn.Close()
		require.NoError(t, err)
	}(t, local)

	return conn
}

func assertUwu(t *testing.T, conn *proxyprotocol.Conn) {
	buf := make([]byte, 3)
	read, err := conn.Read(buf)
	require.NoError(t, err)

	assert.Equal(t, 3, read)
	assert.Equal(t, []byte("uwu"), buf)
}

func TestProxyProtocolParse(t *testing.T) {
	// Basic v4/v6 TCP
	ipv4Conn := testConnection(t, []byte("PROXY TCP4 7.3.3.1 1.3.3.7 14231 443\r\nuwu"))

	assertUwu(t, ipv4Conn)
	assert.Equal(t, "7.3.3.1:14231", ipv4Conn.RemoteAddr().String())
	assert.Equal(t, "1.3.3.7:443", ipv4Conn.LocalAddr().String())

	ipv6Conn := testConnection(t, []byte("PROXY TCP6 fe80::2 fe80::1 28512 443\r\nuwu"))

	assertUwu(t, ipv6Conn)
	assert.Equal(t, "[fe80::2]:28512", ipv6Conn.RemoteAddr().String())
	assert.Equal(t, "[fe80::1]:443", ipv6Conn.LocalAddr().String())

	ipv4Conn = testConnection(t, append(v2Header, 0x21, 0x11, 0x0, 0xc, 0x7, 0x3, 0x3, 0x1, 0x1, 0x3, 0x3, 0x7, 0xd9, 0xec, 0x1, 0xbb, 0x75, 0x77, 0x75))

	assertUwu(t, ipv4Conn)
	assert.Equal(t, "7.3.3.1:55788", ipv4Conn.RemoteAddr().String())
	assert.Equal(t, "1.3.3.7:443", ipv4Conn.LocalAddr().String())

	ipv6Conn = testConnection(t, append(v2Header, 0x21, 0x21, 0x0, 0x24, 0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xd9, 0xec, 0x1, 0xbb, 0x75, 0x77, 0x75))

	assertUwu(t, ipv6Conn)
	assert.Equal(t, "[fe80::2]:55788", ipv6Conn.RemoteAddr().String())
	assert.Equal(t, "[fe80::1]:443", ipv6Conn.LocalAddr().String())

	// Basic unknown
	unknownConn := testConnection(t, []byte("PROXY UNKNOWN\r\nuwu"))
	_, err := unknownConn.Read([]byte{})
	require.Error(t, err)

	// Accept unknown protocol types
	defer test.MockVariableValue(&setting.ProxyProtocolAcceptUnknown, true)()

	unknownConn = testConnection(t, []byte("PROXY UNKNOWN\r\nuwu"))

	assertUwu(t, unknownConn)
	assert.Equal(t, "pipe", unknownConn.RemoteAddr().String())
	assert.Equal(t, "pipe", unknownConn.LocalAddr().String())

	// Discard any unknown information between "UNKNOWN" and CRLF

	unknownConn = testConnection(t, []byte("PROXY UNKNOWN look, I'm hinding in an unknown protocol \\o/\r\nuwu"))

	assertUwu(t, unknownConn)
	assert.Equal(t, "pipe", unknownConn.RemoteAddr().String())
	assert.Equal(t, "pipe", unknownConn.LocalAddr().String())

	// Basic local
	unknownConnV2 := testConnection(t, append(v2Header, 0x20, 0x0, 0x0, 0x0, 0x75, 0x77, 0x75))

	assertUwu(t, unknownConnV2)
	assert.Equal(t, "pipe", unknownConnV2.RemoteAddr().String())
	assert.Equal(t, "pipe", unknownConnV2.LocalAddr().String())
}

func TestProxyProtocolInvalidHeader(t *testing.T) {
	// Short prefix
	conn := testConnection(t, []byte("PROXY\r\n"))
	_, err := conn.Read([]byte{})
	require.ErrorIs(t, err, io.EOF)

	// Wrong prefix
	conn = testConnection(t, []byte("PROXYv1337\r\n"))
	_, err = conn.Read([]byte{})
	require.ErrorContains(t, err, "Unexpected proxy header")
}
