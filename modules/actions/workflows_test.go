// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package actions

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"forgejo.org/modules/git"
	"forgejo.org/modules/setting"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/test"
	webhook_module "forgejo.org/modules/webhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectedWorkflowGetWorkflowPath(t *testing.T) {
	buildWorkflow := DetectedWorkflow{EntryDirectory: ".github/workflows", EntryName: "build.yaml"}
	testWorkflow := DetectedWorkflow{EntryDirectory: ".forgejo/workflows", EntryName: "test.yaml"}

	assert.Equal(t, ".github/workflows/build.yaml", buildWorkflow.GetWorkflowPath())
	assert.Equal(t, ".forgejo/workflows/test.yaml", testWorkflow.GetWorkflowPath())
}

func TestActionsWorkflowsDetectMatched(t *testing.T) {
	testCases := []struct {
		desc           string
		commit         *git.Commit
		triggeredEvent webhook_module.HookEventType
		payload        api.Payloader
		yamlOn         string
		expected       bool
	}{
		{
			desc:           "HookEventCreate(create) matches GithubEventCreate(create)",
			triggeredEvent: webhook_module.HookEventCreate,
			payload:        nil,
			yamlOn:         "on: create",
			expected:       true,
		},
		{
			desc:           "HookEventIssues(issues) `opened` action matches GithubEventIssues(issues)",
			triggeredEvent: webhook_module.HookEventIssues,
			payload:        &api.IssuePayload{Action: api.HookIssueOpened},
			yamlOn:         "on: issues",
			expected:       true,
		},
		{
			desc:           "HookEventIssueComment(issue_comment) `created` action matches GithubEventIssueComment(issue_comment)",
			triggeredEvent: webhook_module.HookEventIssueComment,
			payload:        &api.IssueCommentPayload{Action: api.HookIssueCommentCreated},
			yamlOn:         "on:\n  issue_comment:\n    types: [created]",
			expected:       true,
		},

		{
			desc:           "HookEventIssues(issues) `milestoned` action matches GithubEventIssues(issues)",
			triggeredEvent: webhook_module.HookEventIssues,
			payload:        &api.IssuePayload{Action: api.HookIssueMilestoned},
			yamlOn:         "on: issues",
			expected:       true,
		},

		{
			desc:           "HookEventPullRequestSync(pull_request_sync) matches GithubEventPullRequest(pull_request)",
			triggeredEvent: webhook_module.HookEventPullRequestSync,
			payload:        &api.PullRequestPayload{Action: api.HookIssueSynchronized},
			yamlOn:         "on: pull_request",
			expected:       true,
		},
		{
			desc:           "HookEventPullRequest(pull_request) `label_updated` action doesn't match GithubEventPullRequest(pull_request) with no activity type",
			triggeredEvent: webhook_module.HookEventPullRequest,
			payload:        &api.PullRequestPayload{Action: api.HookIssueLabelUpdated},
			yamlOn:         "on: pull_request",
			expected:       false,
		},
		{
			desc:           "HookEventPullRequest(pull_request) `closed` action doesn't match GithubEventPullRequest(pull_request) with no activity type",
			triggeredEvent: webhook_module.HookEventPullRequest,
			payload:        &api.PullRequestPayload{Action: api.HookIssueClosed},
			yamlOn:         "on: pull_request",
			expected:       false,
		},
		{
			desc:           "HookEventPullRequest(pull_request) `closed` action doesn't match GithubEventPullRequest(pull_request) with branches",
			triggeredEvent: webhook_module.HookEventPullRequest,
			payload: &api.PullRequestPayload{
				Action: api.HookIssueClosed,
				PullRequest: &api.PullRequest{
					Base: &api.PRBranchInfo{},
				},
			},
			yamlOn:   "on:\n  pull_request:\n    branches: [main]",
			expected: false,
		},
		{
			desc:           "HookEventPullRequest(pull_request) `label_updated` action matches GithubEventPullRequest(pull_request) with `label` activity type",
			triggeredEvent: webhook_module.HookEventPullRequest,
			payload:        &api.PullRequestPayload{Action: api.HookIssueLabelUpdated},
			yamlOn:         "on:\n  pull_request:\n    types: [labeled]",
			expected:       true,
		},
		{
			desc:           "HookEventPullRequestReviewComment(pull_request_review_comment) matches GithubEventPullRequestReviewComment(pull_request_review_comment)",
			triggeredEvent: webhook_module.HookEventPullRequestReviewComment,
			payload:        &api.PullRequestPayload{Action: api.HookIssueReviewed},
			yamlOn:         "on:\n  pull_request_review_comment:\n    types: [created]",
			expected:       true,
		},
		{
			desc:           "HookEventPullRequestReviewRejected(pull_request_review_rejected) doesn't match GithubEventPullRequestReview(pull_request_review) with `dismissed` activity type (we don't support `dismissed` at present)",
			triggeredEvent: webhook_module.HookEventPullRequestReviewRejected,
			payload:        &api.PullRequestPayload{Action: api.HookIssueReviewed},
			yamlOn:         "on:\n  pull_request_review:\n    types: [dismissed]",
			expected:       false,
		},
		{
			desc:           "HookEventRelease(release) `published` action matches GithubEventRelease(release) with `published` activity type",
			triggeredEvent: webhook_module.HookEventRelease,
			payload:        &api.ReleasePayload{Action: api.HookReleasePublished},
			yamlOn:         "on:\n  release:\n    types: [published]",
			expected:       true,
		},
		{
			desc:           "HookEventRelease(updated) `updated` action matches GithubEventRelease(edited) with `edited` activity type",
			triggeredEvent: webhook_module.HookEventRelease,
			payload:        &api.ReleasePayload{Action: api.HookReleaseUpdated},
			yamlOn:         "on:\n  release:\n    types: [edited]",
			expected:       true,
		},

		{
			desc:           "HookEventPackage(package) `created` action doesn't match GithubEventRegistryPackage(registry_package) with `updated` activity type",
			triggeredEvent: webhook_module.HookEventPackage,
			payload:        &api.PackagePayload{Action: api.HookPackageCreated},
			yamlOn:         "on:\n  registry_package:\n    types: [updated]",
			expected:       false,
		},
		{
			desc:           "HookEventWiki(wiki) matches GithubEventGollum(gollum)",
			triggeredEvent: webhook_module.HookEventWiki,
			payload:        nil,
			yamlOn:         "on: gollum",
			expected:       true,
		},
		{
			desc:           "HookEventSchedule(schedule) matches GithubEventSchedule(schedule)",
			triggeredEvent: webhook_module.HookEventSchedule,
			payload:        nil,
			yamlOn:         "on: schedule",
			expected:       true,
		},
		{
			desc:           "HookEventWorkflowDispatch(workflow_dispatch) matches GithubEventWorkflowDispatch(workflow_dispatch)",
			triggeredEvent: webhook_module.HookEventWorkflowDispatch,
			payload:        nil,
			yamlOn:         "on: workflow_dispatch",
			expected:       true,
		},
		{
			desc:           "push to tag matches workflow with paths condition (should skip paths check)",
			triggeredEvent: webhook_module.HookEventPush,
			payload: &api.PushPayload{
				Ref:    "refs/tags/v1.0.0",
				Before: "0000000",
				Commits: []*api.PayloadCommit{
					{
						ID:      "abcdef123456",
						Added:   []string{"src/main.go"},
						Message: "Release v1.0.0",
					},
				},
			},
			commit:   nil,
			yamlOn:   "on:\n  push:\n    paths:\n      - src/**",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			evts, err := GetEventsFromContent([]byte(tc.yamlOn))
			require.NoError(t, err)
			assert.Len(t, evts, 1)
			assert.Equal(t, tc.expected, detectMatched(nil, tc.commit, tc.triggeredEvent, tc.payload, evts[0]))
		})
	}
}

func TestActionsWorkflowsListWorkflowsReturnsNoWorkflowsIfThereAreNone(t *testing.T) {
	t.Cleanup(test.MockVariableValue(&setting.Git.HomePath, t.TempDir()))
	require.NoError(t, git.InitSimple(t.Context()))

	committer := git.Signature{
		Email: "jane@example.com",
		Name:  "Jane",
		When:  time.Now(),
	}
	repoHome := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(repoHome, "README.md"), []byte("My project"), 0o644))

	require.NoError(t, git.InitRepository(t.Context(), repoHome, false, git.Sha1ObjectFormat.Name()))
	require.NoError(t, git.AddChanges(repoHome, true))
	require.NoError(t, git.CommitChanges(repoHome, git.CommitChangesOptions{Message: "Import", Committer: &committer}))

	gitRepo, err := git.OpenRepository(t.Context(), repoHome)
	require.NoError(t, err)
	defer gitRepo.Close()

	headBranch, err := gitRepo.GetHEADBranch()
	require.NoError(t, err)

	lastCommitID, err := gitRepo.GetBranchCommitID(headBranch.Name)
	require.NoError(t, err)

	lastCommit, err := gitRepo.GetCommit(lastCommitID)
	require.NoError(t, err)

	source, workflows, err := ListWorkflows(lastCommit)
	require.NoError(t, err)

	assert.Empty(t, source)
	assert.Empty(t, workflows)
}

func TestActionsWorkflowsListWorkflowsIgnoresNonWorkflowFiles(t *testing.T) {
	t.Cleanup(test.MockVariableValue(&setting.Git.HomePath, t.TempDir()))
	require.NoError(t, git.InitSimple(t.Context()))

	committer := git.Signature{
		Email: "jane@example.com",
		Name:  "Jane",
		When:  time.Now(),
	}
	githubWorkflow := []byte(`
name: GitHub Workflow
on:
  push:
jobs:
  do-something:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'Hello GitHub'
`)
	repoHome := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(repoHome, ".forgejo/workflows"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(repoHome, ".forgejo/workflows", "README.md"), []byte("My project"), 0o644))

	// Prepare a valid workflow in .github/workflows to verify that it is ignored because .forgejo/workflows is present.
	require.NoError(t, os.MkdirAll(filepath.Join(repoHome, ".github/workflows"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(repoHome, ".github/workflows", "github.yaml"), githubWorkflow, 0o644))

	require.NoError(t, git.InitRepository(t.Context(), repoHome, false, git.Sha1ObjectFormat.Name()))
	require.NoError(t, git.AddChanges(repoHome, true))
	require.NoError(t, git.CommitChanges(repoHome, git.CommitChangesOptions{Message: "Import", Committer: &committer}))

	gitRepo, err := git.OpenRepository(t.Context(), repoHome)
	require.NoError(t, err)
	defer gitRepo.Close()

	headBranch, err := gitRepo.GetHEADBranch()
	require.NoError(t, err)

	lastCommitID, err := gitRepo.GetBranchCommitID(headBranch.Name)
	require.NoError(t, err)

	lastCommit, err := gitRepo.GetCommit(lastCommitID)
	require.NoError(t, err)

	source, workflows, err := ListWorkflows(lastCommit)
	require.NoError(t, err)

	assert.Equal(t, ".forgejo/workflows", source)
	assert.Empty(t, workflows)
}

func TestActionsWorkflowsListWorkflowsReturnsForgejoWorkflowsOnly(t *testing.T) {
	t.Cleanup(test.MockVariableValue(&setting.Git.HomePath, t.TempDir()))
	require.NoError(t, git.InitSimple(t.Context()))

	committer := git.Signature{
		Email: "jane@example.com",
		Name:  "Jane",
		When:  time.Now(),
	}
	forgejoWorkflow := []byte(`
name: Forgejo Workflow
on:
  push:
jobs:
  do-something:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'Hello Forgejo'
`)
	githubWorkflow := []byte(`
name: GitHub Workflow
on:
  push:
jobs:
  do-something:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'Hello GitHub'
`)
	repoHome := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(repoHome, ".forgejo/workflows"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(repoHome, ".forgejo/workflows", "forgejo.yaml"), forgejoWorkflow, 0o644))

	require.NoError(t, os.MkdirAll(filepath.Join(repoHome, ".github/workflows"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(repoHome, ".github/workflows", "github.yaml"), githubWorkflow, 0o644))

	require.NoError(t, git.InitRepository(t.Context(), repoHome, false, git.Sha1ObjectFormat.Name()))
	require.NoError(t, git.AddChanges(repoHome, true))
	require.NoError(t, git.CommitChanges(repoHome, git.CommitChangesOptions{Message: "Import", Committer: &committer}))

	gitRepo, err := git.OpenRepository(t.Context(), repoHome)
	require.NoError(t, err)
	defer gitRepo.Close()

	headBranch, err := gitRepo.GetHEADBranch()
	require.NoError(t, err)

	lastCommitID, err := gitRepo.GetBranchCommitID(headBranch.Name)
	require.NoError(t, err)

	lastCommit, err := gitRepo.GetCommit(lastCommitID)
	require.NoError(t, err)

	source, workflows, err := ListWorkflows(lastCommit)
	require.NoError(t, err)

	assert.Len(t, workflows, 1)
	assert.Equal(t, ".forgejo/workflows", source)
	assert.Equal(t, "forgejo.yaml", workflows[0].Name())
}

func TestActionsWorkflowsListWorkflowsReturnsGitHubWorkflowsIfForgejoWorkflowsAbsent(t *testing.T) {
	t.Cleanup(test.MockVariableValue(&setting.Git.HomePath, t.TempDir()))
	require.NoError(t, git.InitSimple(t.Context()))

	committer := git.Signature{
		Email: "jane@example.com",
		Name:  "Jane",
		When:  time.Now(),
	}
	buildWorkflow := []byte(`
name: Build
on:
  push:
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'We are building'
`)
	testWorkflow := []byte(`
name: Test
on:
  push:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'We are testing'
`)
	repoHome := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(repoHome, ".github/workflows"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(repoHome, ".github/workflows", "build.yaml"), buildWorkflow, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repoHome, ".github/workflows", "test.yml"), testWorkflow, 0o644))

	require.NoError(t, git.InitRepository(t.Context(), repoHome, false, git.Sha1ObjectFormat.Name()))
	require.NoError(t, git.AddChanges(repoHome, true))
	require.NoError(t, git.CommitChanges(repoHome, git.CommitChangesOptions{Message: "Import", Committer: &committer}))

	gitRepo, err := git.OpenRepository(t.Context(), repoHome)
	require.NoError(t, err)
	defer gitRepo.Close()

	headBranch, err := gitRepo.GetHEADBranch()
	require.NoError(t, err)

	lastCommitID, err := gitRepo.GetBranchCommitID(headBranch.Name)
	require.NoError(t, err)

	lastCommit, err := gitRepo.GetCommit(lastCommitID)
	require.NoError(t, err)

	source, workflows, err := ListWorkflows(lastCommit)
	require.NoError(t, err)

	assert.Len(t, workflows, 2)
	assert.Equal(t, ".github/workflows", source)
	assert.Equal(t, "build.yaml", workflows[0].Name())
	assert.Equal(t, "test.yml", workflows[1].Name())
}
