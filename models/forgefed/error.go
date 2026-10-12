// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package forgefed

import (
	"fmt"
)

type ErrFederationHostNotFound struct {
	SearchKey   string
	SearchValue string
}

func (err ErrFederationHostNotFound) Error() string {
	return fmt.Sprintf("ErrFederationHostNotFound: search key: %s, search value: %s", err.SearchKey, err.SearchValue)
}

func IsErrFederationHostNotFound(err error) bool {
	_, ok := err.(ErrFederationHostNotFound)
	return ok
}
