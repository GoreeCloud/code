// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"forgejo.org/models/db"
	issues_model "forgejo.org/models/issues"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/git"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/templates"
	"forgejo.org/modules/test"
	"forgejo.org/modules/web"
	"forgejo.org/services/context"
	"forgejo.org/services/contexttest"
	"forgejo.org/services/forms"
	"forgejo.org/services/pull"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderConversation(t *testing.T) {
	unittest.PrepareTestEnv(t)

	pr, _ := issues_model.GetPullRequestByID(db.DefaultContext, 2)
	_ = pr.LoadIssue(db.DefaultContext)
	_ = pr.Issue.LoadPoster(db.DefaultContext)
	_ = pr.Issue.LoadRepo(db.DefaultContext)

	require.NoError(t, pr.LoadHeadRepo(t.Context()))
	repo, err := git.OpenRepository(t.Context(), pr.HeadRepo.RepoPath())
	defer repo.Close()
	require.NoError(t, err)
	prHeadCommitID, err := repo.GetBranchCommitID(pr.HeadBranch)
	require.NoError(t, err)

	run := func(name string, cb func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder)) {
		t.Run(name, func(t *testing.T) {
			ctx, resp := contexttest.MockContext(t, "/")
			ctx.Render = templates.HTMLRenderer()
			contexttest.LoadUser(t, ctx, pr.Issue.PosterID)
			contexttest.LoadRepo(t, ctx, pr.BaseRepoID)
			contexttest.LoadGitRepo(t, ctx)
			defer ctx.Repo.GitRepo.Close()
			cb(t, ctx, resp)
		})
	}

	var preparedComment *issues_model.Comment
	run("prepare", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		comment, err := pull.CreateCodeComment(ctx, pr.Issue.Poster, ctx.Repo.GitRepo, pr.Issue,
			1, 0, "content", "", false, 0, pr.MergeBase,
			prHeadCommitID, nil)
		require.NoError(t, err)

		comment.Invalidated = true
		err = issues_model.UpdateCommentInvalidate(ctx, comment)
		require.NoError(t, err)

		preparedComment = comment
	})
	if !assert.NotNil(t, preparedComment) {
		return
	}
	run("diff with outdated", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		ctx.Data["ShowOutdatedComments"] = true
		renderConversation(ctx, preparedComment, "diff")
		assert.Contains(t, resp.Body.String(), `<div class="content comment-container"`)
	})
	run("diff without outdated", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		ctx.Data["ShowOutdatedComments"] = false
		renderConversation(ctx, preparedComment, "diff")
		// unlike gitea, Forgejo renders the conversation (with the "outdated" label)
		assert.Contains(t, resp.Body.String(), `repo.issues.review.outdated_description`)
	})
	run("timeline with outdated", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		ctx.Data["ShowOutdatedComments"] = true
		renderConversation(ctx, preparedComment, "timeline")
		assert.Contains(t, resp.Body.String(), `<div id="code-comments-`)
	})
	run("timeline is not affected by ShowOutdatedComments=false", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		ctx.Data["ShowOutdatedComments"] = false
		renderConversation(ctx, preparedComment, "timeline")
		assert.Contains(t, resp.Body.String(), `<div id="code-comments-`)
	})
	run("diff non-existing review", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		reviews, err := issues_model.FindReviews(db.DefaultContext, issues_model.FindReviewOptions{
			IssueID: 2,
		})
		require.NoError(t, err)
		for _, r := range reviews {
			require.NoError(t, issues_model.DeleteReview(db.DefaultContext, r))
		}
		ctx.Data["ShowOutdatedComments"] = true
		renderConversation(ctx, preparedComment, "diff")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.NotContains(t, resp.Body.String(), `status-page-500`)
	})
	run("timeline non-existing review", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		reviews, err := issues_model.FindReviews(db.DefaultContext, issues_model.FindReviewOptions{
			IssueID: 2,
		})
		require.NoError(t, err)
		for _, r := range reviews {
			require.NoError(t, issues_model.DeleteReview(db.DefaultContext, r))
		}
		ctx.Data["ShowOutdatedComments"] = true
		renderConversation(ctx, preparedComment, "timeline")
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.NotContains(t, resp.Body.String(), `status-page-500`)
	})

	// Test multi-line comment rendering
	var multiLineComment *issues_model.Comment
	run("prepare multi-line comment", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		comment, err := pull.CreateCodeComment(ctx, pr.Issue.Poster, ctx.Repo.GitRepo, pr.Issue,
			1, 2, "multi-line content", "", false, 0, pr.MergeBase,
			prHeadCommitID, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 2, comment.ExtraLinesCount)
		multiLineComment = comment
	})
	if !assert.NotNil(t, multiLineComment) {
		return
	}
	run("timeline multi-line comment renders", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		renderConversation(ctx, multiLineComment, "timeline")
		body := resp.Body.String()
		assert.Contains(t, body, `<div id="code-comments-`)
		assert.Contains(t, body, "multi-line content")
		// Verify the "Lines X-Y" label is rendered in the conversation header
		assert.Contains(t, body, "Lines ")
	})
	run("diff multi-line comment renders", func(t *testing.T, ctx *context.Context, resp *httptest.ResponseRecorder) {
		ctx.Data["ShowOutdatedComments"] = true
		renderConversation(ctx, multiLineComment, "diff")
		body := resp.Body.String()
		assert.Contains(t, body, `<div class="content comment-container"`)
		// Verify the conversation-holder has data-extra-lines-count attribute
		assert.Contains(t, body, `data-extra-lines-count="2"`)
	})
}

// TestCreateCodeCommentRejectsNegativeExtraLinesCount checks that the CreateCodeComment handler
// rejects a negative extra_lines_count with a 400 response, before reaching any comment creation.
func TestCreateCodeCommentRejectsNegativeExtraLinesCount(t *testing.T) {
	unittest.PrepareTestEnv(t)

	pr, err := issues_model.GetPullRequestByID(db.DefaultContext, 2)
	require.NoError(t, err)
	require.NoError(t, pr.LoadIssue(db.DefaultContext))

	ctx, resp := contexttest.MockContext(t, "/")
	contexttest.LoadUser(t, ctx, pr.Issue.PosterID)
	contexttest.LoadRepo(t, ctx, pr.BaseRepoID)
	ctx.SetParams(":index", strconv.FormatInt(pr.Issue.Index, 10))

	web.SetForm(ctx, &forms.CodeCommentForm{
		Origin:          "diff",
		Content:         "a comment",
		Side:            "proposed",
		Line:            1,
		TreePath:        "README.md",
		ExtraLinesCount: -1,
	})

	CreateCodeComment(ctx)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

// TestCreateCodeCommentRejectsExceedingMaxLines checks that the CreateCodeComment handler rejects a
// multi-line comment spanning more than setting.UI.MaxCodeCommentLines lines with a 400 response.
func TestCreateCodeCommentRejectsExceedingMaxLines(t *testing.T) {
	unittest.PrepareTestEnv(t)

	defer test.MockVariableValue(&setting.UI.MaxCodeCommentLines, 50)()

	pr, err := issues_model.GetPullRequestByID(db.DefaultContext, 2)
	require.NoError(t, err)
	require.NoError(t, pr.LoadIssue(db.DefaultContext))

	ctx, resp := contexttest.MockContext(t, "/")
	contexttest.LoadUser(t, ctx, pr.Issue.PosterID)
	contexttest.LoadRepo(t, ctx, pr.BaseRepoID)
	ctx.SetParams(":index", strconv.FormatInt(pr.Issue.Index, 10))

	web.SetForm(ctx, &forms.CodeCommentForm{
		Origin:          "diff",
		Content:         "a comment",
		Side:            "proposed",
		Line:            1,
		TreePath:        "README.md",
		ExtraLinesCount: 50, // range spans 51 lines > 50
	})

	CreateCodeComment(ctx)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}
