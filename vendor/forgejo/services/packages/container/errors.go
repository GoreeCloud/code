// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package container

import (
	"net/http"
)

// https://github.com/opencontainers/distribution-spec/blob/main/spec.md#error-codes
var (
	ErrBlobUnknown         = &NamedError{Code: "BLOB_UNKNOWN", StatusCode: http.StatusNotFound}
	ErrBlobUploadInvalid   = &NamedError{Code: "BLOB_UPLOAD_INVALID", StatusCode: http.StatusBadRequest}
	ErrBlobUploadUnknown   = &NamedError{Code: "BLOB_UPLOAD_UNKNOWN", StatusCode: http.StatusNotFound}
	ErrDigestInvalid       = &NamedError{Code: "DIGEST_INVALID", StatusCode: http.StatusBadRequest}
	ErrManifestBlobUnknown = &NamedError{Code: "MANIFEST_BLOB_UNKNOWN", StatusCode: http.StatusNotFound}
	ErrManifestInvalid     = &NamedError{Code: "MANIFEST_INVALID", StatusCode: http.StatusBadRequest}
	ErrManifestUnknown     = &NamedError{Code: "MANIFEST_UNKNOWN", StatusCode: http.StatusNotFound}
	ErrNameInvalid         = &NamedError{Code: "NAME_INVALID", StatusCode: http.StatusBadRequest}
	ErrNameUnknown         = &NamedError{Code: "NAME_UNKNOWN", StatusCode: http.StatusNotFound}
	ErrSizeInvalid         = &NamedError{Code: "SIZE_INVALID", StatusCode: http.StatusBadRequest}
	ErrUnauthorized        = &NamedError{Code: "UNAUTHORIZED", StatusCode: http.StatusUnauthorized}
	ErrUnsupported         = &NamedError{Code: "UNSUPPORTED", StatusCode: http.StatusNotImplemented}
)

type NamedError struct {
	Code       string
	StatusCode int
	Message    string
}

func (e *NamedError) Error() string {
	return e.Message
}

// WithMessage creates a new instance of the error with a different message
func (e *NamedError) WithMessage(message string) *NamedError {
	return &NamedError{
		Code:       e.Code,
		StatusCode: e.StatusCode,
		Message:    message,
	}
}

// WithStatusCode creates a new instance of the error with a different status code
func (e *NamedError) WithStatusCode(statusCode int) *NamedError {
	return &NamedError{
		Code:       e.Code,
		StatusCode: statusCode,
		Message:    e.Message,
	}
}
