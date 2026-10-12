// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package markup

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failReader struct{}

func (*failReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("FAIL")
}

func TestRender_postProcessOrCopy(t *testing.T) {
	renderContext := &RenderContext{Ctx: t.Context()}

	t.Run("CopyOK", func(t *testing.T) {
		input := "SOMETHING"
		output := &bytes.Buffer{}
		require.NoError(t, postProcessOrCopy(renderContext, nil, strings.NewReader(input), output))
		assert.Equal(t, input, output.String())
	})

	renderer := GetRendererByType("markdown")

	t.Run("PostProcessOK", func(t *testing.T) {
		input := "SOMETHING"
		output := &bytes.Buffer{}
		defer test.MockVariableValue(&defaultProcessors, []processor{})()
		require.NoError(t, postProcessOrCopy(renderContext, renderer, strings.NewReader(input), output))
		assert.Equal(t, input, output.String())
	})

	t.Run("PostProcessError", func(t *testing.T) {
		input := &failReader{}
		defer test.MockVariableValue(&defaultProcessors, []processor{})()
		assert.ErrorContains(t, postProcessOrCopy(renderContext, renderer, input, &bytes.Buffer{}), "FAIL")
	})
}
