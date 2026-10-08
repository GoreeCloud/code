// Copyright 2015 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlob_Data(t *testing.T) {
	output := "file2\n"
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	repo, err := openRepositoryWithDefaultContext(bareRepo1Path)
	require.NoError(t, err)

	defer repo.Close()

	testBlob, err := repo.GetBlob("6c493ff740f9380390d5c9ddef4af18697ac9375")
	require.NoError(t, err)

	r, err := testBlob.DataAsync()
	require.NoError(t, err)
	require.NotNil(t, r)

	data, err := io.ReadAll(r)
	require.NoError(t, r.Close())

	require.NoError(t, err)
	assert.Equal(t, output, string(data))
}

func TestBlob(t *testing.T) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	repo, err := openRepositoryWithDefaultContext(bareRepo1Path)
	require.NoError(t, err)

	defer repo.Close()

	testBlob, err := repo.GetBlob("6c493ff740f9380390d5c9ddef4af18697ac9375")
	require.NoError(t, err)

	t.Run("GetContentBase64", func(t *testing.T) {
		r, err := testBlob.GetContentBase64(100)
		require.NoError(t, err)
		require.Equal(t, "ZmlsZTIK", r)

		r, err = testBlob.GetContentBase64(-1)
		require.ErrorAs(t, err, &BlobTooLargeError{})
		require.Empty(t, r)

		r, err = testBlob.GetContentBase64(4)
		require.ErrorAs(t, err, &BlobTooLargeError{})
		require.Empty(t, r)

		r, err = testBlob.GetContentBase64(6)
		require.NoError(t, err)
		require.Equal(t, "ZmlsZTIK", r)
	})

	t.Run("NewTruncatedReader", func(t *testing.T) {
		// read fewer than available
		rc, size, err := testBlob.NewTruncatedReader(100)
		require.NoError(t, err)
		require.Equal(t, int64(6), size)

		buf := make([]byte, 1)
		n, err := rc.Read(buf)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		require.Equal(t, "f", string(buf))
		n, err = rc.Read(buf)
		require.NoError(t, err)
		require.Equal(t, 1, n)
		require.Equal(t, "i", string(buf))

		require.NoError(t, rc.Close())

		// read more than available
		rc, size, err = testBlob.NewTruncatedReader(100)
		require.NoError(t, err)
		require.Equal(t, int64(6), size)

		buf = make([]byte, 100)
		n, err = rc.Read(buf)
		require.NoError(t, err)
		require.Equal(t, 6, n)
		require.Equal(t, "file2\n", string(buf[:n]))

		n, err = rc.Read(buf)
		require.Error(t, err)
		require.Equal(t, io.EOF, err)
		require.Equal(t, 0, n)

		require.NoError(t, rc.Close())

		// read more than truncated
		rc, size, err = testBlob.NewTruncatedReader(4)
		require.NoError(t, err)
		require.Equal(t, int64(6), size)

		buf = make([]byte, 10)
		n, err = rc.Read(buf)
		require.NoError(t, err)
		require.Equal(t, 4, n)
		require.Equal(t, "file", string(buf[:n]))

		n, err = rc.Read(buf)
		require.Error(t, err)
		require.Equal(t, io.EOF, err)
		require.Equal(t, 0, n)

		require.NoError(t, rc.Close())
	})

	t.Run("NonExisting", func(t *testing.T) {
		nonExistingBlob, err := repo.GetBlob("00003ff740f9380390d5c9ddef4af18690000000")
		require.NoError(t, err)

		rc, size, err := nonExistingBlob.NewTruncatedReader(100)
		require.Error(t, err)
		require.IsType(t, ErrNotExist{}, err)
		require.Empty(t, rc)
		require.Empty(t, size)
	})
}

func Benchmark_Blob_Data(b *testing.B) {
	bareRepo1Path := filepath.Join(testReposDir, "repo1_bare")
	repo, err := openRepositoryWithDefaultContext(bareRepo1Path)
	if err != nil {
		b.Fatal(err)
	}
	defer repo.Close()

	testBlob, err := repo.GetBlob("6c493ff740f9380390d5c9ddef4af18697ac9375")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		r, err := testBlob.DataAsync()
		if err != nil {
			b.Fatal(err)
		}
		io.ReadAll(r)
		_ = r.Close()
	}
}
