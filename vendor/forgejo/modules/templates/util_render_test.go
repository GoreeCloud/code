// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package templates

import (
	"context"
	"html/template"
	"testing"

	"forgejo.org/models/db"
	issues_model "forgejo.org/models/issues"
	org_model "forgejo.org/models/organization"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/modules/translation"

	"github.com/stretchr/testify/assert"
)

const testInput = `  space @mention-user  
/just/a/path.bin
https://example.com/file.bin
[local link](file.bin)
[remote link](https://example.com)
[[local link|file.bin]]
[[remote link|https://example.com]]
![local image](image.jpg)
![remote image](https://example.com/image.jpg)
[[local image|image.jpg]]
[[remote link|https://example.com/image.jpg]]
https://example.com/user/repo/compare/88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb#hash
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb pare
https://example.com/user/repo/commit/88fc37a3c0a4dda553bdcfc80c178a58247f42fb
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb mit
:+1:
mail@domain.com
@mention-user test
#123
  space
` + "`code :+1: #123 code`\n"

var testMetas = map[string]string{
	"user":     "user13",
	"repo":     "repo11",
	"repoPath": "../../tests/gitea-repositories-meta/user13/repo11.git/",
	"mode":     "comment",
}

func TestApostrophesInMentions(t *testing.T) {
	rendered := RenderMarkdownToHtml(t.Context(), "@mention-user's comment")
	assert.Equal(t, template.HTML("<p><a href=\"/mention-user\" class=\"mention\" rel=\"nofollow\">@mention-user</a>&#39;s comment</p>\n"), rendered)
}

func TestNonExistentUserMention(t *testing.T) {
	rendered := RenderMarkdownToHtml(t.Context(), "@ThisUserDoesNotExist @mention-user")
	assert.Equal(t, template.HTML("<p>@ThisUserDoesNotExist <a href=\"/mention-user\" class=\"mention\" rel=\"nofollow\">@mention-user</a></p>\n"), rendered)
}

func TestRenderCommitBody(t *testing.T) {
	type args struct {
		ctx   context.Context
		msg   string
		metas map[string]string
	}
	tests := []struct {
		name string
		args args
		want template.HTML
	}{
		{
			name: "multiple lines",
			args: args{
				ctx: t.Context(),
				msg: "first line\nsecond line",
			},
			want: "second line",
		},
		{
			name: "multiple lines with leading newlines",
			args: args{
				ctx: t.Context(),
				msg: "\n\n\n\nfirst line\nsecond line",
			},
			want: "second line",
		},
		{
			name: "multiple lines with trailing newlines",
			args: args{
				ctx: t.Context(),
				msg: "first line\nsecond line\n\n\n",
			},
			want: "second line",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, RenderCommitBody(tt.args.ctx, tt.args.msg, tt.args.metas), "RenderCommitBody(%v, %v, %v)", tt.args.ctx, tt.args.msg, tt.args.metas)
		})
	}

	expected := `/just/a/path.bin
<a href="https://example.com/file.bin" class="link">https://example.com/file.bin</a>
[local link](file.bin)
[remote link](<a href="https://example.com" class="link">https://example.com</a>)
[[local link|file.bin]]
[[remote link|<a href="https://example.com" class="link">https://example.com</a>]]
![local image](image.jpg)
![remote image](<a href="https://example.com/image.jpg" class="link">https://example.com/image.jpg</a>)
[[local image|image.jpg]]
[[remote link|<a href="https://example.com/image.jpg" class="link">https://example.com/image.jpg</a>]]
<a href="https://example.com/user/repo/compare/88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb#hash" class="compare"><code class="nohighlight">88fc37a3c0...12fc37a3c0 (hash)</code></a>
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb pare
<a href="https://example.com/user/repo/commit/88fc37a3c0a4dda553bdcfc80c178a58247f42fb" class="commit"><code class="nohighlight">88fc37a3c0</code></a>
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb mit
<span class="emoji" aria-label="thumbs up" data-alias="+1">👍</span>
<a href="mailto:mail@domain.com" class="mailto">mail@domain.com</a>
<a href="/mention-user" class="mention">@mention-user</a> test
<a href="/user13/repo11/issues/123" class="ref-issue">#123</a>
  space
` + "`code <span class=\"emoji\" aria-label=\"thumbs up\" data-alias=\"+1\">👍</span> <a href=\"/user13/repo11/issues/123\" class=\"ref-issue\">#123</a> code`"
	assert.EqualValues(t, expected, RenderCommitBody(t.Context(), testInput, testMetas))
}

func TestRenderCommitMessage(t *testing.T) {
	expected := `space <a href="/mention-user" class="mention">@mention-user</a>  `

	assert.EqualValues(t, expected, RenderCommitMessage(t.Context(), testInput, testMetas))
}

func TestRenderCommitMessageLinkSubject(t *testing.T) {
	expected := `<a href="https://example.com/link" class="default-link muted">space </a><a href="/mention-user" class="mention">@mention-user</a>`

	assert.EqualValues(t, expected, RenderCommitMessageLinkSubject(t.Context(), testInput, "https://example.com/link", testMetas))
}

func TestRenderIssueTitle(t *testing.T) {
	expected := `  space @mention-user  
/just/a/path.bin
https://example.com/file.bin
[local link](file.bin)
[remote link](https://example.com)
[[local link|file.bin]]
[[remote link|https://example.com]]
![local image](image.jpg)
![remote image](https://example.com/image.jpg)
[[local image|image.jpg]]
[[remote link|https://example.com/image.jpg]]
https://example.com/user/repo/compare/88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb#hash
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb pare
https://example.com/user/repo/commit/88fc37a3c0a4dda553bdcfc80c178a58247f42fb
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb mit
<span class="emoji" aria-label="thumbs up" data-alias="+1">👍</span>
mail@domain.com
@mention-user test
<a href="/user13/repo11/issues/123" class="ref-issue">#123</a>
  space
<code class="inline-code-block">code :+1: #123 code</code>
`
	assert.EqualValues(t, expected, RenderIssueTitle(t.Context(), testInput, testMetas))
}

func TestRenderRefIssueTitle(t *testing.T) {
	expected := `  space @mention-user  
/just/a/path.bin
https://example.com/file.bin
[local link](file.bin)
[remote link](https://example.com)
[[local link|file.bin]]
[[remote link|https://example.com]]
![local image](image.jpg)
![remote image](https://example.com/image.jpg)
[[local image|image.jpg]]
[[remote link|https://example.com/image.jpg]]
https://example.com/user/repo/compare/88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb#hash
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb pare
https://example.com/user/repo/commit/88fc37a3c0a4dda553bdcfc80c178a58247f42fb
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb mit
<span class="emoji" aria-label="thumbs up" data-alias="+1">👍</span>
mail@domain.com
@mention-user test
#123
  space
<code class="inline-code-block">code :+1: #123 code</code>
`
	assert.EqualValues(t, expected, RenderRefIssueTitle(t.Context(), testInput))
}

func TestRenderMarkdownToHtml(t *testing.T) {
	expected := `<p>space <a href="/mention-user" class="mention" rel="nofollow">@mention-user</a><br/>
/just/a/path.bin
<a href="https://example.com/file.bin" rel="nofollow">https://example.com/file.bin</a>
<a href="/file.bin" rel="nofollow">local link</a>
<a href="https://example.com" rel="nofollow">remote link</a>
<a href="/src/file.bin" rel="nofollow">local link</a>
<a href="https://example.com" rel="nofollow">remote link</a>
<a href="/image.jpg" target="_blank" rel="nofollow noopener"><img src="/image.jpg" alt="local image" loading="lazy"/></a>
<a href="https://example.com/image.jpg" target="_blank" rel="nofollow noopener"><img src="https://example.com/image.jpg" alt="remote image" loading="lazy"/></a>
<a href="/image.jpg" rel="nofollow"><img src="/image.jpg" title="local image" alt=""/></a>
<a href="https://example.com/image.jpg" rel="nofollow"><img src="https://example.com/image.jpg" title="remote link" alt=""/></a>
<a href="https://example.com/user/repo/compare/88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb#hash" rel="nofollow"><code>88fc37a3c0...12fc37a3c0 (hash)</code></a>
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb...12fc37a3c0a4dda553bdcfc80c178a58247f42fb pare
<a href="https://example.com/user/repo/commit/88fc37a3c0a4dda553bdcfc80c178a58247f42fb" rel="nofollow"><code>88fc37a3c0</code></a>
com 88fc37a3c0a4dda553bdcfc80c178a58247f42fb mit
<span class="emoji" aria-label="thumbs up" data-alias="+1">👍</span>
<a href="mailto:mail@domain.com" rel="nofollow">mail@domain.com</a>
<a href="/mention-user" class="mention" rel="nofollow">@mention-user</a> test
#123
space
<code>code :+1: #123 code</code></p>
`
	assert.EqualValues(t, expected, RenderMarkdownToHtml(t.Context(), testInput))
}

func TestRenderLabels(t *testing.T) {
	unittest.PrepareTestEnv(t)

	tr := &translation.MockLocale{}
	label := unittest.AssertExistsAndLoadBean(t, &issues_model.Label{ID: 1})
	labelScoped := unittest.AssertExistsAndLoadBean(t, &issues_model.Label{ID: 7})
	labelMalicious := unittest.AssertExistsAndLoadBean(t, &issues_model.Label{ID: 11})
	labelArchived := unittest.AssertExistsAndLoadBean(t, &issues_model.Label{ID: 12})

	ctx := NewContext(t.Context())
	ctx.Locale = tr

	rendered := RenderLabels(ctx, []*issues_model.Label{label}, "user2/repo1", false)
	assert.Contains(t, rendered, "user2/repo1/issues?labels=1")
	assert.Contains(t, rendered, ">label1<")
	assert.Contains(t, rendered, "data-tooltip-content='First label'")
	assert.Contains(t, rendered, "aria-description='First label'")
	rendered = RenderLabels(ctx, []*issues_model.Label{label}, "user2/repo1", true)
	assert.Contains(t, rendered, "user2/repo1/pulls?labels=1")
	assert.Contains(t, rendered, ">label1<")
	rendered = RenderLabels(ctx, []*issues_model.Label{labelScoped}, "user2/repo1", false)
	assert.Contains(t, rendered, "user2/repo1/issues?labels=7")
	assert.Contains(t, rendered, ">scope<")
	assert.Contains(t, rendered, ">label1<")
	rendered = RenderLabels(ctx, []*issues_model.Label{labelMalicious}, "user2/repo1", false)
	assert.Contains(t, rendered, "user2/repo1/issues?labels=11")
	assert.Contains(t, rendered, ">  &lt;script&gt;malicious&lt;/script&gt; <")
	assert.Contains(t, rendered, ">&#39;?&amp;<")
	assert.Contains(t, rendered, "data-tooltip-content='Malicious label &#39; &lt;script&gt;malicious&lt;/script&gt;'")
	assert.Contains(t, rendered, "aria-description='Malicious label &#39; &lt;script&gt;malicious&lt;/script&gt;'")
	rendered = RenderLabels(ctx, []*issues_model.Label{labelArchived}, "user2/repo1", false)
	assert.Contains(t, rendered, "user2/repo1/issues?labels=12")
	assert.Contains(t, rendered, ">archived label&lt;&gt;<")
}

func TestRenderUser(t *testing.T) {
	unittest.PrepareTestEnv(t)

	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 2})
	org := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 3})
	ghost := user_model.NewGhostUser()

	assert.Contains(t, RenderUser(db.DefaultContext, *user),
		"<a href='/user2' rel='nofollow'><strong>user2</strong></a>")
	assert.Contains(t, RenderUser(db.DefaultContext, *org),
		"<a href='/org3' rel='nofollow'><strong>org3</strong></a>")
	assert.Contains(t, RenderUser(db.DefaultContext, *ghost),
		"<strong>Ghost</strong>")

	defer test.MockVariableValue(&setting.UI.DefaultShowFullName, true)()
	assert.Contains(t, RenderUser(db.DefaultContext, *user),
		"<a href='/user2' rel='nofollow'><strong>&lt; U&lt;se&gt;r Tw&lt;o &gt; &gt;&lt;</strong></a>")
	assert.Contains(t, RenderUser(db.DefaultContext, *org),
		"<a href='/org3' rel='nofollow'><strong>&lt;&lt;&lt;&lt; &gt;&gt; &gt;&gt; &gt; &gt;&gt; &gt; &gt;&gt;&gt; &gt;&gt;</strong></a>")
	assert.Contains(t, RenderUser(db.DefaultContext, *ghost),
		"<strong>Ghost</strong>")
}

func TestRenderReviewRequest(t *testing.T) {
	unittest.PrepareTestEnv(t)

	target1 := issues_model.RequestReviewTarget{User: &user_model.User{ID: 1, Name: "user1", FullName: "User <One>"}}
	target2 := issues_model.RequestReviewTarget{Team: &org_model.Team{ID: 2, Name: "Team2", OrgID: 3}}
	target3 := issues_model.RequestReviewTarget{Team: org_model.NewGhostTeam()}
	assert.Contains(t, RenderReviewRequest(db.DefaultContext, []issues_model.RequestReviewTarget{target1, target2, target3}),
		"<a href='/user1' rel='nofollow'><strong>user1</strong></a>, "+
			"<a href='/org/org3/teams/Team2' rel='nofollow'><strong>Team2</strong></a>, "+
			"<strong>Ghost team</strong>")

	defer test.MockVariableValue(&setting.UI.DefaultShowFullName, true)()
	assert.Contains(t, RenderReviewRequest(db.DefaultContext, []issues_model.RequestReviewTarget{target1}),
		"<a href='/user1' rel='nofollow'><strong>User &lt;One&gt;</strong></a>")
}
