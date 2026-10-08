// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swift

import (
	"archive/zip"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"forgejo.org/modules/json"
	"forgejo.org/modules/util"

	"github.com/hashicorp/go-version"
)

var (
	ErrMissingManifestFile    = util.NewInvalidArgumentErrorf("Package.swift file is missing")
	ErrManifestFileTooLarge   = util.NewInvalidArgumentErrorf("Package.swift file is too large")
	ErrInvalidManifestVersion = util.NewInvalidArgumentErrorf("manifest version is invalid")

	manifestPattern     = regexp.MustCompile(`\APackage(?:@swift-(\d+(?:\.\d+)?(?:\.\d+)?))?\.swift\z`)
	toolsVersionPattern = regexp.MustCompile(`\A// swift-tools-version:(\d+(?:\.\d+)?(?:\.\d+)?)`)
)

const (
	maxManifestFileSize = 128 * 1024

	PropertyScope         = "swift.scope"
	PropertyName          = "swift.name"
	PropertyRepositoryURL = "swift.repository_url"
)

// Package represents a Swift package
type Package struct {
	Manifests map[string]*Manifest
	Metadata  *PackageRelease
}

// Manifest represents a Package.swift file
type Manifest struct {
	Content      string `json:"content"`
	ToolsVersion string `json:"tools_version,omitempty"`
}

// https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/registryserverspecification/#PackageRelease-type
type PackageRelease struct {
	Author                  *Author  `json:"author,omitempty"`
	Description             string   `json:"description,omitempty"`
	LicenseURL              string   `json:"licenseURL,omitempty"`
	OriginalPublicationTime string   `json:"originalPublicationTime,omitempty"`
	ReadmeURL               string   `json:"readmeURL,omitempty"`
	RepositoryURLs          []string `json:"repositoryURLs,omitempty"`
}

// https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/registryserverspecification/#Author-type
type Author struct {
	Name         string        `json:"name,omitempty"`
	Email        string        `json:"email,omitempty"`
	Description  string        `json:"description,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
	URL          string        `json:"url,omitempty"`
}

// https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/registryserverspecification/#Organization-type
type Organization struct {
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// ParsePackage parses the Swift package upload
func ParsePackage(sr io.ReaderAt, size int64, mr io.Reader) (*Package, error) {
	zr, err := zip.NewReader(sr, size)
	if err != nil {
		return nil, err
	}

	p := &Package{
		Manifests: make(map[string]*Manifest),
		Metadata:  &PackageRelease{},
	}

	for _, file := range zr.File {
		manifestMatch := manifestPattern.FindStringSubmatch(path.Base(file.Name))
		if len(manifestMatch) == 0 {
			continue
		}

		if file.UncompressedSize64 > maxManifestFileSize {
			return nil, ErrManifestFileTooLarge
		}

		f, err := zr.Open(file.Name)
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(f)

		if err := f.Close(); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		swiftVersion := ""
		if len(manifestMatch) == 2 && manifestMatch[1] != "" {
			v, err := version.NewSemver(manifestMatch[1])
			if err != nil {
				return nil, ErrInvalidManifestVersion
			}
			swiftVersion = TrimmedVersionString(v)
		}

		manifest := &Manifest{
			Content: string(content),
		}

		toolsMatch := toolsVersionPattern.FindStringSubmatch(manifest.Content)
		if len(toolsMatch) == 2 {
			v, err := version.NewSemver(toolsMatch[1])
			if err != nil {
				return nil, ErrInvalidManifestVersion
			}

			manifest.ToolsVersion = TrimmedVersionString(v)
		}

		p.Manifests[swiftVersion] = manifest
	}

	if _, found := p.Manifests[""]; !found {
		return nil, ErrMissingManifestFile
	}

	if mr != nil {
		var pr *PackageRelease
		if err := json.NewDecoder(mr).Decode(&pr); err != nil {
			return nil, err
		}
		if pr.Author != nil {
			if pr.Author.Name == "" {
				return nil, util.NewInvalidArgumentErrorf("if metadata.author exists, its name can't be empty")
			}
			if pr.Author.Organization != nil && pr.Author.Organization.Name == "" {
				return nil, util.NewInvalidArgumentErrorf("if metadata.author.organization exists, its name can't be empty")
			}
		}
		p.Metadata = pr
	}

	return p, nil
}

// TrimmedVersionString returns the version string without the patch segment if it is zero
func TrimmedVersionString(v *version.Version) string {
	segments := v.Segments64()

	var b strings.Builder
	fmt.Fprintf(&b, "%d.%d", segments[0], segments[1])
	if segments[2] != 0 {
		fmt.Fprintf(&b, ".%d", segments[2])
	}
	return b.String()
}
