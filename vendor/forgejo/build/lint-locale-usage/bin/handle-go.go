// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"go/ast"
	goParser "go/parser"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"

	llu "forgejo.org/build/lint-locale-usage"
	lluAsymKey "forgejo.org/models/asymkey/lint-locale-usage"
	lluUnit "forgejo.org/models/unit/lint-locale-usage"
	lluMigrate "forgejo.org/services/migrations/lint-locale-usage"
)

// the `Handle*File` functions follow the following calling convention:
// * `fname` is the name of the input file
// * `src` is either `nil` (then the function invokes `ReadFile` to read the file)
//   or the contents of the file as {`[]byte`, or a `string`}

func HandleGoFile(handler llu.Handler, fname string, src any) error {
	fset := token.NewFileSet()
	node, err := goParser.ParseFile(fset, fname, src, goParser.SkipObjectResolution|goParser.ParseComments)
	if err != nil {
		return llu.LocatedError{
			Location: fname,
			Kind:     "Go parser",
			Err:      err,
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// search for function calls of the form `anything.Tr(any-string-lit, ...)`

		switch n2 := n.(type) {
		case *ast.CallExpr:
			if len(n2.Args) == 0 {
				return true
			}
			funSel, ok := n2.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ltf, ok := handler.LocaleTrFunctions[funSel.Sel.Name]
			if !ok {
				return true
			}

			var gotUnexpectedInvoke *int

			for _, argNum := range ltf {
				if len(n2.Args) <= int(argNum) {
					argc := len(n2.Args)
					gotUnexpectedInvoke = &argc
				} else {
					handler.HandleGoTrArgument(fset, n2.Args[int(argNum)], "")
				}
			}

			if gotUnexpectedInvoke != nil {
				handler.OnUnexpectedInvoke(fset, funSel.Sel.NamePos, funSel.Sel.Name, *gotUnexpectedInvoke)
			}

		case *ast.CompositeLit:
			if strings.HasSuffix(fname, "models/unit/unit.go") {
				lluUnit.HandleCompositeUnit(handler, fset, n2)
			} else if strings.Contains(fname, "models/asymkey/") {
				lluAsymKey.HandleCompositeErrorReason(handler, fset, n2)
			}

		case *ast.FuncDecl:
			matchInsPrefix := handler.HandleGoCommentGroup(fset, n2.Doc, "llu:returnsTrKey")
			if matchInsPrefix != nil {
				results := n2.Type.Results.List
				if len(results) != 1 {
					handler.OnWarning(fset, n2.Type.Func, fmt.Sprintf("function %s has unexpected return type; expected single return value", n2.Name.Name))
					return true
				}

				ast.Inspect(n2.Body, func(n ast.Node) bool {
					// search for return stmts
					// TODO: what about nested functions?
					if ret, ok := n.(*ast.ReturnStmt); ok {
						for _, res := range ret.Results {
							ast.Inspect(res, func(n ast.Node) bool {
								if expr, ok := n.(ast.Expr); ok {
									handler.HandleGoTrArgument(fset, expr, *matchInsPrefix)
								}
								return true
							})
						}
						return false
					}
					return true
				})
			}

			if strings.HasSuffix(fname, "services/migrations/migrate.go") {
				lluMigrate.HandleMessengerInFunc(handler, fset, n2)
			}
			return true
		case *ast.GenDecl:
			switch n2.Tok {
			case token.CONST, token.VAR:
				matchInsPrefix := handler.HandleGoCommentGroup(fset, n2.Doc, " llu:TrKeys")
				if matchInsPrefix == nil {
					return true
				}
				for _, spec := range n2.Specs {
					// interpret all contained strings as message IDs
					ast.Inspect(spec, func(n ast.Node) bool {
						if argLit, ok := n.(*ast.BasicLit); ok {
							handler.HandleGoTrBasicLit(fset, argLit, *matchInsPrefix)
							return false
						}
						return true
					})
				}

			case token.TYPE:
				// modules/web/middleware/binding.go:Validate uses the convention that structs
				// entries can have tags.
				// In particular, `locale:$msgid` should be handled; any fields with `form:-` shouldn't.
				// Problem: we don't know which structs are forms, actually.

				for _, spec := range n2.Specs {
					tspec := spec.(*ast.TypeSpec)
					structNode, ok := tspec.Type.(*ast.StructType)
					if !ok || !(strings.HasSuffix(tspec.Name.Name, "Form") ||
						(tspec.Doc != nil &&
							slices.ContainsFunc(tspec.Doc.List, func(c *ast.Comment) bool {
								return c.Text == "// swagger:model"
							}))) {
						continue
					}
					for _, field := range structNode.Fields.List {
						if field.Names == nil {
							continue
						}
						if len(field.Names) != 1 {
							handler.OnWarning(fset, field.Type.Pos(), "unsupported multiple field names")
							continue
						}
						msgidPos := field.Names[0].NamePos
						msgid := "form." + field.Names[0].Name
						if field.Tag != nil && field.Tag.Kind == token.STRING {
							rawTag, err := strconv.Unquote(field.Tag.Value)
							if err != nil {
								handler.OnWarning(fset, field.Tag.ValuePos, "invalid tag value encountered")
								continue
							}
							tag := reflect.StructTag(rawTag)
							if tag.Get("form") == "-" {
								continue
							}
							tmp := tag.Get("locale")
							if len(tmp) != 0 {
								msgidPos = field.Tag.ValuePos
								msgid = tmp
							}
						}
						handler.OnMsgid(fset, msgidPos, msgid, true)
					}
				}
			}
		}

		return true
	})

	return nil
}
