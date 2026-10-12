// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package lintLocaleUsage

import (
	"go/ast"
	"go/token"
	"strconv"

	llu "forgejo.org/build/lint-locale-usage"
)

func HandleCompositeUnit(handler llu.Handler, fset *token.FileSet, n *ast.CompositeLit) {
	ident, ok := n.Type.(*ast.Ident)
	if !ok || ident.Name != "Unit" {
		return
	}

	if len(n.Elts) != 6 {
		handler.OnWarning(fset, n.Pos(), "unexpected initialization of 'Unit' (unexpected number of arguments)")
		return
	}
	// NameKey has index 2
	//   invoked like '{{ctx.Locale.Tr $unit.NameKey}}'
	nameKey, ok := n.Elts[2].(*ast.BasicLit)
	if !ok || nameKey.Kind != token.STRING {
		handler.OnWarning(fset, n.Elts[2].Pos(), "unexpected initialization of 'Unit' (expected string literal as NameKey)")
		return
	}

	// extract string content
	arg, err := strconv.Unquote(nameKey.Value)
	if err == nil {
		// found interesting strings
		handler.OnMsgid(fset, nameKey.ValuePos, arg, false)
	}
}
