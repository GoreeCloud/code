// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package meilisearch

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"forgejo.org/modules/indexer/issues/internal"
	"forgejo.org/modules/indexer/issues/internal/tests"

	"github.com/meilisearch/meilisearch-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeilisearchIndexer(t *testing.T) {
	t.Skip("meilisearch not found in Forgejo test yet")
	// The meilisearch instance started by pull-db-tests.yml > test-unit > services > meilisearch
	url := "http://meilisearch:7700"
	key := "" // auth has been disabled in test environment

	if os.Getenv("CI") == "" {
		// Make it possible to run tests against a local meilisearch instance
		url = os.Getenv("TEST_MEILISEARCH_URL")
		if url == "" {
			t.Skip("TEST_MEILISEARCH_URL not set and not running in CI")
			return
		}
		key = os.Getenv("TEST_MEILISEARCH_KEY")
	}

	require.Eventually(t, func() bool {
		resp, err := http.Get(url)
		return err == nil && resp.StatusCode == http.StatusOK
	}, time.Minute, time.Microsecond*100, "Failed to wait for meilisearch to be up")

	indexer := NewIndexer(url, key, fmt.Sprintf("test_meilisearch_indexer_%d", time.Now().Unix()))
	defer indexer.Close()

	tests.TestIndexer(t, indexer)
}

func TestConvertHits(t *testing.T) {
	for _, invalidID := range []string{"\"aa\"", "{\"aa\":\"123\"}", "[\"aa\"]"} {
		_, err := convertHits(&meilisearch.SearchResponse{
			Hits: meilisearch.Hits{
				meilisearch.Hit{"id": []byte(invalidID)},
			},
		})
		require.ErrorIs(t, err, ErrMalformedResponse)
	}

	validResponse := &meilisearch.SearchResponse{
		Hits: meilisearch.Hits{
			meilisearch.Hit{
				"id":       []byte("11"),
				"title":    []byte("\"a title\""),
				"content":  []byte("\"issue body with no match\""),
				"comments": []byte("[\"hey what's up?\", \"I'm currently bowling\", \"nice\"]"),
			},
			meilisearch.Hit{
				"id":       []byte("22"),
				"title":    []byte("\"Bowling as title\""),
				"content":  []byte("\"\""),
				"comments": []byte("[]"),
			},
			meilisearch.Hit{
				"id":       []byte("33"),
				"title":    []byte("\"Bowl-ing as fuzzy match\""),
				"content":  []byte("\"\""),
				"comments": []byte("[]"),
			},
		},
	}
	hits, err := convertHits(validResponse)
	require.NoError(t, err)
	assert.Equal(t, []internal.Match{{ID: 11}, {ID: 22}, {ID: 33}}, hits)
}

func TestDoubleQuoteKeyword(t *testing.T) {
	assert.Empty(t, doubleQuoteKeyword(""))
	assert.Equal(t, `"a" "b" "c"`, doubleQuoteKeyword("a b c"))
	assert.Equal(t, `"a" "d" "g"`, doubleQuoteKeyword("a  d g"))
	assert.Equal(t, `"a" "d" "g"`, doubleQuoteKeyword("a  d g"))
	assert.Equal(t, `"a" "d" "g"`, doubleQuoteKeyword(`a  "" "d" """g`))
}
