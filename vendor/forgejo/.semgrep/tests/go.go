// Copyright 2014 The Gogs Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "net/http/pprof" // Used for debugging if enabled and a web server is running

	"forgejo.org/modules/setting"
)

func setPortEmptyCaseBad(port string) error {
	setting.AppURL = strings.Replace(setting.AppURL, setting.HTTPPort, port, 1)
	setting.HTTPPort = port

	// ruleid:forgejo-switch-empty-case
	switch setting.Protocol {
	case setting.HTTPUnix:
	case setting.FCGI:
	case setting.FCGIUnix:
	default:
		defaultLocalURL := string(setting.Protocol) + "://"
	}

	// ok:forgejo-switch-empty-case
	switch setting.Protocol {
	case setting.HTTPUnix:
		break
	case setting.FCGI:
		break
	case setting.FCGIUnix:
		break
	default:
		defaultLocalURL := string(setting.Protocol) + "://"
	}

	return nil
}
