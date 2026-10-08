// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package federation

import (
	"fmt"
	"net/http"
)

type ErrNotAcceptable struct {
	Message string
}

func NewErrNotAcceptablef(format string, a ...any) ErrNotAcceptable {
	message := fmt.Sprintf(format, a...)
	return ErrNotAcceptable{Message: message}
}

func (err ErrNotAcceptable) Error() string {
	return fmt.Sprintf("NotAcceptable: %v", err.Message)
}

type ErrInternal struct {
	Message string
}

func NewErrInternalf(format string, a ...any) ErrInternal {
	message := fmt.Sprintf(format, a...)
	return ErrInternal{Message: message}
}

func (err ErrInternal) Error() string {
	return fmt.Sprintf("InternalServerError: %v", err.Message)
}

func HTTPStatus(err error) int {
	switch err.(type) {
	case ErrNotAcceptable:
		return http.StatusNotAcceptable
	default:
		return http.StatusInternalServerError
	}
}
