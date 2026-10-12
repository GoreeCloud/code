// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package repo

import (
	"net/http"
	"testing"

	"forgejo.org/models/unittest"
	"forgejo.org/modules/git"
	"forgejo.org/modules/test"
	"forgejo.org/modules/web"
	"forgejo.org/services/contexttest"
	"forgejo.org/services/forms"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetCommitNotes(t *testing.T) {
	unittest.PrepareTestEnv(t)
	commitID := "65f1bf27bc3bf70f64657658635e66094edbcb4d"
	path := "/user2/repo1/commit/" + commitID
	ctx, _ := contexttest.MockContext(t, path)
	ctx.SetParams(":sha", commitID)
	contexttest.LoadUser(t, ctx, 2)
	contexttest.LoadRepo(t, ctx, 1)
	contexttest.LoadGitRepo(t, ctx)
	notes := `This is a new note.\nSee https://frogejo.org.`
	web.SetForm(ctx, &forms.CommitNotesForm{
		Notes: notes,
	})
	SetCommitNotes(ctx)
	assert.Equal(t, http.StatusSeeOther, ctx.Resp.Status())
	assert.Equal(t, path, test.RedirectURL(ctx.Resp))
	note, err := git.GetNote(ctx, ctx.Repo.GitRepo, commitID)
	require.NoError(t, err)
	assert.Equal(t, []byte(notes+"\n"), note.Message)
}

func TestRemoveCommitNotes(t *testing.T) {
	unittest.PrepareTestEnv(t)
	commitID := "65f1bf27bc3bf70f64657658635e66094edbcb4d"
	path := "/user2/repo1/commit/" + commitID
	ctx, _ := contexttest.MockContext(t, path)
	ctx.SetParams(":sha", commitID)
	contexttest.LoadUser(t, ctx, 2)
	contexttest.LoadRepo(t, ctx, 1)
	contexttest.LoadGitRepo(t, ctx)
	RemoveCommitNotes(ctx)
	assert.Equal(t, http.StatusSeeOther, ctx.Resp.Status())
	assert.Equal(t, path, test.RedirectURL(ctx.Resp))
	note, err := git.GetNote(ctx, ctx.Repo.GitRepo, commitID)
	require.Error(t, err)
	assert.True(t, git.IsErrNotExist(err))
	assert.Nil(t, note)
}
