// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package code

import (
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHighlightSearchResultCode(t *testing.T) {
	opts := []struct {
		Title  string
		File   string
		Lines  []int
		Range  [][3]int
		Code   string
		Result []template.HTML
	}{
		{
			Title: "One Match Text",
			File:  "test.txt",
			Range: [][3]int{{1, 5, 9}},
			Code:  "First Line\nMark this only\nThe End",
			Result: []template.HTML{
				"First Line",
				"Mark <span class=\"search-highlight\">this</span> only",
				"The End",
			},
		},
		{
			Title: "Two Match Text",
			File:  "test.txt",
			Range: [][3]int{
				{1, 5, 9},
				{2, 5, 9},
			},
			Code: "First Line\nMark this only\nMark this too\nThe End",
			Result: []template.HTML{
				"First Line",
				"Mark <span class=\"search-highlight\">this</span> only",
				"Mark <span class=\"search-highlight\">this</span> too",
				"The End",
			},
		},
		{
			Title: "Unicode Before",
			File:  "test.txt",
			Range: [][3]int{{1, 10, 14}},
			Code:  "First Line\nMark 👉 this only\nThe End",
			Result: []template.HTML{
				"First Line",
				"Mark 👉 <span class=\"search-highlight\">this</span> only",
				"The End",
			},
		},
		{
			Title: "Unicode Between",
			File:  "test.txt",
			Range: [][3]int{{1, 5, 14}},
			Code:  "First Line\nMark this 😊 only\nThe End",
			Result: []template.HTML{
				"First Line",
				"Mark <span class=\"search-highlight\">this 😊</span> only",
				"The End",
			},
		},
		{
			Title: "Unicode Before And Between",
			File:  "test.txt",
			Range: [][3]int{{1, 10, 19}},
			Code:  "First Line\nMark 👉 this 😊 only\nThe End",
			Result: []template.HTML{
				"First Line",
				"Mark 👉 <span class=\"search-highlight\">this 😊</span> only",
				"The End",
			},
		},
		{
			Title: "Golang",
			File:  "test.go",
			Range: [][3]int{{1, 14, 23}},
			Code:  "func main() {\n\tfmt.Println(\"mark this\")\n}",
			Result: []template.HTML{
				"<span class=\"kd\">func</span><span class=\"w\"> </span><span class=\"nf\">main</span><span class=\"p\">(</span><span class=\"p\">)</span><span class=\"w\"> </span><span class=\"p\">{</span><span class=\"w\">",
				"</span><span class=\"w\">\t</span><span class=\"nx\">fmt</span><span class=\"p\">.</span><span class=\"nf\">Println</span><span class=\"p\">(</span><span class=\"s\">&#34;<span class=\"search-highlight\">mark this</span>&#34;</span><span class=\"p\">)</span><span class=\"w\">",
				"</span><span class=\"p\">}</span>",
			},
		},
		{
			Title: "Golang Unicode",
			File:  "test.go",
			Range: [][3]int{{1, 14, 28}},
			Code:  "func main() {\n\tfmt.Println(\"mark this 😊\")\n}",
			Result: []template.HTML{
				"<span class=\"kd\">func</span><span class=\"w\"> </span><span class=\"nf\">main</span><span class=\"p\">(</span><span class=\"p\">)</span><span class=\"w\"> </span><span class=\"p\">{</span><span class=\"w\">",
				"</span><span class=\"w\">\t</span><span class=\"nx\">fmt</span><span class=\"p\">.</span><span class=\"nf\">Println</span><span class=\"p\">(</span><span class=\"s\">&#34;<span class=\"search-highlight\">mark this 😊</span>&#34;</span><span class=\"p\">)</span><span class=\"w\">",
				"</span><span class=\"p\">}</span>",
			},
		},
	}
	for _, o := range opts {
		t.Run(o.Title, func(t *testing.T) {
			lines := []int{}
			for i := range strings.Count(strings.TrimSuffix(o.Code, "\n"), "\n") + 1 {
				lines = append(lines, i+1)
			}
			res := HighlightSearchResultCode(o.File, lines, o.Range, o.Code)
			assert.Len(t, res, len(o.Result))
			assert.Len(t, res, len(lines))

			for i, r := range res {
				require.Equal(t, lines[i], r.Num)
				require.Equal(t, o.Result[i], r.FormattedContent)
			}
		})
	}
}
