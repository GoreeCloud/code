// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import "time"

// Secret represents a secret
// swagger:model
type Secret struct {
	// the secret's name
	Name string `json:"name"`
	// swagger:strfmt date-time
	Created time.Time `json:"created_at"`
}

// CreateOrUpdateSecretOption defines the properties of the secret to create or update.
// swagger:model
type CreateOrUpdateSecretOption struct {
	// Data of the secret. Special characters will be retained. Line endings will be normalized to LF to match the
	// behaviour of browsers. Encode the data with Base64 if line endings should be retained.
	//
	// required: true
	Data string `json:"data" binding:"Required"`
}
