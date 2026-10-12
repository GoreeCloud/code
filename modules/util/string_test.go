// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSnakeCase(t *testing.T) {
	cases := map[string]string{
		// all old cases from the legacy package
		"HTTPServer":         "http_server",
		"_camelCase":         "_camel_case",
		"NoHTTPS":            "no_https",
		"Wi_thF":             "wi_th_f",
		"_AnotherTES_TCaseP": "_another_tes_t_case_p",
		"ALL":                "all",
		"_HELLO_WORLD_":      "_hello_world_",
		"HELLO_WORLD":        "hello_world",
		"HELLO____WORLD":     "hello____world",
		"TW":                 "tw",
		"_C":                 "_c",

		"  sentence case  ": "__sentence_case__",
		" Mixed-hyphen case _and SENTENCE_case and UPPER-case": "_mixed_hyphen_case__and_sentence_case_and_upper_case",

		// new cases
		" ":        "_",
		"A":        "a",
		"A0":       "a0",
		"a0":       "a0",
		"Aa0":      "aa0",
		"啊":        "啊",
		"A啊":       "a啊",
		"Aa啊b":     "aa啊b",
		"A啊B":      "a啊_b",
		"Aa啊B":     "aa啊_b",
		"TheCase2": "the_case2",
		"ObjIDs":   "obj_i_ds", // the strange database column name which already exists
	}
	for input, expected := range cases {
		assert.Equal(t, expected, ToSnakeCase(input))
	}
}

func TestASCIIEqualFold(t *testing.T) {
	cases := map[string]struct {
		First    string
		Second   string
		Expected bool
	}{
		"Empty String":          {First: "", Second: "", Expected: true},
		"Single Letter Ident":   {First: "h", Second: "h", Expected: true},
		"Single Letter Equal":   {First: "h", Second: "H", Expected: true},
		"Single Letter Unequal": {First: "h", Second: "g", Expected: false},
		"Simple Match Ident":    {First: "someString", Second: "someString", Expected: true},
		"Simple Match Equal":    {First: "someString", Second: "someSTRIng", Expected: true},
		"Simple Match Unequal":  {First: "someString", Second: "sameString", Expected: false},
		"Different Length":      {First: "abcdef", Second: "abcdefg", Expected: false},
		"Unicode Kelvin":        {First: "ghijklm", Second: "GHIJ\u212ALM", Expected: false},
	}

	for name := range cases {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			Actual := ASCIIEqualFold(c.First, c.Second)
			assert.Equal(t, c.Expected, Actual)
		})
	}
}

func TestRemoveAllStr(t *testing.T) {
	for name, c := range map[string]struct {
		str, res string
		prefix   bool
		all      []string
	}{
		"Empty": {
			str:    "",
			res:    "",
			all:    []string{"hello"},
			prefix: false,
		},
		"Exact": {
			str:    "hello",
			res:    "",
			all:    []string{"hello"},
			prefix: false,
		},
		"One of Two": {
			str:    "hello world",
			res:    "world",
			all:    []string{"hello"},
			prefix: false,
		},
		"Prefix": {
			str:    "is:open is:closed hello",
			res:    "hello",
			all:    []string{"is:"},
			prefix: true,
		},
		"Prefix Multiple": {
			str:    "is:open is:closed hello has:fun",
			res:    "hello",
			all:    []string{"is:", "has:"},
			prefix: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.res, RemoveAllStr(c.str, c.prefix, c.all...))
		})
	}
}
