// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package singleresponse

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/cfg"
)

var Analyzer = &analysis.Analyzer{
	Name:     "singleresponse",
	Doc:      "checks that Forgejo web response methods are only invoked once in a control flow",
	Requires: []*analysis.Analyzer{inspect.Analyzer, ctrlflow.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	cfgs := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)

	webFuncs := map[string]map[string]any{
		"*forgejo.org/services/context.APIContext": {
			"Error":                 true,
			"InternalServerError":   true,
			"NotFound":              true,
			"NotFoundOrServerError": true,
			"ServerError":           true,
		},
		"*forgejo.org/services/context.Base": {
			"Error":               true,
			"JSON":                true,
			"JSONWithContentType": true,
			"PlainText":           true,
			"PlainTextBytes":      true,
			"Redirect":            true,
			"ServeContent":        true,
		},
		"*forgejo.org/services/context.Context": {
			"HTML":                  true,
			"JSONError":             true,
			"JSONOK":                true,
			"JSONRedirect":          true,
			"JSONTemplate":          true,
			"NotFound":              true,
			"NotFoundOrServerError": true,
			"RedirectToFirst":       true,
			"RenderWithErr":         true,
			"ServerError":           true,
		},
		// Future: RedirectToUser does not accept a ctx LHS, but rather a first parameter -- needs different
		// implementation of detection, or, refactoring: "RedirectToUser": true,
	}

	insp.Nodes([]ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}, func(n ast.Node, push bool) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			// Skip test methods which are assumed to know what they're doing.
			if strings.HasPrefix(fn.Name.Name, "Test") {
				return false
			}
			cfg := cfgs.FuncDecl(fn)
			if cfg == nil {
				return true
			}
			inspectFunction(cfg, pass, webFuncs)
		case *ast.FuncLit:
			cfg := cfgs.FuncLit(fn)
			if cfg == nil {
				return true
			}
			inspectFunction(cfg, pass, webFuncs)
		}
		return false
	})

	return nil, nil //nolint:nilnil
}

func inspectFunction(cfg *cfg.CFG, pass *analysis.Pass, webFuncs map[string]map[string]any) {
	for _, block := range cfg.Blocks {
		for nodeIdx, node := range block.Nodes {
			ast.Inspect(node, func(n ast.Node) bool {
				// Don't recurse inside of a function literal inside of a function declaration, as this isn't
				// related to the control flow that we're currently iterating through.
				_, isFuncLit := n.(*ast.FuncLit)
				if isFuncLit {
					return false
				}

				call, isCall := n.(*ast.CallExpr)
				if !isCall {
					return true
				}

				// SelectorExpr: "an expression followed by a selector", like "ctx.Error".  All the functions
				// we're interested in match this pattern.
				selector, isSelector := call.Fun.(*ast.SelectorExpr)
				if !isSelector {
					return false
				}

				// We almost get the right information easily from the selector by using
				// pass.TypesInfo.Uses[selector.X] -- but that will be the type of the variable that we're
				// invoking a method on, and not the type of the method receiver.  eg. on `ctx
				// *context.Context`, `ctx.ServerError(...)` will always be `*context.Context`, even if
				// `ServerError` is actually implemented on `*context.Base`.
				//
				// We need to dig a little deeper here to get the function type, then its signature, and then
				// it's receiver type, and we'll really have the method that will be invoked rather than just
				// the variable that it is called upon.
				selection, hasSelection := pass.TypesInfo.Selections[selector]
				if !hasSelection {
					return false
				}
				objFn, ok := selection.Obj().(*types.Func)
				if !ok {
					return false
				}
				fnSig, ok := objFn.Type().(*types.Signature)
				if !ok {
					return false
				}
				callType := fnSig.Recv().Type().String()

				typeMap, inTypeMap := webFuncs[callType]
				if inTypeMap {
					callName := selector.Sel.Name
					_, inFuncMap := typeMap[callName]
					if inFuncMap {
						// OK... we've found a call to a terminating function at
						// cfg.Blocks[blockIdx].Nodes[nodeIdx].
						trace := false
						// For code-time debugging/analysis, set trace=true when digging into why something isn't
						// working:
						// if callName == "InternalServerError" {
						// 	trace = true
						// }
						sketchy := inspectCallSite(block, nodeIdx, trace)
						if sketchy != nil {
							pass.Reportf(node.Pos(), "Invocation of %s / %s, and control flow continues afterwards.", callType, callName)
						}
					}
				}

				return false
			})
		}
	}
}

type sketchyCall struct{}

func inspectCallSite(callingBlock *cfg.Block, callingNodeIndex int, trace bool) *sketchyCall {
	// Inspect the remainder of the block passed in, after callingNodeIndex, for "bad" statements
	if trace {
		println("remainder of block...")
	}
	for _, nextStmt := range callingBlock.Nodes[callingNodeIndex+1:] {
		if trace {
			println(fmt.Sprintf("\tnextStmt = %#v", nextStmt))
		}
		// Only `return` is permitted after one of the web return functions; maybe this needs to expand in the future
		// but haven't identified any cases in Forgejo yet.
		_, stmtOk := nextStmt.(*ast.ReturnStmt)
		if !stmtOk {
			if trace {
				println(fmt.Sprintf("\tfound sketchy statement = %#v", nextStmt))
			}
			// Future: add information about what was following the call, so that the diagnostic can be more specific
			// about the problematic next statement identified... but so far it seems pretty easy to analyze and fix.
			return &sketchyCall{}
		}
	}
	if trace {
		println("nothing found in remainder of block")
		println(fmt.Sprintf("%d Succs blocks will be investigated", len(callingBlock.Succs)))
	}

	// Now, assuming that there was nothing problematic found in the remainder of the block, use the control-flow graph
	// to identify where code execution would continue and see if there's anything inappropriate in it.
	//
	// https://pkg.go.dev/golang.org/x/tools@v0.46.0/go/cfg#Block -> A block may have 0-2 successors: zero for a return
	// block or a block that calls a function such as panic that never returns; one for a normal (jump) block; and two
	// for a conditional (if) block.
	//
	// It's possible for the next block to have either no nodes, or, no nodes that continue to do work and trigger
	// detection... but then to proceed into *another* block that does.  So this investigation has to be done
	// recursively.  Control-flow graph should prevent us from needing to stop this recursive detection; we'll hit a
	// return statement or end of function and that's the end of the CFG, and that's also the time we'd want to stop
	// looking, so no additional exit logic should be needed.
	for i, succ := range callingBlock.Succs {
		if trace {
			println(fmt.Sprintf("Succs[%d], block index %d, recursing:", i, succ.Index))
		}
		// `-1` is used to start at index 0 in the nodes.
		sketchy := inspectCallSite(succ, -1, trace)
		if trace {
			println(fmt.Sprintf("Succs[%d], block index %d, had sketchy = %#v", i, succ.Index, sketchy))
		}
		if sketchy != nil {
			return sketchy
		}
	}

	return nil
}
