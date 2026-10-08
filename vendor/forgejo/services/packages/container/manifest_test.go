// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package container

import (
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	packages_module "forgejo.org/modules/packages"
	packages_service "forgejo.org/services/packages"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SaveAndGetManifestAndBlob(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	user2 := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	manifestMediaType := "application/vnd.docker.distribution.manifest.v2+json"
	image := "test"
	tag := "latest"

	mci, err := NewManifestCreationInfo(user2, user2, manifestMediaType, image, tag)
	require.NoError(t, err)

	blobContent, _ := base64.StdEncoding.DecodeString(`H4sIAAAJbogA/2IYBaNgFIxYAAgAAP//Lq+17wAEAAA=`)

	blobReader := &io.LimitedReader{R: strings.NewReader(string(blobContent)), N: int64(MaxManifestSize + 1)}
	blobBuf, err := packages_module.CreateHashedBufferFromReaderWithSize(blobReader, MaxManifestSize+1)
	blobDigest := DigestFromHashSummer(blobBuf)
	require.NoError(t, err)
	defer blobBuf.Close()

	blobPCI := &packages_service.PackageCreationInfo{
		PackageInfo: packages_service.PackageInfo{
			Owner: user2,
			Name:  image,
		},
		Creator: user2,
	}

	_, err = SaveAsPackageBlob(t.Context(), blobBuf, blobPCI)
	require.NoError(t, err)

	configDigest := "sha256:4607e093bec406eaadb6f3a340f63400c9d3a7038680744c406903766b938f0d"
	configContent := `{"architecture":"amd64","config":{"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/true"],"ArgsEscaped":true,"Image":"sha256:9bd8b88dc68b80cffe126cc820e4b52c6e558eb3b37680bfee8e5f3ed7b8c257"},"container":"b89fe92a887d55c0961f02bdfbfd8ac3ddf66167db374770d2d9e9fab3311510","container_config":{"Hostname":"b89fe92a887d","Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\"/true\"]"],"ArgsEscaped":true,"Image":"sha256:9bd8b88dc68b80cffe126cc820e4b52c6e558eb3b37680bfee8e5f3ed7b8c257"},"created":"2022-01-01T00:00:00.000000000Z","docker_version":"20.10.12","history":[{"created":"2022-01-01T00:00:00.000000000Z","created_by":"/bin/sh -c #(nop) COPY file:0e7589b0c800daaf6fa460d2677101e4676dd9491980210cb345480e513f3602 in /true "},{"created":"2022-01-01T00:00:00.000000001Z","created_by":"/bin/sh -c #(nop)  CMD [\"/true\"]","empty_layer":true}],"os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:0ff3b91bdf21ecdf2f2f3d4372c2098a14dbe06cd678e8f0a85fd4902d00e2e2"]}}`

	cfgReader := &io.LimitedReader{R: strings.NewReader(configContent), N: int64(MaxManifestSize + 1)}
	cfgBuf, err := packages_module.CreateHashedBufferFromReaderWithSize(cfgReader, MaxManifestSize+1)
	require.NoError(t, err)
	defer cfgBuf.Close()

	confPci := &packages_service.PackageCreationInfo{
		PackageInfo: packages_service.PackageInfo{
			Owner:   user2,
			Name:    image,
			Version: configDigest,
		},
		Creator: user2,
	}

	_, err = SaveAsPackageBlob(t.Context(), cfgBuf, confPci)
	require.NoError(t, err)

	manifestDigest := "sha256:4f10484d1c1bb13e3956b4de1cd42db8e0f14a75be1617b60f2de3cd59c803c6"
	manifestContent := `{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","digest":"sha256:4607e093bec406eaadb6f3a340f63400c9d3a7038680744c406903766b938f0d","size":1069},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","digest":"sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4","size":32}]}`

	reader := &io.LimitedReader{R: strings.NewReader(manifestContent), N: int64(MaxManifestSize + 1)}
	buf, err := packages_module.CreateHashedBufferFromReaderWithSize(reader, MaxManifestSize+1)
	require.NoError(t, err)
	defer buf.Close()

	digest, err := ProcessManifest(t.Context(), *mci, buf)
	require.NoError(t, err)
	assert.Equal(t, digest, manifestDigest)

	pdf, err := GetLocalManifest(t.Context(), user2.ID, image, tag)
	require.NoError(t, err)
	assert.Equal(t, "sha256:"+pdf.Blob.HashSHA256, digest)

	pdf, err = GetLocalBlob(t.Context(), user2.ID, blobDigest, image)
	require.NoError(t, err)
	assert.Equal(t, "sha256:"+pdf.Blob.HashSHA256, blobDigest)

	tl, v, err := GetLocalTagList(t.Context(), user2.LowerName, image, "", 1, user2.ID)
	require.NoError(t, err)
	assert.Len(t, tl.Tags, 1)
	assert.Equal(t, "latest", v.Get("last"))
}
