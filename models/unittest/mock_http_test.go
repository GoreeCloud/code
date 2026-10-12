// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package unittest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: This is a test of the unittest helper itself
func TestMockWebServer(t *testing.T) {
	server := NewMockWebServer(t, "https://example.com", "testdata", false)
	defer server.Close()
	request, err := http.NewRequest("GET", server.URL+"/", nil)
	require.NoError(t, err)
	response, err := server.Client().Do(request)
	require.NoError(t, err)
	assert.Len(t, response.Header["Header"], 1)
	assert.Equal(t, "value", response.Header["Header"][0])
}
