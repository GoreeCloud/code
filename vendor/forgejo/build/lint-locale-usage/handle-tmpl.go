// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package lintLocaleUsage

import (
	"fmt"
	"go/token"
	"os"
	"strings"
	"text/template"
	tmplParser "text/template/parse"

	fjTemplates "forgejo.org/modules/templates"
	"forgejo.org/modules/util"
)

// derived from source: modules/templates/scopedtmpl/scopedtmpl.go, L169-L213
func (handler Handler) handleTemplateNode(fset *token.FileSet, node tmplParser.Node) {
	switch node.Type() {
	case tmplParser.NodeAction:
		handler.handleTemplatePipeNode(fset, node.(*tmplParser.ActionNode).Pipe)
	case tmplParser.NodeList:
		nodeList := node.(*tmplParser.ListNode)
		handler.handleTemplateFileNodes(fset, nodeList.Nodes)
	case tmplParser.NodePipe:
		handler.handleTemplatePipeNode(fset, node.(*tmplParser.PipeNode))
	case tmplParser.NodeTemplate:
		handler.handleTemplatePipeNode(fset, node.(*tmplParser.TemplateNode).Pipe)
	case tmplParser.NodeIf:
		nodeIf := node.(*tmplParser.IfNode)
		handler.handleTemplateBranchNode(fset, nodeIf.BranchNode)
	case tmplParser.NodeRange:
		nodeRange := node.(*tmplParser.RangeNode)
		handler.handleTemplateBranchNode(fset, nodeRange.BranchNode)
	case tmplParser.NodeWith:
		nodeWith := node.(*tmplParser.WithNode)
		handler.handleTemplateBranchNode(fset, nodeWith.BranchNode)

	case tmplParser.NodeCommand:
		nodeCommand := node.(*tmplParser.CommandNode)

		handler.handleTemplateFileNodes(fset, nodeCommand.Args)

		if len(nodeCommand.Args) < 2 {
			return
		}

		funcname := ""
		switch nodeCommand.Args[0].Type() {
		case tmplParser.NodeChain:
			nodeChain := nodeCommand.Args[0].(*tmplParser.ChainNode)
			if nodeIdent, ok := nodeChain.Node.(*tmplParser.IdentifierNode); ok {
				if nodeIdent.Ident != "ctx" || len(nodeChain.Field) != 2 || nodeChain.Field[0] != "Locale" {
					return
				}
				funcname = nodeChain.Field[1]
			}

		case tmplParser.NodeField:
			nodeField := nodeCommand.Args[0].(*tmplParser.FieldNode)
			if len(nodeField.Ident) != 2 || nodeField.Ident[0] != "locale" {
				return
			}
			resolvedPos := fset.PositionFor(token.Pos(nodeCommand.Pos), false)
			if !strings.Contains(resolvedPos.Filename, "templates/mail/") {
				handler.OnWarning(fset, token.Pos(nodeCommand.Pos), "encountered unexpected .locale usage")
			}
			funcname = nodeField.Ident[1]

		case tmplParser.NodeVariable:
			nodeVar := nodeCommand.Args[0].(*tmplParser.VariableNode)
			if len(nodeVar.Ident) != 3 || !(nodeVar.Ident[0] == "$" && nodeVar.Ident[1] == "locale") {
				return
			}
			funcname = nodeVar.Ident[2]
		}

		var gotUnexpectedInvoke *int
		ltf, ok := handler.LocaleTrFunctions[funcname]
		if !ok {
			return
		}

		for _, argNum := range ltf {
			if len(nodeCommand.Args) >= int(argNum+2) {
				handler.handleTemplateMsgid(fset, nodeCommand.Args[int(argNum+1)])
			} else {
				argc := len(nodeCommand.Args) - 1
				gotUnexpectedInvoke = &argc
			}
		}

		if gotUnexpectedInvoke != nil {
			handler.OnUnexpectedInvoke(fset, token.Pos(nodeCommand.Pos), funcname, *gotUnexpectedInvoke)
		}

	default:
	}
}

func (handler Handler) handleTemplateMsgid(fset *token.FileSet, node tmplParser.Node) {
	// the column numbers are a bit "off", but much better than nothing
	pos := token.Pos(node.Position())

	switch node.Type() {
	case tmplParser.NodeString:
		nodeString := node.(*tmplParser.StringNode)
		// found interesting strings
		handler.OnMsgid(fset, pos, nodeString.Text, false)

	case tmplParser.NodePipe:
		nodePipe := node.(*tmplParser.PipeNode)
		handler.handleTemplatePipeNode(fset, nodePipe)

		if len(nodePipe.Cmds) == 0 {
			handler.OnWarning(fset, pos, fmt.Sprintf("unsupported invocation of locate function (no commands): %s", node.String()))
		} else if len(nodePipe.Cmds) != 1 {
			handler.OnWarning(fset, pos, fmt.Sprintf("unsupported invocation of locate function (too many commands): %s", node.String()))
			return
		}
		nodeCommand := nodePipe.Cmds[0]
		if len(nodeCommand.Args) < 2 {
			handler.OnWarning(fset, pos, fmt.Sprintf("unsupported invocation of locate function (not enough arguments): %s", node.String()))
			return
		}

		nodeIdent, ok := nodeCommand.Args[0].(*tmplParser.IdentifierNode)
		if !ok || (nodeIdent.Ident != "print" && nodeIdent.Ident != "printf") {
			// handler.OnWarning(fset, pos, fmt.Sprintf("unsupported invocation of locate function (bad command): %s", node.String()))
			return
		}

		nodeString, ok := nodeCommand.Args[1].(*tmplParser.StringNode)
		if !ok {
			//handler.OnWarning(
			//	fset,
			//	pos,
			//	fmt.Sprintf("unsupported invocation of locate function (string should be first argument to %s): %s", nodeIdent.Ident, node.String()),
			//)
			return
		}

		msgidPrefix := nodeString.Text
		stringPos := token.Pos(nodeString.Pos)

		if len(nodeCommand.Args) == 2 {
			// found interesting strings
			handler.OnMsgid(fset, stringPos, msgidPrefix, false)
		} else {
			if nodeIdent.Ident == "printf" {
				parts := strings.SplitN(msgidPrefix, "%", 2)
				if len(parts) != 2 {
					handler.OnWarning(
						fset,
						stringPos,
						fmt.Sprintf("unsupported invocation of locate function (format string doesn't match \"prefix%%smth\" pattern): %s", nodeString.String()),
					)
					return
				}
				msgidPrefix = parts[0]
			}

			msgidPrefixFin, truncated := PrepareMsgidPrefix(msgidPrefix)
			if truncated {
				handler.OnWarning(fset, stringPos, fmt.Sprintf("needed to truncate message id prefix: %s", msgidPrefix))
			}

			// found interesting strings
			handler.OnMsgidPrefix(fset, stringPos, msgidPrefixFin, truncated)
		}

	default:
		// handler.OnWarning(fset, pos, fmt.Sprintf("unknown invocation of locate function: %s", node.String()))
	}
}

func (handler Handler) handleTemplatePipeNode(fset *token.FileSet, pipeNode *tmplParser.PipeNode) {
	if pipeNode == nil {
		return
	}

	// NOTE: we can't pass `pipeNode.Cmds` to handleTemplateFileNodes due to incompatible argument types
	for _, node := range pipeNode.Cmds {
		handler.handleTemplateNode(fset, node)
	}
}

func (handler Handler) handleTemplateBranchNode(fset *token.FileSet, branchNode tmplParser.BranchNode) {
	handler.handleTemplatePipeNode(fset, branchNode.Pipe)
	handler.handleTemplateFileNodes(fset, branchNode.List.Nodes)
	if branchNode.ElseList != nil {
		handler.handleTemplateFileNodes(fset, branchNode.ElseList.Nodes)
	}
}

func (handler Handler) handleTemplateFileNodes(fset *token.FileSet, nodes []tmplParser.Node) {
	for _, node := range nodes {
		handler.handleTemplateNode(fset, node)
	}
}

// the `Handle*File` functions follow the following calling convention:
// * `fname` is the name of the input file
// * `src` is either `nil` (then the function invokes `ReadFile` to read the file)
//   or the contents of the file as {`[]byte`, or a `string`}

func (handler Handler) HandleTemplateFile(fname string, src any) error {
	var tmplContent []byte
	switch src2 := src.(type) {
	case nil:
		var err error
		tmplContent, err = os.ReadFile(fname)
		if err != nil {
			return LocatedError{
				Location: fname,
				Kind:     "ReadFile",
				Err:      err,
			}
		}
	case []byte:
		tmplContent = src2
	case string:
		// SAFETY: we do not modify tmplContent below
		tmplContent = util.UnsafeStringToBytes(src2)
	default:
		panic("invalid type for 'src'")
	}

	fset := token.NewFileSet()
	fset.AddFile(fname, 1, len(tmplContent)).SetLinesForContent(tmplContent)
	// SAFETY: we do not modify tmplContent2 below
	tmplContent2 := util.UnsafeBytesToString(tmplContent)

	tmpl := template.New(fname)
	tmpl.Funcs(fjTemplates.NewFuncMap())
	tmplParsed, err := tmpl.Parse(tmplContent2)
	if err != nil {
		return LocatedError{
			Location: fname,
			Kind:     "Template parser",
			Err:      err,
		}
	}
	handler.handleTemplateFileNodes(fset, tmplParsed.Root.Nodes)
	return nil
}
