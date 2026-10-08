// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package cmd

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func Test_installSignals(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("Windows does not terminate in an awaitable manner")
		return
	}

	for _, s := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		t.Run(fmt.Sprintf("Context is terminated on %s", s), func(t *testing.T) {
			// Register the signal handler. context.Background() is chosen deliberately,
			// because unlike t.Context(), we can be sure that it's not cancelled by a
			// different handler.
			ctx, cancel := installSignals(context.Background())
			t.Cleanup(cancel)

			// Send the signal in the background.
			go syscall.Kill(syscall.Getpid(), s)

			select {
			case <-time.Tick(time.Second * 10):
				t.Fatalf("Context not cancelled via signal after 10 seconds")
			case <-ctx.Done():
				t.Logf("Context was cancelled")
			}
		})
	}
}
