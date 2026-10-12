// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package pypi

// Metadata represents the metadata of a PyPI package
type Metadata struct {
	Author          string `json:"author,omitempty"`
	Description     string `json:"description,omitempty"`
	LongDescription string `json:"long_description,omitempty"`
	Summary         string `json:"summary,omitempty"`
	ProjectURL      string `json:"project_url,omitempty"`
	License         string `json:"license,omitempty"`
	RequiresPython  string `json:"requires_python,omitempty"`
}

type FileHashesJSON struct {
	SHA256 string `json:"sha256"`
}

type FileJSON struct {
	Filename       string         `json:"filename"`
	URL            string         `json:"url"`
	Hashes         FileHashesJSON `json:"hashes"`
	RequiresPython string         `json:"requires-python"`
	Size           int64          `json:"size"`
}

type PackageMetaJSON struct {
	APIVersion string `json:"api-version"`
}

type PackageJSON struct {
	Name     string          `json:"name"`
	Meta     PackageMetaJSON `json:"meta"`
	Versions []string        `json:"versions"`
	Files    []FileJSON      `json:"files"`
}
