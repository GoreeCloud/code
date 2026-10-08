// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package storage

import (
	"bytes"
	"io"
	"testing"

	"forgejo.org/modules/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type spyCloser struct {
	io.Reader
	closed int
}

func (s *spyCloser) Close() error {
	s.closed++
	return nil
}

var _ io.ReadCloser = &spyCloser{}

func testStorageIterator(t *testing.T, typStr Type, cfg *setting.Storage) {
	l, err := NewStorage(typStr, cfg)
	require.NoError(t, err)

	testFiles := []struct {
		path, content string
		size          int64
	}{
		{"a/1.txt", "a1", -1},
		{"/a/1.txt", "aa1", -1}, // same as above, but with leading slash that will be trim
		{"ab/1.txt", "ab1", 3},
		{"b/1.txt", "b1", 2}, // minio closes when the size is set
		{"b/2.txt", "b2", -1},
		{"b/3.txt", "b3", -1},
		{"b/x 4.txt", "bx4", -1},
	}
	for _, f := range testFiles {
		sc := &spyCloser{bytes.NewBufferString(f.content), 0}
		_, err = l.Save(f.path, sc, f.size)
		require.NoError(t, err)
		assert.Equal(t, 0, sc.closed)
	}

	expectedList := map[string][]string{
		"a":           {"a/1.txt"},
		"b":           {"b/1.txt", "b/2.txt", "b/3.txt", "b/x 4.txt"},
		"":            {"a/1.txt", "b/1.txt", "b/2.txt", "b/3.txt", "b/x 4.txt", "ab/1.txt"},
		"/":           {"a/1.txt", "b/1.txt", "b/2.txt", "b/3.txt", "b/x 4.txt", "ab/1.txt"},
		"a/b/../../a": {"a/1.txt"},
	}
	for dir, expected := range expectedList {
		count := 0
		err = l.IterateObjects(dir, func(path string, f Object) error {
			defer f.Close()
			assert.Contains(t, expected, path)
			count++
			return nil
		})
		require.NoError(t, err)
		assert.Len(t, expected, count)
	}
}
