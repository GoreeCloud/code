package integration

import (
	"net/http"
	"testing"

	code_indexer "forgejo.org/modules/indexer/code"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/tests"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestExploreCodeSearchIndexer(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	defer test.MockVariableValue(&setting.Indexer.RepoIndexerEnabled, true)()

	t.Run("Exact", func(t *testing.T) {
		req := NewRequest(t, "GET", "/explore/code?q=file&mode=exact")
		resp := MakeRequest(t, req, http.StatusOK)
		doc := NewHTMLParser(t, resp.Body).Find(".explore")

		active, ok := doc.Find("[data-test-tag=fuzzy-dropdown] .active input").Attr("value")
		assert.True(t, ok)
		assert.Equal(t, "exact", active)

		doc.Find(".file-body").Each(func(i int, sel *goquery.Selection) {
			assert.Positive(t, sel.Find(".code-inner").Find(".search-highlight").Length())
		})
	})

	t.Run("Fuzzy", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Indexer.RepoIndexerEnableFuzzy, true)()
		code_indexer.CodeSearchOptions = []string{"exact", "union", "fuzzy"} // usually set by Init

		req := NewRequest(t, "GET", "/explore/code?q=file&mode=fuzzy")
		resp := MakeRequest(t, req, http.StatusOK)
		doc := NewHTMLParser(t, resp.Body).Find(".explore")

		active, ok := doc.Find("[data-test-tag=fuzzy-dropdown] .active input").Attr("value")
		assert.True(t, ok)
		assert.Equal(t, "fuzzy", active)
	})

	t.Run("No Fuzzy", func(t *testing.T) {
		defer test.MockVariableValue(&setting.Indexer.RepoIndexerEnableFuzzy, false)()
		code_indexer.CodeSearchOptions = []string{"exact", "union"} // usually set by Init

		req := NewRequest(t, "GET", "/explore/code?q=file&mode=fuzzy")
		resp := MakeRequest(t, req, http.StatusOK)
		doc := NewHTMLParser(t, resp.Body).Find(".explore")

		active, ok := doc.Find("[data-test-tag=fuzzy-dropdown] .active input").Attr("value")
		assert.True(t, ok)
		assert.Equal(t, "union", active)
	})
}
