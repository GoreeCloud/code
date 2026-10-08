// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package quota_test

import (
	"testing"

	quota_model "forgejo.org/models/quota"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
)

func TestQuotaGroupAllRulesMustAllow(t *testing.T) {
	unlimitedRule := quota_model.Rule{
		Limit: -1,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeAll,
		},
	}
	denyRule := quota_model.Rule{
		Limit: 0,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeAll,
		},
	}
	group := quota_model.Group{
		Rules: []quota_model.Rule{
			unlimitedRule,
			denyRule,
		},
	}

	used := quota_model.Used{}
	used.Size.Repos.Public = 1024

	// Within a group, *all* matching rules must allow. Thus, if we have a deny-all rule,
	// and an unlimited rule, the deny rule wins.
	match, allow := group.Evaluate(used, quota_model.LimitSubjectSizeAll)
	assert.True(t, match)
	assert.False(t, allow)
}

func TestQuotaGroupRuleScenario1(t *testing.T) {
	group := quota_model.Group{
		Rules: []quota_model.Rule{
			{
				Limit: 1024,
				Subjects: quota_model.LimitSubjects{
					quota_model.LimitSubjectSizeAssetsAttachmentsReleases,
					quota_model.LimitSubjectSizeGitLFS,
					quota_model.LimitSubjectSizeAssetsPackagesAll,
				},
			},
			{
				Limit: 0,
				Subjects: quota_model.LimitSubjects{
					quota_model.LimitSubjectSizeGitLFS,
				},
			},
		},
	}

	used := quota_model.Used{}
	used.Size.Assets.Attachments.Releases = 512
	used.Size.Assets.Packages.All = 256
	used.Size.Git.LFS = 16

	match, allow := group.Evaluate(used, quota_model.LimitSubjectSizeAssetsAttachmentsReleases)
	assert.True(t, match, "size:assets:attachments:releases is covered")
	assert.True(t, allow, "size:assets:attachments:releases is allowed")

	match, allow = group.Evaluate(used, quota_model.LimitSubjectSizeAssetsPackagesAll)
	assert.True(t, match, "size:assets:packages:all is covered")
	assert.True(t, allow, "size:assets:packages:all is allowed")

	match, allow = group.Evaluate(used, quota_model.LimitSubjectSizeGitLFS)
	assert.True(t, match, "size:git:lfs is covered")
	assert.False(t, allow, "size:git:lfs is denied")

	match, allow = group.Evaluate(used, quota_model.LimitSubjectSizeAll)
	assert.False(t, match, "size:all is not covered")
	assert.False(t, allow, "size:all is not allowed (not covered)")
}

func TestQuotaGroupRuleCombination(t *testing.T) {
	repoRule := quota_model.Rule{
		Limit: 4096,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeReposAll,
		},
	}
	packagesRule := quota_model.Rule{
		Limit: 0,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeAssetsPackagesAll,
		},
	}

	used := quota_model.Used{}
	used.Size.Repos.Public = 1024
	used.Size.Assets.Packages.All = 1024

	group := quota_model.Group{
		Rules: []quota_model.Rule{
			repoRule,
			packagesRule,
		},
	}

	// Git LFS does not match any rule
	match, allow := group.Evaluate(used, quota_model.LimitSubjectSizeGitLFS)
	assert.False(t, match)
	assert.False(t, allow)

	// repos:all has a matching rule and is allowed
	match, allow = group.Evaluate(used, quota_model.LimitSubjectSizeReposAll)
	assert.True(t, match)
	assert.True(t, allow)

	// packages:all has a matching rule and is denied
	match, allow = group.Evaluate(used, quota_model.LimitSubjectSizeAssetsPackagesAll)
	assert.True(t, match)
	assert.False(t, allow)
}

func TestQuotaGroupListsRequireOnlyOneAllow(t *testing.T) {
	unlimitedRule := quota_model.Rule{
		Limit: -1,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeAll,
		},
	}
	denyRule := quota_model.Rule{
		Limit: 0,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeAll,
		},
	}

	denyGroup := quota_model.Group{
		Rules: []quota_model.Rule{
			denyRule,
		},
	}
	unlimitedGroup := quota_model.Group{
		Rules: []quota_model.Rule{
			unlimitedRule,
		},
	}

	groups := quota_model.GroupList{&denyGroup, &unlimitedGroup}

	used := quota_model.Used{}
	used.Size.Repos.Public = 1024

	// In a group list, an action is allowed if any group matches and allows it.
	allow := groups.Evaluate(used, quota_model.LimitSubjectSizeAll)
	assert.True(t, allow)
}

func TestQuotaGroupListAllDeny(t *testing.T) {
	denyRule := quota_model.Rule{
		Limit: 0,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeAll,
		},
	}
	limitedRule := quota_model.Rule{
		Limit: 1024,
		Subjects: quota_model.LimitSubjects{
			quota_model.LimitSubjectSizeAll,
		},
	}

	denyGroup := quota_model.Group{
		Rules: []quota_model.Rule{
			denyRule,
		},
	}
	limitedGroup := quota_model.Group{
		Rules: []quota_model.Rule{
			limitedRule,
		},
	}

	groups := quota_model.GroupList{&denyGroup, &limitedGroup}

	used := quota_model.Used{}
	used.Size.Repos.Public = 2048

	allow := groups.Evaluate(used, quota_model.LimitSubjectSizeAll)
	assert.False(t, allow)
}

// An empty group list should result in the use of the built in Default
// group: size:all defaulting to unlimited
func TestQuotaDefaultGroup(t *testing.T) {
	groups := quota_model.GroupList{}

	used := quota_model.Used{}
	used.Size.Repos.Public = 2048

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
			defer test.MockVariableValue(&setting.Quota.Default.Total, testSet.limit)()

			for subject := quota_model.LimitSubjectFirst; subject <= quota_model.LimitSubjectLast; subject++ {
				t.Run(subject.String(), func(t *testing.T) {
					allow := groups.Evaluate(used, subject)
					assert.Equal(t, testSet.expectAllow, allow)
				})
			}
		})
	}
}
