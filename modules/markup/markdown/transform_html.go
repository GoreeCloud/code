// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package markdown

import (
	"strings"

	"forgejo.org/modules/markup"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func (g *ASTTransformer) addTypeToButton(v *ast.RawHTML, segment string) {
	segment = strings.TrimPrefix(segment, "<button")
	newTag := ast.NewString([]byte(`<button type="button"` + segment))
	newTag.SetCode(true)
	v.Parent().ReplaceChild(v.Parent(), v, newTag)
}

func (g *ASTTransformer) transformRawHTML(_ *markup.RenderContext, v *ast.RawHTML, reader text.Reader) {
	segment := string(v.Segments.Value(reader.Source()))

	if strings.HasPrefix(segment, "<button") {
		g.addTypeToButton(v, segment)
	}
}
