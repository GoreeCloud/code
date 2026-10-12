// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package lintLocaleUsage

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func (handler Handler) HandleGoTrBasicLit(fset *token.FileSet, argLit *ast.BasicLit, prefix string) {
	if argLit.Kind == token.STRING {
		// extract string content
		arg, err := strconv.Unquote(argLit.Value)
		if err != nil {
			return
		}
		// found interesting strings
		arg = prefix + arg
		if strings.HasSuffix(arg, ".") || strings.HasSuffix(arg, "_") {
			prep, trunc := PrepareMsgidPrefix(arg)
			if trunc {
				handler.OnWarning(fset, argLit.ValuePos, fmt.Sprintf("needed to truncate message id prefix: %s", arg))
			}
			handler.OnMsgidPrefix(fset, argLit.ValuePos, prep, trunc)
		} else {
			handler.OnMsgid(fset, argLit.ValuePos, arg, false)
		}
	}
}

func (handler Handler) HandleGoTrArgument(fset *token.FileSet, n ast.Expr, prefix string) {
	switch n := n.(type) {
	case *ast.BasicLit:
		handler.HandleGoTrBasicLit(fset, n, prefix)

	case *ast.BinaryExpr:
		if n.Op != token.ADD {
			// pass
		} else if argLit, ok := n.X.(*ast.BasicLit); ok && argLit.Kind == token.STRING {
			// extract string content
			arg, err := strconv.Unquote(argLit.Value)
			if err != nil {
				return
			}
			// found interesting strings
			arg = prefix + arg
			prep, trunc := PrepareMsgidPrefix(arg)
			if trunc {
				handler.OnWarning(fset, argLit.ValuePos, fmt.Sprintf("needed to truncate message id prefix: %s", arg))
			}
			handler.OnMsgidPrefix(fset, argLit.ValuePos, prep, trunc)
		}

	case *ast.CallExpr:
		if selExpr, ok := n.Fun.(*ast.SelectorExpr); ok {
			if xIdent, xok := selExpr.X.(*ast.Ident); !xok || xIdent.Name != "fmt" {
				return
			}
			if selExpr.Sel.Name != "Sprintf" {
				handler.OnWarning(fset, selExpr.Sel.NamePos, fmt.Sprintf("unexpected formatting function encountered: %s", selExpr.Sel.Name))
				return
			}
			if len(n.Args) == 0 {
				handler.OnWarning(fset, selExpr.Sel.NamePos, fmt.Sprintf("unexpected formatting function invocation (no arguments) of '%s'", selExpr.Sel.Name))
				return
			}

			if argLit, ok := n.Args[0].(*ast.BasicLit); ok && argLit.Kind == token.STRING {
				// extract string content
				arg, err := strconv.Unquote(argLit.Value)
				if err != nil {
					return
				}
				if strings.Contains(arg, " ") {
					handler.OnWarning(fset, argLit.ValuePos, fmt.Sprintf(
						"formatting function invocation of '%s' with weird msgid format string: %s",
						selExpr.Sel.Name,
						arg,
					))
					return
				}
				// found interesting strings
				handler.OnMsgidPattern(fset, argLit.ValuePos, prefix+arg)
			}
		}
	}
}

func (handler Handler) HandleGoCommentGroup(fset *token.FileSet, cg *ast.CommentGroup, commentPrefix string) *string {
	if cg == nil {
		return nil
	}
	var matches []token.Pos
	matchInsPrefix := ""
	commentPrefix = "//" + commentPrefix
	for _, comment := range cg.List {
		ctxt := strings.TrimSpace(comment.Text)
		if ctxt == commentPrefix {
			matches = append(matches, comment.Slash)
		} else if after, found := strings.CutPrefix(ctxt, commentPrefix+"Suffix "); found {
			matches = append(matches, comment.Slash)
			matchInsPrefix = strings.TrimSpace(after)
		}
	}
	switch len(matches) {
	case 0:
		return nil
	case 1:
		return &matchInsPrefix
	default:
		handler.OnWarning(
			fset,
			matches[0],
			fmt.Sprintf("encountered multiple %s... directives, ignoring", strings.TrimSpace(commentPrefix)),
		)
		return &matchInsPrefix
	}
}
