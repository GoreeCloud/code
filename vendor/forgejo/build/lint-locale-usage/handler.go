// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package lintLocaleUsage

import (
	"go/token"
	"strings"
)

type LocatedError struct {
	Location string
	Kind     string
	Err      error
}

func (e LocatedError) Error() string {
	var sb strings.Builder

	sb.WriteString(e.Location)
	sb.WriteString(":\t")
	if e.Kind != "" {
		sb.WriteString(e.Kind)
		sb.WriteString(": ")
	}
	sb.WriteString("ERROR: ")
	sb.WriteString(e.Err.Error())

	return sb.String()
}

func InitLocaleTrFunctions() map[string][]uint {
	ret := make(map[string][]uint)

	f0 := []uint{0}
	ret["Tr"] = f0
	ret["TrString"] = f0
	ret["TrHTML"] = f0

	ret["TrPluralString"] = []uint{1}
	ret["TrN"] = []uint{1, 2}

	return ret
}

type Handler struct {
	OnMsgid            func(fset *token.FileSet, pos token.Pos, msgid string, weak bool)
	OnMsgidPrefix      func(fset *token.FileSet, pos token.Pos, msgidPrefix string, truncated bool)
	OnUnexpectedInvoke func(fset *token.FileSet, pos token.Pos, funcname string, argc int)
	OnWarning          func(fset *token.FileSet, pos token.Pos, msg string)
	LocaleTrFunctions  map[string][]uint
}

// Truncating a message id prefix to the last dot
func PrepareMsgidPrefix(s string) (string, bool) {
	index := strings.LastIndexByte(s, 0x2e)
	if index == -1 {
		return "", true
	}
	return s[:index], index != len(s)-1
}
