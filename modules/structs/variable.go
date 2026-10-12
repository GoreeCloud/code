// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

// CreateVariableOption defines the properties of the variable to create.
// swagger:model
type CreateVariableOption struct {
	// Value of the variable to create. Special characters will be retained. Line endings will be normalized to LF to
	// match the behaviour of browsers. Encode the data with Base64 if line endings should be retained.
	//
	// required: true
	Value string `json:"value" binding:"Required"`
}

// UpdateVariableOption defines the properties of the variable to update.
// swagger:model
type UpdateVariableOption struct {
	// New name for the variable. If the field is empty, the variable name won't be updated. Forgejo will convert it to
	// uppercase.
	Name string `json:"name"`
	// Value of the variable to update. Special characters will be retained. Line endings will be normalized to LF to
	// match the behaviour of browsers. Encode the data with Base64 if line endings should be retained.
	//
	// required: true
	Value string `json:"value" binding:"Required"`
}

// ActionVariable return value of the query API
// swagger:model
type ActionVariable struct {
	// the owner to which the variable belongs
	OwnerID int64 `json:"owner_id"`
	// the repository to which the variable belongs
	RepoID int64 `json:"repo_id"`
	// the name of the variable
	Name string `json:"name"`
	// the value of the variable
	Data string `json:"data"`
}
