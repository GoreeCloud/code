// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package pull

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"forgejo.org/models"
	issues_model "forgejo.org/models/issues"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/gitrepo"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_expandDefaultMergeMessage(t *testing.T) {
	type args struct {
		template string
		vars     map[string]string
	}

	title := "PullRequestTitle"
	description := "Pull\nRequest\nDescription\n"
	vars := map[string]string{
		"PullRequestTitle":       title,
		"PullRequestDescription": description,
	}
	defaultTitle := "default_title"
	defaultBody := "default_body"
	expectedTitle := fmt.Sprintf("Merge %s", title)
	expectedDescription := fmt.Sprintf("Description: %s", description)
	expectedDescriptionMultiLine := fmt.Sprintf("Description:\n\n%s", description)
	expectedDescriptionMerged := fmt.Sprintf("\n\nMerge %s\n\nDescription:\n\n%s", title, description)
	emptyString := ""

	tests := []struct {
		name     string
		args     args
		want     string
		wantBody string
	}{
		{
			name:     "empty template",
			args:     args{template: "", vars: vars},
			want:     defaultTitle,
			wantBody: defaultBody,
		},
		{
			name:     "single line",
			args:     args{template: "Merge ${PullRequestTitle}", vars: vars},
			want:     expectedTitle,
			wantBody: defaultBody,
		},
		{
			name:     "empty message (space)",
			args:     args{template: " ", vars: vars},
			want:     emptyString,
			wantBody: defaultBody,
		},
		{
			name:     "empty message (with newline)",
			args:     args{template: " \n", vars: vars},
			want:     emptyString,
			wantBody: defaultBody,
		},
		{
			name:     "single newline",
			args:     args{template: "\n", vars: vars},
			want:     defaultTitle,
			wantBody: defaultBody,
		},
		{
			name:     "empty description (newline)",
			args:     args{template: "\n\n", vars: vars},
			want:     defaultTitle,
			wantBody: emptyString,
		},
		{
			name:     "empty description (space)",
			args:     args{template: "\n ", vars: vars},
			want:     defaultTitle,
			wantBody: emptyString,
		},
		{
			name:     "empty title and description (spaces)",
			args:     args{template: " \n ", vars: vars},
			want:     emptyString,
			wantBody: emptyString,
		},
		{
			name:     "empty title and description (space and newline)",
			args:     args{template: " \n\n", vars: vars},
			want:     emptyString,
			wantBody: emptyString,
		},
		{
			name:     "simple message and description",
			args:     args{template: "Merge ${PullRequestTitle}\nDescription: ${PullRequestDescription}", vars: vars},
			want:     expectedTitle,
			wantBody: expectedDescription,
		},
		{
			name:     "multiple lines",
			args:     args{template: "Merge ${PullRequestTitle}\nDescription:\n\n${PullRequestDescription}\n", vars: vars},
			want:     expectedTitle,
			wantBody: expectedDescriptionMultiLine,
		},
		{
			name:     "description only",
			args:     args{template: "\nDescription: ${PullRequestDescription}\n", vars: vars},
			want:     defaultTitle,
			wantBody: expectedDescription,
		},
		{
			name:     "leading newlines",
			args:     args{template: "\n\n\nMerge ${PullRequestTitle}\n\nDescription:\n\n${PullRequestDescription}\n", vars: vars},
			want:     defaultTitle,
			wantBody: expectedDescriptionMerged,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultTitle, resultBody, err := expandDefaultMergeMessage(tt.args.template, tt.args.vars, defaultTitle, defaultBody)
			require.NoError(t, err)
			assert.Equalf(t, tt.want, resultTitle, "Wrong title for test '%s' -> expandDefaultMergeMessage(%q, %q)", tt.name, tt.args.template, tt.args.vars)
			assert.Equalf(t, tt.wantBody, resultBody, "Wrong body for test '%s' -> expandDefaultMergeMessage(%q, %q)", tt.name, tt.args.template, tt.args.vars)
		})
	}
}

func prepareLoadMergeMessageTemplates(targetDir string) error {
	for _, template := range []string{"MERGE", "REBASE", "REBASE-MERGE", "SQUASH", "MANUALLY-MERGED", "REBASE-UPDATE-ONLY"} {
		file, err := os.Create(path.Join(targetDir, template+"_TEMPLATE.md"))
		defer file.Close()

		if err == nil {
			_, err = file.WriteString("Contents for " + template)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func TestLoadMergeMessageTemplates(t *testing.T) {
	defer test.MockVariableValue(&setting.CustomPath, t.TempDir())()
	templateTemp := path.Join(setting.CustomPath, "default_merge_message")

	require.NoError(t, os.MkdirAll(templateTemp, 0o755))
	require.NoError(t, prepareLoadMergeMessageTemplates(templateTemp))

	testStyles := []repo_model.MergeStyle{
		repo_model.MergeStyleMerge,
		repo_model.MergeStyleRebase,
		repo_model.MergeStyleRebaseMerge,
		repo_model.MergeStyleSquash,
		repo_model.MergeStyleManuallyMerged,
		repo_model.MergeStyleRebaseUpdate,
	}

	// Load all templates
	require.NoError(t, LoadMergeMessageTemplates())

	// Check their correctness
	assert.Len(t, mergeMessageTemplates, len(testStyles))
	for _, mergeStyle := range testStyles {
		assert.Equal(t, "Contents for "+strings.ToUpper(string(mergeStyle)), mergeMessageTemplates[mergeStyle])
	}

	// Unload all templates
	require.NoError(t, os.RemoveAll(templateTemp))
	require.NoError(t, LoadMergeMessageTemplates())
	assert.Empty(t, mergeMessageTemplates)
}

func TestMergeMergedPR(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	pr := unittest.AssertExistsAndLoadBean(t, &issues_model.PullRequest{ID: 1})
	doer := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})

	require.NoError(t, pr.LoadBaseRepo(t.Context()))

	gitRepo, err := gitrepo.OpenRepository(t.Context(), pr.BaseRepo)
	require.NoError(t, err)
	defer gitRepo.Close()

	assert.True(t, pr.HasMerged)
	pr.HasMerged = false

	err = Merge(t.Context(), pr, doer, gitRepo, repo_model.MergeStyleRebase, "", "I should not exist", false)
	require.Error(t, err)
	assert.True(t, models.IsErrPullRequestHasMerged(err))

	if mergeErr, ok := err.(models.ErrPullRequestHasMerged); ok {
		assert.Equal(t, pr.ID, mergeErr.ID)
		assert.Equal(t, pr.IssueID, mergeErr.IssueID)
		assert.Equal(t, pr.HeadBranch, mergeErr.HeadBranch)
		assert.Equal(t, pr.BaseBranch, mergeErr.BaseBranch)
	}
}
