// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swift

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"forgejo.org/modules/util"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	packageDescription   = "Package Description"
	packageRepositoryURL = "https://gitea.io/gitea/gitea"
	packageAuthor        = "KN4CK3R"
	packageEmail         = "example@example.com"
	packageLicense       = "https://opensource.org/license/mit"
)

func TestParsePackage(t *testing.T) {
	createArchive := func(files map[string][]byte) *bytes.Reader {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		for filename, content := range files {
			w, _ := zw.Create(filename)
			w.Write(content)
		}
		zw.Close()
		return bytes.NewReader(buf.Bytes())
	}

	t.Run("MissingManifestFile", func(t *testing.T) {
		data := createArchive(map[string][]byte{"dummy.txt": {}})

		p, err := ParsePackage(data, data.Size(), nil)
		assert.Nil(t, p)
		require.ErrorIs(t, err, ErrMissingManifestFile)
	})

	t.Run("ManifestFileTooLarge", func(t *testing.T) {
		data := createArchive(map[string][]byte{
			"Package.swift": make([]byte, maxManifestFileSize+1),
		})

		p, err := ParsePackage(data, data.Size(), nil)
		assert.Nil(t, p)
		require.ErrorIs(t, err, ErrManifestFileTooLarge)
	})

	t.Run("WithoutMetadata", func(t *testing.T) {
		content1 := "// swift-tools-version:5.7\n//\n//  Package.swift"
		content2 := "// swift-tools-version:5.6\n//\n//  Package@swift-5.6.swift"

		data := createArchive(map[string][]byte{
			"Package.swift":           []byte(content1),
			"Package@swift-5.5.swift": []byte(content2),
		})

		p, err := ParsePackage(data, data.Size(), nil)
		assert.NotNil(t, p)
		require.NoError(t, err)

		assert.NotNil(t, p.Metadata)
		assert.Empty(t, p.Metadata.RepositoryURLs)
		assert.Len(t, p.Manifests, 2)
		m := p.Manifests[""]
		assert.Equal(t, "5.7", m.ToolsVersion)
		assert.Equal(t, content1, m.Content)
		m = p.Manifests["5.5"]
		assert.Equal(t, "5.6", m.ToolsVersion)
		assert.Equal(t, content2, m.Content)
	})

	t.Run("WithMetadata", func(t *testing.T) {
		data := createArchive(map[string][]byte{
			"Package.swift": []byte("// swift-tools-version:5.7\n//\n//  Package.swift"),
		})

		p, err := ParsePackage(
			data,
			data.Size(),
			strings.NewReader(`{"description":"`+packageDescription+`","licenseURL":"`+packageLicense+`","author":{"name":"`+packageAuthor+`"},"repositoryURLs":["`+packageRepositoryURL+`"]}`),
		)
		assert.NotNil(t, p)
		require.NoError(t, err)

		assert.NotNil(t, p.Metadata)
		assert.Len(t, p.Manifests, 1)
		m := p.Manifests[""]
		assert.Equal(t, "5.7", m.ToolsVersion)

		assert.Equal(t, packageDescription, p.Metadata.Description)
		assert.Equal(t, packageLicense, p.Metadata.LicenseURL)
		assert.Equal(t, packageAuthor, p.Metadata.Author.Name)
		assert.ElementsMatch(t, []string{packageRepositoryURL}, p.Metadata.RepositoryURLs)
	})

	t.Run("WithInvalidMetadata", func(t *testing.T) {
		data := createArchive(map[string][]byte{
			"Package.swift": []byte("// swift-tools-version:5.7\n//\n//  Package.swift"),
		})

		p, err := ParsePackage(
			data,
			data.Size(),
			strings.NewReader(`{"description":"`+packageDescription+`","licenseURL":"`+packageLicense+`","author":{"email":"`+packageEmail+`"},"repositoryURLs":["`+packageRepositoryURL+`"]}`),
		)
		assert.Nil(t, p)
		require.ErrorIs(t, err, util.ErrInvalidArgument)
	})
}

func TestTrimmedVersionString(t *testing.T) {
	cases := []struct {
		Version  *version.Version
		Expected string
	}{
		{
			Version:  version.Must(version.NewVersion("1")),
			Expected: "1.0",
		},
		{
			Version:  version.Must(version.NewVersion("1.0")),
			Expected: "1.0",
		},
		{
			Version:  version.Must(version.NewVersion("1.0.0")),
			Expected: "1.0",
		},
		{
			Version:  version.Must(version.NewVersion("1.0.1")),
			Expected: "1.0.1",
		},
		{
			Version:  version.Must(version.NewVersion("1.0+meta")),
			Expected: "1.0",
		},
		{
			Version:  version.Must(version.NewVersion("1.0.0+meta")),
			Expected: "1.0",
		},
		{
			Version:  version.Must(version.NewVersion("1.0.1+meta")),
			Expected: "1.0.1",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expected, TrimmedVersionString(c.Version))
	}
}
