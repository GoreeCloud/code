// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package method

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenFromAuthorizationBearer(t *testing.T) {
	cases := map[string]struct {
		Header        string
		ExpectedToken string
		Expected      bool
	}{
		"Token Uppercase":   {Header: "Token 1234567890123456789012345687901325467890", ExpectedToken: "1234567890123456789012345687901325467890", Expected: true},
		"Token Lowercase":   {Header: "token 1234567890123456789012345687901325467890", ExpectedToken: "1234567890123456789012345687901325467890", Expected: true},
		"Token Unicode":     {Header: "to\u212Aen 1234567890123456789012345687901325467890", ExpectedToken: "", Expected: false},
		"Bearer Uppercase":  {Header: "Bearer 1234567890123456789012345687901325467890", ExpectedToken: "1234567890123456789012345687901325467890", Expected: true},
		"Bearer Lowercase":  {Header: "bearer 1234567890123456789012345687901325467890", ExpectedToken: "1234567890123456789012345687901325467890", Expected: true},
		"Missing type":      {Header: "1234567890123456789012345687901325467890", ExpectedToken: "", Expected: false},
		"Three Parts":       {Header: "abc 1234567890 test", ExpectedToken: "", Expected: false},
		"Token Three Parts": {Header: "Token 1234567890 test", ExpectedToken: "", Expected: false},
	}

	for name := range cases {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Add("Authorization", c.Header)
			maybeToken := tokenFromAuthorizationBearer(req)
			if hasToken, token := maybeToken.Get(); hasToken {
				assert.True(t, c.Expected)
				assert.Equal(t, c.ExpectedToken, token)
			} else {
				assert.False(t, c.Expected)
			}
		})
	}
}
