// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package lintLocaleUsage

import (
	"go/ast"
	"go/token"

	llu "forgejo.org/build/lint-locale-usage"
)

// special case: models/asymkey/*.go,
//
//	handle &ObjectVerification{...}
func HandleCompositeErrorReason(handler llu.Handler, fset *token.FileSet, n *ast.CompositeLit) {
	ident, ok := n.Type.(*ast.Ident)
	if !ok || ident.Name != "ObjectVerification" {
		return
	}

	// fields are normally named
	for _, i := range n.Elts {
		if kve, ok := i.(*ast.KeyValueExpr); ok {
			ident, ok = kve.Key.(*ast.Ident)
			if ok && ident.Name == "Reason" {
				handler.HandleGoTrArgument(fset, kve.Value, "")
			}
		} else {
			handler.OnWarning(fset, i.Pos(), "unable to parse ObjectVerification field assignment")
		}
	}
}
