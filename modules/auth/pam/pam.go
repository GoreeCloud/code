// Copyright 2014 The Gogs Authors. All rights reserved.
// SPDX-License-Identifier: MIT

//go:build pam

package pam

import (
	"errors"
	"fmt"

	"github.com/msteinert/pam/v2"
)

// ErrInvalidCredentials is returned when PAM reports an authentication
// or account error (wrong password, unknown user, expired account, etc.).
var ErrInvalidCredentials = errors.New("invalid PAM credentials")

// Supported is true when built with PAM
var Supported = true

// Auth pam auth service
func Auth(serviceName, userName, passwd string) (string, error) {
	t, err := pam.StartFunc(serviceName, userName, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return passwd, nil
		case pam.PromptEchoOn, pam.ErrorMsg, pam.TextInfo:
			return "", nil
		}
		return "", errors.New("Unrecognized PAM message style")
	})
	if err != nil {
		return "", err
	}
	defer t.End()

	if err = t.Authenticate(0); err != nil {
		if errors.Is(err, pam.ErrAuth) || errors.Is(err, pam.ErrUserUnknown) {
			return "", fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
		}
		return "", err
	}

	if err = t.AcctMgmt(0); err != nil {
		if errors.Is(err, pam.ErrAcctExpired) || errors.Is(err, pam.ErrPermDenied) {
			return "", fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
		}
		return "", err
	}

	// PAM login names might suffer transformations in the PAM stack.
	// We should take whatever the PAM stack returns for it.
	return t.GetItem(pam.User)
}
