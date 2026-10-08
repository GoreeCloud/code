// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package util

import (
	"slices"
	"strings"
	"unsafe"
)

func isSnakeCaseUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}

func isSnakeCaseLowerOrNumber(c byte) bool {
	return 'a' <= c && c <= 'z' || '0' <= c && c <= '9'
}

// ToSnakeCase convert the input string to snake_case format.
//
// Some samples.
//
//	"FirstName"  => "first_name"
//	"HTTPServer" => "http_server"
//	"NoHTTPS"    => "no_https"
//	"GO_PATH"    => "go_path"
//	"GO PATH"    => "go_path"      // space is converted to underscore.
//	"GO-PATH"    => "go_path"      // hyphen is converted to underscore.
func ToSnakeCase(input string) string {
	if len(input) == 0 {
		return ""
	}

	var res []byte
	if len(input) == 1 {
		c := input[0]
		if isSnakeCaseUpper(c) {
			res = []byte{c + 'a' - 'A'}
		} else if isSnakeCaseLowerOrNumber(c) {
			res = []byte{c}
		} else {
			res = []byte{'_'}
		}
	} else {
		res = make([]byte, 0, len(input)*4/3)
		pos := 0
		needSep := false
		for pos < len(input) {
			c := input[pos]
			if c >= 0x80 {
				res = append(res, c)
				pos++
				continue
			}
			isUpper := isSnakeCaseUpper(c)
			if isUpper || isSnakeCaseLowerOrNumber(c) {
				end := pos + 1
				if isUpper {
					// skip the following upper letters
					for end < len(input) && isSnakeCaseUpper(input[end]) {
						end++
					}
					if end-pos > 1 && end < len(input) && isSnakeCaseLowerOrNumber(input[end]) {
						end--
					}
				}
				// skip the following lower or number letters
				for end < len(input) && (isSnakeCaseLowerOrNumber(input[end]) || input[end] >= 0x80) {
					end++
				}
				if needSep {
					res = append(res, '_')
				}
				res = append(res, input[pos:end]...)
				pos = end
				needSep = true
			} else {
				res = append(res, '_')
				pos++
				needSep = false
			}
		}
		for i := 0; i < len(res); i++ {
			if isSnakeCaseUpper(res[i]) {
				res[i] += 'a' - 'A'
			}
		}
	}
	return UnsafeBytesToString(res)
}

// UnsafeBytesToString uses Go's unsafe package to convert a byte slice to a string.
func UnsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// UnsafeStringToBytes uses Go's unsafe package to convert a string to a byte slice.
func UnsafeStringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// AsciiEqualFold is taken from Golang, but reimplemented here, since the original is not exposed to public
// Taken from: https://cs.opensource.google/go/go/+/refs/tags/go1.24.4:src/net/http/internal/ascii/print.go
func ASCIIEqualFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if ASCIILower(s[i]) != ASCIILower(t[i]) {
			return false
		}
	}
	return true
}

// AsciiLower returns the ASCII lowercase version of b.
func ASCIILower(b byte) byte {
	if 'A' <= b && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func RemoveAllStr(s string, prefix bool, all ...string) string {
	if len(s) == 0 {
		return ""
	}

	sb, first := strings.Builder{}, true
	for field := range strings.FieldsSeq(s) {
		if hasAny(field, prefix, all...) {
			continue
		}
		if first {
			first = false
			goto write
		}
		sb.WriteRune(' ')
	write:
		sb.WriteString(field)
	}
	return sb.String()
}

func hasAny(s string, prefix bool, all ...string) bool {
	if !prefix {
		return slices.Contains(all, s)
	}

	for _, field := range all {
		if strings.HasPrefix(s, field) {
			return true
		}
	}
	return false
}
