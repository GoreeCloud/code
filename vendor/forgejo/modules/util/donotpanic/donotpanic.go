// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package donotpanic

import (
	"fmt"

	"forgejo.org/modules/log"
)

type FuncWithError func() error

func SafeFuncWithError(fun FuncWithError) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC recovered: %v\nStacktrace: %s", r, log.Stack(2))
			rErr, ok := r.(error)
			if ok {
				err = fmt.Errorf("PANIC recover with error: %w", rErr)
			} else {
				err = fmt.Errorf("PANIC recover: %v", r)
			}
		}
	}()

	return fun()
}
