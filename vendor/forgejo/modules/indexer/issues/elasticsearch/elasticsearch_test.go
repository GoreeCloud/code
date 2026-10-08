// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package elasticsearch

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"forgejo.org/modules/indexer/issues/internal/tests"

	"github.com/stretchr/testify/require"
)

func TestElasticsearchIndexer(t *testing.T) {
	url := os.Getenv("TEST_ELASTICSEARCH_URL")
	if url == "" {
		t.Skip("TEST_ELASTICSEARCH_URL not set")
		return
	}

	require.Eventually(t, func() bool {
		resp, err := http.Get(url)
		return err == nil && resp.StatusCode == http.StatusOK
	}, time.Minute, time.Microsecond*100, "Failed to wait for elasticsearch to be up")

	indexer := NewIndexer(url, fmt.Sprintf("test_elasticsearch_indexer_%d", time.Now().Unix()))
	defer indexer.Close()

	tests.TestIndexer(t, indexer)
}
