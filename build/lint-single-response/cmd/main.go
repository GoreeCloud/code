// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	singleresponse "forgejo.org/build/lint-single-response"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(singleresponse.Analyzer)
}
