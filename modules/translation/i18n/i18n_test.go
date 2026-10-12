// Copyright 2024-2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package i18n

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var MockPluralRule PluralFormRule = func(n int64) PluralFormIndex {
	if n == 0 {
		return PluralFormZero
	}
	if n == 1 {
		return PluralFormOne
	}
	if n >= 2 && n <= 4 {
		return PluralFormFew
	}
	return PluralFormOther
}

var MockPluralRuleEnglish PluralFormRule = func(n int64) PluralFormIndex {
	if n == 1 {
		return PluralFormOne
	}
	return PluralFormOther
}

var (
	UsedPluralFormsEnglish = []PluralFormIndex{PluralFormOne, PluralFormOther}
	UsedPluralFormsMock    = []PluralFormIndex{PluralFormZero, PluralFormOne, PluralFormFew, PluralFormOther}
)

func TestLocaleStoreJSON(t *testing.T) {
	testDataJSON2 := []byte(`
{
	"section.json": "the JSON is %s",
	"section.commits": {
		"one": "one %d commit",
		"few": "some %d commits",
		"other": "lots of %d commits"
	},
	"section.incomplete": {
		"few": "some %d objects (translated)"
	},
	"nested": {
		"outer": {
			"inner": {
				"json": "Hello World",
				"issue": {
					"one": "one %d issue",
					"few": "some %d issues",
					"other": "lots of %d issues"
				}
			}
		}
	}
}
`)
	testDataJSON1 := []byte(`
{
	"section.incomplete": {
		"one": "[untranslated] some %d object",
		"other": "[untranslated] some %d objects"
	}
}
`)

	ls := NewLocaleStore()

	// Currently LocaleStore has to be first populated with langcodes via AddLocaleByIni
	require.NoError(t, ls.AddLocaleByIni("lang1", "Lang1", MockPluralRuleEnglish, UsedPluralFormsEnglish, []byte(""), nil))
	require.NoError(t, ls.AddLocaleByIni("lang2", "Lang2", MockPluralRule, UsedPluralFormsMock, []byte(""), nil))

	require.NoError(t, ls.AddToLocaleFromJSON("lang1", testDataJSON1))
	require.NoError(t, ls.AddToLocaleFromJSON("lang2", testDataJSON2))

	ls.SetDefaultLang("lang1")
	lang2, _ := ls.Locale("lang2")

	result := lang2.TrString("section.json", "valid")
	assert.Equal(t, "the JSON is valid", result)

	result = lang2.TrString("nested.outer.inner.json")
	assert.Equal(t, "Hello World", result)

	result = lang2.TrString("section.commits")
	assert.Equal(t, "lots of %d commits", result)

	result2 := lang2.TrPluralString(1, "section.commits", 1)
	assert.EqualValues(t, "one 1 commit", result2)

	result2 = lang2.TrPluralString(3, "section.commits", 3)
	assert.EqualValues(t, "some 3 commits", result2)

	result2 = lang2.TrPluralString(8, "section.commits", 8)
	assert.EqualValues(t, "lots of 8 commits", result2)

	result2 = lang2.TrPluralString(0, "section.commits")
	assert.EqualValues(t, "section.commits", result2)

	result2 = lang2.TrPluralString(1, "nested.outer.inner.issue", 1)
	assert.EqualValues(t, "one 1 issue", result2)

	result2 = lang2.TrPluralString(3, "nested.outer.inner.issue", 3)
	assert.EqualValues(t, "some 3 issues", result2)

	result2 = lang2.TrPluralString(9, "nested.outer.inner.issue", 9)
	assert.EqualValues(t, "lots of 9 issues", result2)

	result2 = lang2.TrPluralString(3, "section.incomplete", 3)
	assert.EqualValues(t, "some 3 objects (translated)", result2)

	result2 = lang2.TrPluralString(1, "section.incomplete", 1)
	assert.EqualValues(t, "[untranslated] some 1 object", result2)

	result2 = lang2.TrPluralString(7, "section.incomplete", 7)
	assert.EqualValues(t, "[untranslated] some 7 objects", result2)
}

func TestMissingTranslationHandling(t *testing.T) {
	ls := NewLocaleStore()

	// Currently LocaleStore has to be first populated with langcodes via AddLocaleByIni
	require.NoError(t, ls.AddLocaleByIni("en-US", "English", MockPluralRuleEnglish, UsedPluralFormsEnglish, []byte(""), nil))
	require.NoError(t, ls.AddLocaleByIni("fun", "Funlang", MockPluralRule, UsedPluralFormsMock, []byte(""), nil))

	require.NoError(t, ls.AddToLocaleFromJSON("en-US", []byte(`
{
	"incorrect_root_url": "This Forgejo instance...",
	"meta.last_line": "Hi there!"
}`)))
	require.NoError(t, ls.AddToLocaleFromJSON("fun", []byte(`
{
	"meta.last_line": "This language only has one line that is never used by the UI. It will never have a translation for incorrect_root_url"
}`)))

	ls.SetDefaultLang("en-US")

	// Get "fun" locale, make sure it's available
	funLocale, found := ls.Locale("fun")
	assert.True(t, found)

	// Get translation for a string that this locale doesn't have
	s := funLocale.TrString("incorrect_root_url")

	// Verify fallback to English
	assert.True(t, strings.HasPrefix(s, "This Forgejo instance..."))
}
