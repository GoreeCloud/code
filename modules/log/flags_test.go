// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package log

import (
	"testing"

	"forgejo.org/modules/json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlags(t *testing.T) {
	assert.Equal(t, Ldefault, Flags{}.Bits())
	assert.EqualValues(t, 0, FlagsFromString("").Bits())
	assert.Equal(t, Ldate|Ltime, FlagsFromString("date,time").Bits())

	assert.Equal(t, "stdflags", FlagsFromString("stdflags").String())
	assert.Equal(t, "medfile", FlagsFromString("medfile").String())

	bs, err := json.Marshal(FlagsFromString("utc,level"))
	require.NoError(t, err)
	assert.Equal(t, `"level,utc"`, string(bs))
	var flags Flags
	require.NoError(t, json.Unmarshal(bs, &flags))
	assert.Equal(t, LUTC|Llevel, flags.Bits())
}
