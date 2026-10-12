// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package code

import "forgejo.org/modules/indexer/code/internal"

func HighlightSearchResultCode(filename string, lineNums []int, highlightRanges [][3]int, code string) []internal.ResultLine {
	return internal.HighlightSearchResultCode(filename, lineNums, highlightRanges, code)
}
