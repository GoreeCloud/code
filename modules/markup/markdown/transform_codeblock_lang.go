// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package markdown

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func (g *ASTTransformer) transformCodeblockLanguage(v *ast.FencedCodeBlock, reader text.Reader) {
	if v.Info == nil {
		return
	}
	src := reader.Source()
	info := v.Info.Segment.Value(src)

	// Parse Pandoc style attributes
	// https://pandoc.org/MANUAL.html#extension-fenced_code_attributes
	//
	// For example,
	// ```{.haskell .numberLines}
	// ...
	// ```
	// Should have a language of "haskell", not "{.haskell .numberLines}"
	if trimmed := bytes.TrimSpace(info); bytes.HasPrefix(trimmed, []byte{'{'}) && bytes.HasSuffix(trimmed, []byte{'}'}) {
		attributes := trimmed[1 : len(trimmed)-1]
		for attribute := range bytes.SplitSeq(attributes, []byte{' '}) {
			if class, found := bytes.CutPrefix(attribute, []byte{'.'}); found {
				if lexer := lexers.Get(string(class)); lexer != nil {
					lang := class
					langInx := bytes.Index(info, lang)
					start := v.Info.Segment.Start + langInx
					end := start + len(lang)
					v.Info = ast.NewTextSegment(text.NewSegment(start, end))
					return
				}
			}
		}
		return
	}

	// Strip language after commas
	//
	// For example,
	// ```rust,ignore
	// ...
	// ```
	// Should have a language of "rust", not "rust,ignore"
	if i := bytes.IndexByte(info, ','); i != -1 {
		start := v.Info.Segment.Start
		v.Info = ast.NewTextSegment(text.NewSegment(start, start+i))
	}
}
