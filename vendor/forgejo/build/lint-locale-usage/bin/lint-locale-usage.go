// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	llu "forgejo.org/build/lint-locale-usage"
	"forgejo.org/modules/container"
	"forgejo.org/modules/translation/localeiter"
)

// this works by first gathering all valid source string IDs from `en-US` reference files
// and then checking if all used source strings are actually defined

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

type StringTrie interface {
	Matches(key []string) bool
}

type StringTrieMap map[string]StringTrie

func (m StringTrieMap) Matches(key []string) bool {
	if len(key) == 0 || m == nil {
		return true
	}
	value, ok := m[key[0]]
	if !ok {
		return false
	}
	if value == nil {
		return true
	}
	return value.Matches(key[1:])
}

func (m StringTrieMap) Insert(key []string) {
	if m == nil {
		return
	}

	switch len(key) {
	case 0:
		return

	case 1:
		m[key[0]] = nil

	default:
		if value, ok := m[key[0]]; ok {
			if value == nil {
				return
			}
		} else {
			m[key[0]] = make(StringTrieMap)
		}
		m[key[0]].(StringTrieMap).Insert(key[1:])
	}
}

func ParseAllowedMaskedUsages(fname string, usedMsgids container.Set[string], allowedMaskedPrefixes StringTrieMap, chkMsgid func(msgid string) bool) error {
	file, err := os.Open(fname)
	if err != nil {
		return llu.LocatedError{
			Location: fname,
			Kind:     "Open",
			Err:      err,
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lno := 0
	for scanner.Scan() {
		lno++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if linePrefix, found := strings.CutSuffix(line, "."); found {
			allowedMaskedPrefixes.Insert(strings.Split(linePrefix, "."))
		} else {
			if !chkMsgid(line) {
				return llu.LocatedError{
					Location: fmt.Sprintf("%s: line %d", fname, lno),
					Kind:     "undefined msgid",
					Err:      errors.New(line),
				}
			}
			usedMsgids.Add(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return llu.LocatedError{
			Location: fname,
			Kind:     "Scanner",
			Err:      err,
		}
	}
	return nil
}

func Usage() {
	outp := flag.CommandLine.Output()
	fmt.Fprintf(outp, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintf(outp, "\nThis command assumes that it gets started from the project root directory.\n")

	fmt.Fprintf(outp, "\nExit codes:\n")
	for _, i := range []string{
		"0\tsuccess, no issues found",
		"1\tunable to walk directory tree",
		"2\tunable to parse locale ini/json files",
		"3\tunable to parse go or text/template files",
		"4\tfound missing message IDs",
		"5\tfound unused message IDs",
	} {
		fmt.Fprintf(outp, "\t%s\n", i)
	}

	fmt.Fprintf(outp, "\nSpecial Go doc comments:\n")
	for _, i := range []string{
		"//llu:returnsTrKey",
		"\tcan be used in front of functions to indicate",
		"\tthat the function returns message IDs",
		"\tWARNING: this currently doesn't support nested functions properly",
		"",
		"//llu:returnsTrKeySuffix prefix.",
		"\tsimilar to llu:returnsTrKey, but the given prefix is prepended",
		"\tto the found strings before interpreting them as msgids",
		"",
		"// llu:TrKeys",
		"\tcan be used in front of 'const' and 'var' blocks",
		"\tin order to mark all contained strings as message IDs",
		"",
		"// llu:TrKeysSuffix prefix.",
		"\tlike llu:returnsTrKeySuffix, but for 'const' and 'var' blocks",
	} {
		if i == "" {
			fmt.Fprintf(outp, "\n")
		} else {
			fmt.Fprintf(outp, "\t%s\n", i)
		}
	}
}

//nolint:forbidigo
func main() {
	allowMissingMsgids := false
	allowUnusedMsgids := false
	allowWeakMissingMsgids := true
	usedMsgids := make(container.Set[string])
	allowedMaskedPrefixes := make(StringTrieMap)

	// It's possible for execl to hand us an empty os.Args.
	if len(os.Args) == 0 {
		flag.CommandLine = flag.NewFlagSet("lint-locale-usage", flag.ExitOnError)
	} else {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	flag.CommandLine.Usage = Usage
	flag.Usage = Usage

	flag.BoolVar(
		&allowMissingMsgids,
		"allow-missing-msgids",
		false,
		"don't return an error code if missing message IDs are found",
	)
	flag.BoolVar(
		&allowWeakMissingMsgids,
		"allow-weak-missing-msgids",
		true,
		"Don't return an error code if missing 'weak' (e.g. \"form.$msgid\") message IDs are found",
	)
	flag.BoolVar(
		&allowUnusedMsgids,
		"allow-unused-msgids",
		false,
		"don't return an error code if unused message IDs are found",
	)

	msgids := make(container.Set[string])

	localeFile := filepath.Join(filepath.Join("options", "locale"), "locale_en-US.ini")
	localeContent, err := os.ReadFile(localeFile)
	if err != nil {
		fmt.Printf("%s:\tERROR: %s\n", localeFile, err.Error())
		os.Exit(2)
	}

	if err = localeiter.IterateMessagesContent(localeContent, func(trKey, trValue string) error {
		msgids[trKey] = struct{}{}
		return nil
	}); err != nil {
		fmt.Printf("%s:\tERROR: %s\n", localeFile, err.Error())
		os.Exit(2)
	}

	localeFile = filepath.Join(filepath.Join("options", "locale_next"), "locale_en-US.json")
	localeContent, err = os.ReadFile(localeFile)
	if err != nil {
		fmt.Printf("%s:\tERROR: %s\n", localeFile, err.Error())
		os.Exit(2)
	}

	if err := localeiter.IterateMessagesNextContent(localeContent, func(trKey, pluralForm, trValue string) error {
		// ignore plural form
		msgids[trKey] = struct{}{}
		return nil
	}); err != nil {
		fmt.Printf("%s:\tERROR: %s\n", localeFile, err.Error())
		os.Exit(2)
	}

	gotAnyMsgidError := false

	flag.Func(
		"allow-masked-usages-from",
		"supply a file containing a newline-separated list of allowed masked usages",
		func(argval string) error {
			return ParseAllowedMaskedUsages(argval, usedMsgids, allowedMaskedPrefixes, func(msgid string) bool {
				return msgids.Contains(msgid)
			})
		},
	)
	flag.Parse()

	onError := func(err error) {
		if err == nil {
			return
		}
		fmt.Println(err.Error())
		os.Exit(3)
	}

	handler := llu.Handler{
		OnMsgidPrefix: func(fset *token.FileSet, pos token.Pos, msgidPrefix string, truncated bool) {
			msgidPrefixSplit := strings.Split(msgidPrefix, ".")
			if !truncated {
				allowedMaskedPrefixes.Insert(msgidPrefixSplit)
			} else if !allowedMaskedPrefixes.Matches(msgidPrefixSplit) {
				gotAnyMsgidError = true
				fmt.Printf("%s:\tmissing msgid prefix: %s\n", fset.Position(pos).String(), msgidPrefix)
			}
		},
		OnMsgid: func(fset *token.FileSet, pos token.Pos, msgid string, weak bool) {
			if !msgids.Contains(msgid) {
				if weak && allowWeakMissingMsgids {
					return
				}
				gotAnyMsgidError = true
				fmt.Printf("%s:\tmissing msgid: %s\n", fset.Position(pos).String(), msgid)
			} else {
				usedMsgids.Add(msgid)
			}
		},
		OnUnexpectedInvoke: func(fset *token.FileSet, pos token.Pos, funcname string, argc int) {
			gotAnyMsgidError = true
			fmt.Printf("%s:\tunexpected invocation of %s with %d arguments\n", fset.Position(pos).String(), funcname, argc)
		},
		OnWarning: func(fset *token.FileSet, pos token.Pos, msg string) {
			fmt.Printf("%s:\tWARNING: %s\n", fset.Position(pos).String(), msg)
		},
		LocaleTrFunctions: llu.InitLocaleTrFunctions(),
	}

	if err := filepath.WalkDir(".", func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if name == "docker" || name == ".git" || name == "node_modules" {
				return fs.SkipDir
			}
		} else if name == "bindata.go" || fpath == "modules/translation/i18n/i18n_test.go" || fpath == "modules/translation/i18n/i18n_ini_test.go" {
			// skip false positives
		} else if strings.HasSuffix(name, ".go") {
			onError(HandleGoFile(handler, fpath, nil))
		} else if strings.HasSuffix(name, ".tmpl") {
			if strings.HasPrefix(fpath, "tests") && strings.HasSuffix(name, ".ini.tmpl") {
				// skip false positives
			} else {
				onError(handler.HandleTemplateFile(fpath, nil))
			}
		}
		return nil
	}); err != nil {
		fmt.Printf("walkdir ERROR: %s\n", err.Error())
		os.Exit(1)
	}

	unusedMsgids := []string{}

	for msgid := range msgids {
		if !usedMsgids.Contains(msgid) && !allowedMaskedPrefixes.Matches(strings.Split(msgid, ".")) {
			unusedMsgids = append(unusedMsgids, msgid)
		}
	}

	sort.Strings(unusedMsgids)

	if len(unusedMsgids) != 0 {
		fmt.Printf("=== unused msgids (%d): ===\n", len(unusedMsgids))
		for _, msgid := range unusedMsgids {
			fmt.Printf("- %s\n", msgid)
		}
	}

	if !allowMissingMsgids && gotAnyMsgidError {
		os.Exit(4)
	}
	if !allowUnusedMsgids && len(unusedMsgids) != 0 {
		os.Exit(5)
	}
}
