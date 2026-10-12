// Copyright 2023 The Gitea Authors. All rights reserved.
// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package translation

import (
	"testing"

	"forgejo.org/modules/translation/i18n"

	"github.com/stretchr/testify/assert"
)

func TestTrSize(t *testing.T) {
	l := NewLocale("")
	size := int64(1)
	assert.Equal(t, "1 munits.data.b", l.TrSize(size).String())
	size *= 2048
	assert.Equal(t, "2 munits.data.kib", l.TrSize(size).String())
	size *= 2048
	assert.Equal(t, "4 munits.data.mib", l.TrSize(size).String())
	size *= 2048
	assert.Equal(t, "8 munits.data.gib", l.TrSize(size).String())
	size *= 2048
	assert.Equal(t, "16 munits.data.tib", l.TrSize(size).String())
	size *= 2048
	assert.Equal(t, "32 munits.data.pib", l.TrSize(size).String())
	size *= 128
	assert.Equal(t, "4 munits.data.eib", l.TrSize(size).String())
}

func TestPrettyNumber(t *testing.T) {
	i18n.ResetDefaultLocales()

	allLangMap = make(map[string]*LangType)
	allLangMap["id-ID"] = &LangType{Lang: "id-ID", Name: "Bahasa Indonesia"}

	l := NewLocale("id-ID")
	assert.Equal(t, "1.000.000", l.PrettyNumber(1000000))
	assert.Equal(t, "1.000.000,1", l.PrettyNumber(1000000.1))
	assert.Equal(t, "1.000.000", l.PrettyNumber("1000000"))
	assert.Equal(t, "1.000.000", l.PrettyNumber("1000000.0"))
	assert.Equal(t, "1.000.000,1", l.PrettyNumber("1000000.1"))

	l = NewLocale("nosuch")
	assert.Equal(t, "1,000,000", l.PrettyNumber(1000000))
	assert.Equal(t, "1,000,000.1", l.PrettyNumber(1000000.1))
}
