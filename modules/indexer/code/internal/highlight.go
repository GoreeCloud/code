// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package internal

import (
	"bytes"
	"html/template"
	"strings"

	"forgejo.org/modules/highlight"
	"forgejo.org/services/gitdiff"
)

func WriteStrings(buf *bytes.Buffer, strs ...string) error {
	for _, s := range strs {
		_, err := buf.WriteString(s)
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	highlightTagStart = "<span class=\"search-highlight\">"
	highlightTagEnd   = "</span>"
)

func HighlightSearchResultCode(filename string, lineNums []int, highlightRanges [][3]int, code string) []ResultLine {
	hcd := gitdiff.NewHighlightCodeDiff()
	hcd.CollectUsedRunes(code)
	startTag, endTag := hcd.NextPlaceholder(), hcd.NextPlaceholder()
	hcd.PlaceholderTokenMap[startTag] = highlightTagStart
	hcd.PlaceholderTokenMap[endTag] = highlightTagEnd

	// we should highlight the whole code block first, otherwise it doesn't work well with multiple line highlighting
	hl, _ := highlight.Code(filename, "", code)
	conv := hcd.ConvertToPlaceholders(string(hl))
	convLines := strings.Split(conv, "\n")

	// each highlightRange is of the form [line number, start byte offset, end byte offset]
	for _, highlightRange := range highlightRanges {
		ln, start, end := highlightRange[0], highlightRange[1], highlightRange[2]
		line := convLines[ln]
		if line == "" || len(line) <= start || len(line) < end {
			continue
		}

		sr := strings.NewReader(line)
		sb := strings.Builder{}
		count := -1
		isOpen := false
		for r, size, err := sr.ReadRune(); err == nil; r, size, err = sr.ReadRune() {
			if token, ok := hcd.PlaceholderTokenMap[r];
			// token was not found
			!ok {
				count += size
			} else if
			// token was marked as used
			token == "" ||
				// the token is not an valid html tag emitted by chroma
				!(len(token) > 6 && (token[0:5] == "<span" || token[0:6] == "</span")) {
				count++
			} else if !isOpen {
				// open the tag only after all other placeholders
				sb.WriteRune(r)
				continue
			} else if isOpen && count < end {
				// if the tag is open, but a placeholder exists in between
				// close the tag
				sb.WriteRune(endTag)
				// write the placeholder
				sb.WriteRune(r)
				// reopen the tag
				sb.WriteRune(startTag)
				continue
			}

			switch {
			case count >= end:
				// if tag is not open, no need to close
				if !isOpen {
					break
				}
				sb.WriteRune(endTag)
				isOpen = false
			case count >= start:
				// if tag is open, do not open again
				if isOpen {
					break
				}
				isOpen = true
				sb.WriteRune(startTag)
			}

			sb.WriteRune(r)
		}
		if isOpen {
			sb.WriteRune(endTag)
		}
		convLines[ln] = sb.String()
	}
	conv = strings.Join(convLines, "\n")

	highlightedLines := strings.Split(hcd.Recover(conv), "\n")
	// The lineNums outputted by highlight.Code might not match the original lineNums, because "highlight" removes the last `\n`
	lines := make([]ResultLine, min(len(highlightedLines), len(lineNums)))
	for i := range len(lines) {
		lines[i].Num = lineNums[i]
		lines[i].FormattedContent = template.HTML(highlightedLines[i])
	}
	return lines
}
