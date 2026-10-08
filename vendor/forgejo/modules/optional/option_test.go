// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package optional_test

import (
	"testing"

	"forgejo.org/modules/optional"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOption(t *testing.T) {
	var uninitialized optional.Option[int]
	assert.False(t, uninitialized.Has())
	assert.Equal(t, int(0), uninitialized.ValueOrZeroValue())
	assert.Equal(t, int(1), uninitialized.ValueOrDefault(1))

	none := optional.None[int]()
	assert.False(t, none.Has())
	assert.Equal(t, int(0), none.ValueOrZeroValue())
	assert.Equal(t, int(1), none.ValueOrDefault(1))

	some := optional.Some(1)
	assert.True(t, some.Has())
	assert.Equal(t, int(1), some.ValueOrZeroValue())
	assert.Equal(t, int(1), some.ValueOrDefault(2))

	noneBool := optional.None[bool]()
	assert.False(t, noneBool.Has())
	assert.False(t, noneBool.ValueOrZeroValue())
	assert.True(t, noneBool.ValueOrDefault(true))

	someBool := optional.Some(true)
	assert.True(t, someBool.Has())
	assert.True(t, someBool.ValueOrZeroValue())
	assert.True(t, someBool.ValueOrDefault(false))

	var ptr *int
	assert.False(t, optional.FromPtr(ptr).Has())

	int1 := 1
	opt1 := optional.FromPtr(&int1)
	assert.True(t, opt1.Has())
	_, v := opt1.Get()
	assert.Equal(t, int(1), v)

	assert.False(t, optional.FromNonDefault("").Has())

	opt2 := optional.FromNonDefault("test")
	assert.True(t, opt2.Has())
	_, vStr := opt2.Get()
	assert.Equal(t, "test", vStr)

	assert.False(t, optional.FromNonDefault(0).Has())

	opt3 := optional.FromNonDefault(1)
	assert.True(t, opt3.Has())
	_, v = opt3.Get()
	assert.Equal(t, int(1), v)
}

func Test_ParseBool(t *testing.T) {
	assert.Equal(t, optional.None[bool](), optional.ParseBool(""))
	assert.Equal(t, optional.None[bool](), optional.ParseBool("x"))

	assert.Equal(t, optional.Some(false), optional.ParseBool("0"))
	assert.Equal(t, optional.Some(false), optional.ParseBool("f"))
	assert.Equal(t, optional.Some(false), optional.ParseBool("False"))

	assert.Equal(t, optional.Some(true), optional.ParseBool("1"))
	assert.Equal(t, optional.Some(true), optional.ParseBool("t"))
	assert.Equal(t, optional.Some(true), optional.ParseBool("True"))
}

func roundtrip[T any](t *testing.T, orig optional.Option[T]) {
	// invoke (driver.Valuer).Value to get a DB value
	dbValue, err := orig.Value()
	require.NoError(t, err)

	// invoke (sql.Scanner).Scan to read the DB value
	var scanned optional.Option[T]
	err = scanned.Scan(dbValue)
	require.NoError(t, err)

	hasOrig, origValue := orig.Get()
	hasScanned, scannedValue := scanned.Get()

	if hasOrig {
		require.True(t, hasScanned, "must hasScanned")
		assert.Equal(t, origValue, scannedValue)
	} else {
		assert.False(t, hasScanned, "must not hasScanned")
	}
}

func TestOptionValueScan(t *testing.T) {
	t.Run("string roundtrip", func(t *testing.T) {
		roundtrip(t, optional.Some("hello world"))
	})
	t.Run("string null", func(t *testing.T) {
		roundtrip(t, optional.None[string]())
	})
	t.Run("int64 roundtrip", func(t *testing.T) {
		roundtrip(t, optional.Some(int64(1234)))
	})
	t.Run("int64 null", func(t *testing.T) {
		roundtrip(t, optional.None[int64]())
	})
	t.Run("bool roundtrip", func(t *testing.T) {
		roundtrip(t, optional.Some(false))
	})
	t.Run("bool null", func(t *testing.T) {
		roundtrip(t, optional.None[bool]())
	})
}

func TestDelegateSQLType(t *testing.T) {
	assert.Equal(t, "string", optional.Some("hello world").DelegateSQLType().Name())
	assert.Equal(t, "string", optional.None[string]().DelegateSQLType().Name())
	assert.Equal(t, "int64", optional.Some(int64(123)).DelegateSQLType().Name())
	assert.Equal(t, "int64", optional.None[int64]().DelegateSQLType().Name())
}
