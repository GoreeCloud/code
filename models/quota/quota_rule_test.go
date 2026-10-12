// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package quota_test

import (
	"testing"

	quota_model "forgejo.org/models/quota"

	"github.com/stretchr/testify/assert"
)

func makeFullyUsed() quota_model.Used {
	return quota_model.Used{
		Size: quota_model.UsedSize{
			Repos: quota_model.UsedSizeRepos{
				Public:  1024,
				Private: 1024,
			},
			Git: quota_model.UsedSizeGit{
				LFS: 1024,
			},
			Assets: quota_model.UsedSizeAssets{
				Attachments: quota_model.UsedSizeAssetsAttachments{
					Issues:   1024,
					Releases: 1024,
				},
				Artifacts: 1024,
				Packages: quota_model.UsedSizeAssetsPackages{
					All: 1024,
				},
			},
		},
	}
}

func makePartiallyUsed() quota_model.Used {
	return quota_model.Used{
		Size: quota_model.UsedSize{
			Repos: quota_model.UsedSizeRepos{
				Public: 1024,
			},
			Assets: quota_model.UsedSizeAssets{
				Attachments: quota_model.UsedSizeAssetsAttachments{
					Releases: 1024,
				},
			},
		},
	}
}

func setUsed(used quota_model.Used, subject quota_model.LimitSubject, value int64) *quota_model.Used {
	switch subject {
	case quota_model.LimitSubjectSizeReposPublic:
		used.Size.Repos.Public = value
		return &used
	case quota_model.LimitSubjectSizeReposPrivate:
		used.Size.Repos.Private = value
		return &used
	case quota_model.LimitSubjectSizeGitLFS:
		used.Size.Git.LFS = value
		return &used
	case quota_model.LimitSubjectSizeAssetsAttachmentsIssues:
		used.Size.Assets.Attachments.Issues = value
		return &used
	case quota_model.LimitSubjectSizeAssetsAttachmentsReleases:
		used.Size.Assets.Attachments.Releases = value
		return &used
	case quota_model.LimitSubjectSizeAssetsArtifacts:
		used.Size.Assets.Artifacts = value
		return &used
	case quota_model.LimitSubjectSizeAssetsPackagesAll:
		used.Size.Assets.Packages.All = value
		return &used
	case quota_model.LimitSubjectSizeWiki:
	}

	return nil
}

func assertEvaluation(t *testing.T, rule quota_model.Rule, used quota_model.Used, subject quota_model.LimitSubject, expected bool) {
	t.Helper()

	t.Run(subject.String(), func(t *testing.T) {
		match, allow := rule.Evaluate(used, subject)
		assert.True(t, match)
		assert.Equal(t, expected, allow)
	})
}

func TestQuotaRuleNoMatch(t *testing.T) {
	testSets := []struct {
		name  string
		limit int64
	}{
		{"unlimited", -1},
		{"limit-0", 0},
		{"limit-1k", 1024},
		{"limit-1M", 1024 * 1024},
	}

	for _, testSet := range testSets {
		t.Run(testSet.name, func(t *testing.T) {
			rule := quota_model.Rule{
				Limit: testSet.limit,
				Subjects: quota_model.LimitSubjects{
					quota_model.LimitSubjectSizeAssetsAttachmentsAll,
				},
			}
			used := quota_model.Used{}
			used.Size.Repos.Public = 4096

			match, allow := rule.Evaluate(used, quota_model.LimitSubjectSizeReposAll)

			// We have a rule for "size:assets:attachments:all", and query for
			// "size:repos:all". We don't cover that subject, so the rule does not match
			// regardless of the limit.
			assert.False(t, match)
			assert.False(t, allow)
		})
	}
}

func TestQuotaRuleDirectEvaluation(t *testing.T) {
	// This function is meant to test direct rule evaluation: cases where we set
	// a rule for a subject, and we evaluate against the same subject.

	runTest := func(t *testing.T, subject quota_model.LimitSubject, limit, used int64, expected bool) {
		t.Helper()

		rule := quota_model.Rule{
			Limit: limit,
			Subjects: quota_model.LimitSubjects{
				subject,
			},
		}
		usedObj := setUsed(quota_model.Used{}, subject, used)
		if usedObj == nil {
			return
		}

		assertEvaluation(t, rule, *usedObj, subject, expected)
	}

	t.Run("limit:0", func(t *testing.T) {
		// With limit:0, any usage will fail evaluation, including 0
		t.Run("used:0", func(t *testing.T) {
			for subject := quota_model.LimitSubjectFirst; subject <= quota_model.LimitSubjectLast; subject++ {
				runTest(t, subject, 0, 0, false)
			}
		})
		t.Run("used:512", func(t *testing.T) {
			for subject := quota_model.LimitSubjectFirst; subject <= quota_model.LimitSubjectLast; subject++ {
				runTest(t, subject, 0, 512, false)
			}
		})
	})

	t.Run("limit:unlimited", func(t *testing.T) {
		// With no limits, any usage will succeed evaluation
		t.Run("used:512", func(t *testing.T) {
			for subject := quota_model.LimitSubjectFirst; subject <= quota_model.LimitSubjectLast; subject++ {
				runTest(t, subject, -1, 512, true)
			}
		})
	})

	t.Run("limit:1024", func(t *testing.T) {
		// With a set limit, usage below the limit succeeds
		t.Run("used:512", func(t *testing.T) {
			for subject := quota_model.LimitSubjectFirst; subject <= quota_model.LimitSubjectLast; subject++ {
				runTest(t, subject, 1024, 512, true)
			}
		})

		// With a set limit, usage above the limit fails
		t.Run("used:2048", func(t *testing.T) {
			for subject := quota_model.LimitSubjectFirst; subject <= quota_model.LimitSubjectLast; subject++ {
				runTest(t, subject, 1024, 2048, false)
			}
		})
	})
}

func TestQuotaRuleCombined(t *testing.T) {
	used := quota_model.Used{
		Size: quota_model.UsedSize{
			Repos: quota_model.UsedSizeRepos{
				Public: 4096,
			},
			Git: quota_model.UsedSizeGit{
				LFS: 256,
			},
			Assets: quota_model.UsedSizeAssets{
				Attachments: quota_model.UsedSizeAssetsAttachments{
					Issues:   2048,
					Releases: 256,
				},
				Packages: quota_model.UsedSizeAssetsPackages{
					All: 2560,
				},
			},
		},
	}

	expectMatch := map[quota_model.LimitSubject]bool{
		quota_model.LimitSubjectSizeGitLFS:                    true,
		quota_model.LimitSubjectSizeAssetsAttachmentsReleases: true,
		quota_model.LimitSubjectSizeAssetsPackagesAll:         true,
	}

	testSets := []struct {
		name        string
		limit       int64
		expectAllow bool
	}{
		{"unlimited", -1, true},
		{"limit-allow", 1024 * 1024, true},
		{"limit-deny", 1024, false},
	}

	for _, testSet := range testSets {
		t.Run(testSet.name, func(t *testing.T) {
			rule := quota_model.Rule{
				Limit: testSet.limit,
				Subjects: quota_model.LimitSubjects{
					quota_model.LimitSubjectSizeGitLFS,
					quota_model.LimitSubjectSizeAssetsAttachmentsReleases,
					quota_model.LimitSubjectSizeAssetsPackagesAll,
				},
			}

			for subject := quota_model.LimitSubjectFirst; subject <= quota_model.LimitSubjectLast; subject++ {
				t.Run(subject.String(), func(t *testing.T) {
					match, allow := rule.Evaluate(used, subject)

					assert.Equal(t, expectMatch[subject], match)
					if expectMatch[subject] {
						assert.Equal(t, testSet.expectAllow, allow)
					} else {
						assert.False(t, allow)
					}
				})
			}
		})
	}
}

func TestQuotaRuleSizeAll(t *testing.T) {
	type Test struct {
		name        string
		limit       int64
		expectAllow bool
	}

	usedSets := []struct {
		name     string
		used     quota_model.Used
		testSets []Test
	}{
		{
			"empty",
			quota_model.Used{},
			[]Test{
				{"unlimited", -1, true},
				{"limit-1M", 1024 * 1024, true},
				{"limit-5k", 5 * 1024, true},
				{"limit-0", 0, false},
			},
		},
		{
			"partial",
			makePartiallyUsed(),
			[]Test{
				{"unlimited", -1, true},
				{"limit-1M", 1024 * 1024, true},
				{"limit-5k", 5 * 1024, true},
				{"limit-0", 0, false},
			},
		},
		{
			"full",
			makeFullyUsed(),
			[]Test{
				{"unlimited", -1, true},
				{"limit-1M", 1024 * 1024, true},
				{"limit-5k", 5 * 1024, false},
				{"limit-0", 0, false},
			},
		},
	}

	for _, usedSet := range usedSets {
		t.Run(usedSet.name, func(t *testing.T) {
			testSets := usedSet.testSets
			used := usedSet.used

			for _, testSet := range testSets {
				t.Run(testSet.name, func(t *testing.T) {
					rule := quota_model.Rule{
						Limit: testSet.limit,
						Subjects: quota_model.LimitSubjects{
							quota_model.LimitSubjectSizeAll,
						},
					}

					match, allow := rule.Evaluate(used, quota_model.LimitSubjectSizeAll)
					assert.True(t, match)
					assert.Equal(t, testSet.expectAllow, allow)
				})
			}
		})
	}
}
