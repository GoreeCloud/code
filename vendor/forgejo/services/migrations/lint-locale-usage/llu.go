// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package lintLocaleUsage

import (
	"go/ast"
	"go/token"

	llu "forgejo.org/build/lint-locale-usage"
	"forgejo.org/modules/container"
)

// special case: services/migrations/migrate.go
func HandleMessengerInFunc(handler llu.Handler, fset *token.FileSet, n2 *ast.FuncDecl) {
	messenger := make(container.Set[string])
	for _, i := range n2.Type.Params.List {
		if ret, ok := i.Type.(*ast.SelectorExpr); ok && ret.Sel.Name == "Messenger" {
			if ret, ok := ret.X.(*ast.Ident); ok && ret.Name == "base" {
				for _, j := range i.Names {
					messenger.Add(j.Name)
				}
			}
		}
	}
	if len(messenger) == 0 {
		return
	}
	ast.Inspect(n2.Body, func(n ast.Node) bool {
		// search for "messenger" function calls
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ret, ok := call.Fun.(*ast.Ident); !(ok && messenger.Contains(ret.Name)) {
			return true
		}
		if len(call.Args) != 1 {
			handler.OnWarning(fset, call.Lparen, "unexpected invocation of base.Messenger (expected exactly 1 argument)")
			return true
		}
		handler.HandleGoTrArgument(fset, call.Args[0], "")
		return true
	})
}
