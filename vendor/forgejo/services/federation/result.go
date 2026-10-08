// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package federation

import "github.com/go-ap/activitypub"

type ServiceResult struct {
	HTTPStatus   int
	Bytes        []byte
	Activity     activitypub.Activity
	withBytes    bool
	withActivity bool
	statusOnly   bool
}

func NewServiceResultStatusOnly(status int) ServiceResult {
	return ServiceResult{HTTPStatus: status, statusOnly: true}
}

func NewServiceResultWithBytes(status int, bytes []byte) ServiceResult {
	return ServiceResult{HTTPStatus: status, Bytes: bytes, withBytes: true}
}

func (serviceResult ServiceResult) WithBytes() bool {
	return serviceResult.withBytes
}

func (serviceResult ServiceResult) WithActivity() bool {
	return serviceResult.withActivity
}

func (serviceResult ServiceResult) StatusOnly() bool {
	return serviceResult.statusOnly
}
