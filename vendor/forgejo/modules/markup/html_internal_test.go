// Copyright 2018 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors.
// SPDX-License-Identifier: MIT

package markup

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"forgejo.org/modules/git"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/modules/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestAppURL              = "http://localhost:3000/"
	TestOrgRepo             = "gogits/gogs"
	TestRepoURLWithoutSlash = TestAppURL + TestOrgRepo
	TestRepoURL             = TestAppURL + TestOrgRepo + "/"
)

// externalIssueLink an HTML link to an alphanumeric-style issue
func externalIssueLink(baseURL, class, name string) string {
	return link(util.URLJoin(baseURL, name), class, name)
}

// numericLink an HTML to a numeric-style issue
func numericIssueLink(baseURL, class string, index int, marker string) string {
	return link(util.URLJoin(baseURL, strconv.Itoa(index)), class, fmt.Sprintf("%s%d", marker, index))
}

// link an HTML link
func link(href, class, contents string) string {
	if class != "" {
		class = " class=\"" + class + "\""
	}

	return fmt.Sprintf("<a href=\"%s\"%s>%s</a>", href, class, contents)
}

var numericMetas = map[string]string{
	"format": "https://someurl.com/{user}/{repo}/{index}",
	"user":   "someUser",
	"repo":   "someRepo",
	"style":  IssueNameStyleNumeric,
}

var alphanumericMetas = map[string]string{
	"format": "https://someurl.com/{user}/{repo}/{index}",
	"user":   "someUser",
	"repo":   "someRepo",
	"style":  IssueNameStyleAlphanumeric,
}

var regexpMetas = map[string]string{
	"format": "https://someurl.com/{user}/{repo}/{index}",
	"user":   "someUser",
	"repo":   "someRepo",
	"style":  IssueNameStyleRegexp,
}

// these values should match the TestOrgRepo const above
var localMetas = map[string]string{
	"user": "gogits",
	"repo": "gogs",
}

func TestRender_IssueIndexPattern(t *testing.T) {
	// numeric: render inputs without valid mentions
	test := func(s string) {
		testRenderIssueIndexPattern(t, s, s, &RenderContext{
			Ctx: git.DefaultContext,
		})
		testRenderIssueIndexPattern(t, s, s, &RenderContext{
			Ctx:   git.DefaultContext,
			Metas: numericMetas,
		})
	}

	// should not render anything when there are no mentions
	test("")
	test("this is a test")
	test("test 123 123 1234")
	test("#")
	test("# # #")
	test("# 123")
	test("#abcd")
	test("test#1234")
	test("#1234test")
	test("#abcd")
	test("test!1234")
	test("!1234test")
	test(" test !1234test")
	test("/home/gitea/#1234")
	test("/home/gitea/!1234")

	// should not render issue mention without leading space
	test("test#54321 issue")

	// should not render issue mention without trailing space
	test("test #54321issue")
}

func TestRender_IssueIndexPattern2(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()

	// numeric: render inputs with valid mentions
	test := func(s, expectedFmt, marker string, indices ...int) {
		var path, prefix string
		isExternal := false
		if marker == "!" {
			path = "pulls"
			prefix = "http://localhost:3000/someUser/someRepo/pulls/"
		} else {
			path = "issues"
			prefix = "https://someurl.com/someUser/someRepo/"
			isExternal = true
		}

		links := make([]any, len(indices))
		for i, index := range indices {
			links[i] = numericIssueLink(util.URLJoin(TestRepoURL, path), "ref-issue", index, marker)
		}
		expectedNil := fmt.Sprintf(expectedFmt, links...)
		testRenderIssueIndexPattern(t, s, expectedNil, &RenderContext{
			Ctx:   git.DefaultContext,
			Metas: localMetas,
		})

		class := "ref-issue"
		if isExternal {
			class += " ref-external-issue"
		}

		for i, index := range indices {
			links[i] = numericIssueLink(prefix, class, index, marker)
		}
		expectedNum := fmt.Sprintf(expectedFmt, links...)
		testRenderIssueIndexPattern(t, s, expectedNum, &RenderContext{
			Ctx:   git.DefaultContext,
			Metas: numericMetas,
		})
	}

	// should render freestanding mentions
	test("#1234 test", "%s test", "#", 1234)
	test("test #8 issue", "test %s issue", "#", 8)
	test("!1234 test", "%s test", "!", 1234)
	test("test !8 issue", "test %s issue", "!", 8)
	test("test issue #1234", "test issue %s", "#", 1234)
	test("fixes issue #1234.", "fixes issue %s.", "#", 1234)

	// should render mentions in parentheses / brackets
	test("(#54321 issue)", "(%s issue)", "#", 54321)
	test("[#54321 issue]", "[%s issue]", "#", 54321)
	test("test (#9801 extra) issue", "test (%s extra) issue", "#", 9801)
	test("test (!9801 extra) issue", "test (%s extra) issue", "!", 9801)
	test("test (#1)", "test (%s)", "#", 1)

	// should render multiple issue mentions in the same line
	test("#54321 #1243", "%s %s", "#", 54321, 1243)
	test("wow (#54321 #1243)", "wow (%s %s)", "#", 54321, 1243)
	test("(#4)(#5)", "(%s)(%s)", "#", 4, 5)
	test("#1 (#4321) test", "%s (%s) test", "#", 1, 4321)

	// should render with :
	test("#1234: test", "%s: test", "#", 1234)
	test("wow (#54321: test)", "wow (%s: test)", "#", 54321)
}

func TestRender_IssueIndexPattern3(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()

	// alphanumeric: render inputs without valid mentions
	test := func(s string) {
		testRenderIssueIndexPattern(t, s, s, &RenderContext{
			Ctx:   git.DefaultContext,
			Metas: alphanumericMetas,
		})
	}
	test("")
	test("this is a test")
	test("test 123 123 1234")
	test("#")
	test("# 123")
	test("#abcd")
	test("test #123")
	test("abc-1234")         // issue prefix must be capital
	test("ABc-1234")         // issue prefix must be _all_ capital
	test("ABCDEFGHIJK-1234") // the limit is 10 characters in the prefix
	test("ABC1234")          // dash is required
	test("test ABC- test")   // number is required
	test("test -1234 test")  // prefix is required
	test("testABC-123 test") // leading space is required
	test("test ABC-123test") // trailing space is required
	test("ABC-0123")         // no leading zero
}

func TestRender_IssueIndexPattern4(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()

	// alphanumeric: render inputs with valid mentions
	test := func(s, expectedFmt string, names ...string) {
		links := make([]any, len(names))
		for i, name := range names {
			links[i] = externalIssueLink("https://someurl.com/someUser/someRepo/", "ref-issue ref-external-issue", name)
		}
		expected := fmt.Sprintf(expectedFmt, links...)
		testRenderIssueIndexPattern(t, s, expected, &RenderContext{
			Ctx:   git.DefaultContext,
			Metas: alphanumericMetas,
		})
	}
	test("OTT-1234 test", "%s test", "OTT-1234")
	test("test T-12 issue", "test %s issue", "T-12")
	test("test issue ABCDEFGHIJ-1234567890", "test issue %s", "ABCDEFGHIJ-1234567890")
}

func TestRender_IssueIndexPattern5(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()

	// regexp: render inputs without valid mentions
	test := func(s, expectedFmt, pattern string, ids, names []string) {
		metas := regexpMetas
		metas["regexp"] = pattern
		links := make([]any, len(ids))
		for i, id := range ids {
			links[i] = link(util.URLJoin("https://someurl.com/someUser/someRepo/", id), "ref-issue ref-external-issue", names[i])
		}

		expected := fmt.Sprintf(expectedFmt, links...)
		testRenderIssueIndexPattern(t, s, expected, &RenderContext{
			Ctx:   git.DefaultContext,
			Metas: metas,
		})
	}

	test("abc ISSUE-123 def", "abc %s def",
		"ISSUE-(\\d+)",
		[]string{"123"},
		[]string{"ISSUE-123"},
	)

	test("abc (ISSUE 123) def", "abc %s def",
		"\\(ISSUE (\\d+)\\)",
		[]string{"123"},
		[]string{"(ISSUE 123)"},
	)

	test("abc ISSUE-123 def", "abc %s def",
		"(ISSUE-(\\d+))",
		[]string{"ISSUE-123"},
		[]string{"ISSUE-123"},
	)

	testRenderIssueIndexPattern(t, "will not match", "will not match", &RenderContext{
		Ctx:   git.DefaultContext,
		Metas: regexpMetas,
	})
}

func TestRender_IssueIndexPattern_Document(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()
	metas := map[string]string{
		"format": "https://someurl.com/{user}/{repo}/{index}",
		"user":   "someUser",
		"repo":   "someRepo",
		"style":  IssueNameStyleNumeric,
		"mode":   "document",
	}

	testRenderIssueIndexPattern(t, "#1", "#1", &RenderContext{
		Ctx:   git.DefaultContext,
		Metas: metas,
	})
	testRenderIssueIndexPattern(t, "#1312", "#1312", &RenderContext{
		Ctx:   git.DefaultContext,
		Metas: metas,
	})
	testRenderIssueIndexPattern(t, "!1", "!1", &RenderContext{
		Ctx:   git.DefaultContext,
		Metas: metas,
	})
}

func testRenderIssueIndexPattern(t *testing.T, input, expected string, ctx *RenderContext) {
	ctx.Links.AbsolutePrefix = true
	if ctx.Links.Base == "" {
		ctx.Links.Base = TestRepoURL
	}

	var buf strings.Builder
	err := postProcess(ctx, []processor{issueIndexPatternProcessor}, strings.NewReader(input), &buf)
	require.NoError(t, err)
	assert.Equal(t, expected, buf.String(), "input=%q", input)
}

func TestRender_AutoLink(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()

	assert := func(input, expected, base string) {
		var buffer strings.Builder
		err := PostProcess(&RenderContext{
			Ctx: git.DefaultContext,
			Links: Links{
				Base: base,
			},
			Metas: localMetas,
		}, strings.NewReader(input), &buffer)
		require.NoError(t, err, nil)
		assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(buffer.String()))

		buffer.Reset()
		err = PostProcess(&RenderContext{
			Ctx: git.DefaultContext,
			Links: Links{
				Base: base,
			},
			Metas:  localMetas,
			IsWiki: true,
		}, strings.NewReader(input), &buffer)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(buffer.String()))
	}

	t.Run("Issue", func(t *testing.T) {
		// render valid issue URLs
		assert(util.URLJoin(TestRepoURL, "issues", "3333"),
			numericIssueLink(util.URLJoin(TestRepoURL, "issues"), "ref-issue", 3333, "#"),
			TestRepoURL)
	})

	t.Run("Commit", func(t *testing.T) {
		// render valid commit URLs
		tmp := util.URLJoin(TestRepoURL, "commit", "d8a994ef243349f321568f9e36d5c3f444b99cae")
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">d8a994ef24</code></a>", TestRepoURLWithoutSlash)
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">"+TestOrgRepo+"@d8a994ef24</code></a>", "/forgejo/forgejo")
		assert(
			tmp+"#diff-2",
			"<a href=\""+tmp+"#diff-2\" class=\"commit\"><code class=\"nohighlight\">d8a994ef24 (diff-2)</code></a>",
			TestRepoURL,
		)
		assert(
			tmp+"#diff-953bb4f01b7c77fa18f0cd54211255051e647dbc",
			"<a href=\""+tmp+"#diff-953bb4f01b7c77fa18f0cd54211255051e647dbc\" class=\"commit\"><code class=\"nohighlight\">d8a994ef24 (diff-953bb4f01b)</code></a>",
			TestRepoURLWithoutSlash,
		)

		// render other commit URLs
		tmp = "https://external-link.gitea.io/go-gitea/gitea/commit/d8a994ef243349f321568f9e36d5c3f444b99cae#diff-2"
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">external-link.gitea.io/go-gitea/gitea@d8a994ef24 (diff-2)</code></a>", TestOrgRepo)
		defer test.MockVariableValue(&setting.AppURL, "https://external-link.gitea.io/")()
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">d8a994ef24 (diff-2)</code></a>", "https://external-link.gitea.io/go-gitea/gitea")

		tmp = TestAppURL + "gogits/gogs/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20"
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">localhost:3000/gogits/gogs@190d949293</code></a>", "https://external-link.gitea.io/go-gitea/gitea")
		defer test.MockVariableValue(&setting.AppURL, TestAppURL)()
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">190d949293</code></a>", "http://localhost:3000/gogits/gogs")

		tmp = "http://localhost:3000/sub/gogits/gogs/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20"
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">localhost:3000/sub/gogits/gogs@190d949293</code></a>", TestRepoURLWithoutSlash)
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">localhost:3000/sub/gogits/gogs@190d949293</code></a>", "https://external-link.gitea.io/go-gitea/gitea")
		defer test.MockVariableValue(&setting.AppURL, TestAppURL+"sub/")()
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">190d949293</code></a>", "http://localhost:3000/sub/gogits/gogs")

		tmp = "http://localhost:3000/sub1/sub2/sub3/gogits/gogs/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20"
		defer test.MockVariableValue(&setting.AppURL, TestAppURL+"sub1/sub2/sub3/")()
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">190d949293</code></a>", "http://localhost:3000/sub1/sub2/sub3/gogits/gogs")
		defer test.MockVariableValue(&setting.AppURL, TestAppURL)()
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">localhost:3000/sub1/sub2/sub3/gogits/gogs@190d949293</code></a>", "http://localhost:3000/sub1/gogits/gogs")
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">localhost:3000/sub1/sub2/sub3/gogits/gogs@190d949293</code></a>", "https://external-link.gitea.io/go-gitea/gitea")

		// if the repository happens to be named like one of the known app routes (e.g. `src`),
		// we can parse the URL correctly, if there is no sub path
		tmp = "http://localhost:3000/gogits/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20"
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">gogits/src@190d949293</code></a>", TestRepoURL)
		tmp = "http://localhost:3000/gogits/src/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20"
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">gogits/src@190d949293</code></a>", TestRepoURL)
		// but if there is a sub path, we cannot reliably distinguish the repo name from the app route
		tmp = "http://localhost:3000/sub/gogits/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20"
		assert(tmp, "<a href=\""+tmp+"\" class=\"commit\"><code class=\"nohighlight\">sub/gogits@190d949293</code></a>", TestRepoURL)
	})

	t.Run("Compare", func(t *testing.T) {
		tmp := util.URLJoin(TestRepoURL, "compare", "d8a994ef243349f321568f9e36d5c3f444b99cae..190d9492934af498c3f669d6a2431dc5459e5b20")
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">d8a994ef24..190d949293</code></a>", TestRepoURL)
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">"+TestOrgRepo+"@d8a994ef24..190d949293</code></a>", "https://localhost/forgejo/forgejo")

		defer test.MockVariableValue(&setting.AppURL, TestAppURL+"sub/")()
		tmp = "http://localhost:3000/sub/gogits/gogs/compare/190d9492934af498c3f669d6a2431dc5459e5b20..d8a994ef243349f321568f9e36d5c3f444b99cae"
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">190d949293..d8a994ef24</code></a>", "http://localhost:3000/sub/gogits/gogs")
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">gogits/gogs@190d949293..d8a994ef24</code></a>", "http://localhost:3000/sub/gogits/gugs")
		defer test.MockVariableValue(&setting.AppURL, "https://external-link.gitea.io/")()
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">localhost:3000/sub/gogits/gogs@190d949293..d8a994ef24</code></a>", "https://external-link.gitea.io/go-gitea/gitea")

		defer test.MockVariableValue(&setting.AppURL, TestAppURL+"sub1/sub2/sub3/")()
		tmp = "http://localhost:3000/sub1/sub2/sub3/gogits/gogs/compare/190d9492934af498c3f669d6a2431dc5459e5b20..d8a994ef243349f321568f9e36d5c3f444b99cae"
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">190d949293..d8a994ef24</code></a>", "http://localhost:3000/sub1/sub2/sub3/gogits/gogs")
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">gogits/gogs@190d949293..d8a994ef24</code></a>", "/gogits/gous")
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">gogits/gogs@190d949293..d8a994ef24</code></a>", "https://external-link.gitea.io/go-gitea/gitea")

		tmp = "https://codeberg.org/forgejo/forgejo/compare/8bbac4c679bea930c74849c355a60ed3c52f8eb5...e2278e5a38187a1dc84dc41d583ec8b44e7257c1?files=options/locale/locale_fi-FI.ini"
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">codeberg.org/forgejo/forgejo@8bbac4c679...e2278e5a38 (options/locale/locale_fi-FI.ini)</code></a>", TestRepoURL)
		assert(tmp+".", "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">codeberg.org/forgejo/forgejo@8bbac4c679...e2278e5a38 (options/locale/locale_fi-FI.ini)</code></a>.", TestRepoURL)
		defer test.MockVariableValue(&setting.AppURL, "https://codeberg.org/")()
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">8bbac4c679...e2278e5a38 (options/locale/locale_fi-FI.ini)</code></a>", "https://codeberg.org/forgejo/forgejo")

		tmp = "https://codeberg.org/forgejo/forgejo/compare/8bbac4c679bea930c74849c355a60ed3c52f8eb5...e2278e5a38187a1dc84dc41d583ec8b44e7257c1?files=options/locale/locale_fi-FI.ini#L2"
		assert(tmp, "<a href=\""+tmp+"\" class=\"compare\"><code class=\"nohighlight\">8bbac4c679...e2278e5a38 (options/locale/locale_fi-FI.ini#L2)</code></a>", "https://codeberg.org/forgejo/forgejo")
	})

	t.Run("Invalid URLs", func(t *testing.T) {
		tmp := "https://local host/gogits/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20"
		assert(tmp, "<a href=\"https://local\" class=\"link\">https://local</a> host/gogits/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20", TestRepoURL)
	})
}

func TestRender_IssueIndexPatternRef(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()

	test := func(input, expected string) {
		var buf strings.Builder
		err := postProcess(&RenderContext{
			Ctx:   git.DefaultContext,
			Metas: numericMetas,
		}, []processor{issueIndexPatternProcessor}, strings.NewReader(input), &buf)
		require.NoError(t, err)
		assert.Equal(t, expected, buf.String(), "input=%q", input)
	}

	test("alan-turin/Enigma-cryptanalysis#1", `<a href="/alan-turin/enigma-cryptanalysis/issues/1" class="ref-issue">alan-turin/Enigma-cryptanalysis#1</a>`)
}

func TestRender_FullIssueURLs(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, TestAppURL)()

	test := func(input, expected string) {
		var result strings.Builder
		err := postProcess(&RenderContext{
			Ctx: git.DefaultContext,
			Links: Links{
				Base: TestRepoURL,
			},
			Metas: localMetas,
		}, []processor{fullIssuePatternProcessor}, strings.NewReader(input), &result)
		require.NoError(t, err)
		assert.Equal(t, expected, result.String())
	}
	test("Here is a link https://git.osgeo.org/gogs/postgis/postgis/pulls/6",
		"Here is a link https://git.osgeo.org/gogs/postgis/postgis/pulls/6")
	test("Look here http://localhost:3000/person/repo/issues/4",
		`Look here <a href="http://localhost:3000/person/repo/issues/4" class="ref-issue">person/repo#4</a>`)
	test("http://localhost:3000/person/repo/issues/4#issuecomment-1234",
		`<a href="http://localhost:3000/person/repo/issues/4#issuecomment-1234" class="ref-issue">person/repo#4 (comment)</a>`)
	test("http://localhost:3000/gogits/gogs/issues/4",
		`<a href="http://localhost:3000/gogits/gogs/issues/4" class="ref-issue">#4</a>`)
	test("http://localhost:3000/gogits/gogs/issues/4 test",
		`<a href="http://localhost:3000/gogits/gogs/issues/4" class="ref-issue">#4</a> test`)
	test("http://localhost:3000/gogits/gogs/issues/4?a=1&b=2#comment-form test",
		`<a href="http://localhost:3000/gogits/gogs/issues/4?a=1&amp;b=2#comment-form" class="ref-issue">#4</a> test`)
	test("http://localhost:3000/testOrg/testOrgRepo/pulls/2/files#issuecomment-24",
		`<a href="http://localhost:3000/testOrg/testOrgRepo/pulls/2/files#issuecomment-24" class="ref-issue">testOrg/testOrgRepo#2/files (comment)</a>`)
	test("http://localhost:3000/testOrg/testOrgRepo/pulls/2/commits",
		`<a href="http://localhost:3000/testOrg/testOrgRepo/pulls/2/commits" class="ref-issue">testOrg/testOrgRepo#2/commits</a>`)
}

func TestRegExp_hashCurrentPattern(t *testing.T) {
	trueTestCases := []string{
		"d8a994ef243349f321568f9e36d5c3f444b99cae",
		"abcdefabcdefabcdefabcdefabcdefabcdefabcd",
		"(abcdefabcdefabcdefabcdefabcdefabcdefabcd)",
		"[abcdefabcdefabcdefabcdefabcdefabcdefabcd]",
		"abcdefabcdefabcdefabcdefabcdefabcdefabcd.",
		"abcdefabcdefabcdefabcdefabcdefabcdefabcd:",
		"d8a994ef243349f321568f9e36d5c3f444b99cae12424fa123391042fbae2319",
		"abcdefd?",
		"abcdefd!",
		"!abcd3ef",
		":abcd3ef",
		".abcd3ef",
		" (abcd3ef). ",
		"abcd3ef...",
		"...abcd3ef",
		"(!...abcd3ef",
	}
	falseTestCases := []string{
		"test",
		"abcdefg",
		"e59ff077-2d03-4e6b-964d-63fbaea81f",
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmn",
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmO",
		"commit/abcdefd",
		"abcd3ef...defabcd",
		"f..defabcd",
	}

	for _, testCase := range trueTestCases {
		assert.True(t, hashCurrentPattern.MatchString(testCase))
	}
	for _, testCase := range falseTestCases {
		assert.False(t, hashCurrentPattern.MatchString(testCase))
	}
}

func TestRegExp_anySHA1Pattern(t *testing.T) {
	testCases := map[string][]string{
		"https://github.com/jquery/jquery/blob/a644101ed04d0beacea864ce805e0c4f86ba1cd1/test/unit/event.js#L2703": {
			"jquery/jquery/blob",
			"a644101ed04d0beacea864ce805e0c4f86ba1cd1",
			"/test/unit/event.js",
			"",
			"#L2703",
		},
		"https://github.com/jquery/jquery/blob/a644101ed04d0beacea864ce805e0c4f86ba1cd1/test/unit/event.js": {
			"jquery/jquery/blob",
			"a644101ed04d0beacea864ce805e0c4f86ba1cd1",
			"/test/unit/event.js",
			"",
			"",
		},
		"https://github.com/jquery/jquery/commit/0705be475092aede1eddae01319ec931fb9c65fc": {
			"jquery/jquery/commit",
			"0705be475092aede1eddae01319ec931fb9c65fc",
			"",
			"",
			"",
		},
		"https://github.com/jquery/jquery/tree/0705be475092aede1eddae01319ec931fb9c65fc/src": {
			"jquery/jquery/tree",
			"0705be475092aede1eddae01319ec931fb9c65fc",
			"/src",
			"",
			"",
		},
		"https://try.gogs.io/gogs/gogs/commit/d8a994ef243349f321568f9e36d5c3f444b99cae#diff-2": {
			"gogs/gogs/commit",
			"d8a994ef243349f321568f9e36d5c3f444b99cae",
			"",
			"",
			"#diff-2",
		},
		"https://codeberg.org/forgejo/forgejo/src/commit/949ab9a5c4cac742f84ae5a9fa186f8d6eb2cdc0/RELEASE-NOTES.md?display=source&w=1#L7-L9": {
			"forgejo/forgejo/src/commit",
			"949ab9a5c4cac742f84ae5a9fa186f8d6eb2cdc0",
			"/RELEASE-NOTES.md",
			"?display=source&w=1",
			"#L7-L9",
		},
		"http://localhost:3000/gogits/gogs/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20/path/to/file.go#L2-L3": {
			"gogits/gogs/src/commit",
			"190d9492934af498c3f669d6a2431dc5459e5b20",
			"/path/to/file.go",
			"",
			"#L2-L3",
		},
		"http://localhost:3000/sub/gogits/gogs/commit/190d9492934af498c3f669d6a2431dc5459e5b20/path/to/file.go#L2-L3": {
			"sub/gogits/gogs/commit",
			"190d9492934af498c3f669d6a2431dc5459e5b20",
			"/path/to/file.go",
			"",
			"#L2-L3",
		},
		"http://localhost:3000/sub/gogits/gogs/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20/path/to/file.go#L2-L3": {
			"sub/gogits/gogs/src/commit",
			"190d9492934af498c3f669d6a2431dc5459e5b20",
			"/path/to/file.go",
			"",
			"#L2-L3",
		},
		"http://localhost:3000/sub1/sub2/sub3/gogits/gogs/src/commit/190d9492934af498c3f669d6a2431dc5459e5b20/path/to/file.go#L2-L3": {
			"sub1/sub2/sub3/gogits/gogs/src/commit",
			"190d9492934af498c3f669d6a2431dc5459e5b20",
			"/path/to/file.go",
			"",
			"#L2-L3",
		},
	}

	for k, v := range testCases {
		assert.Equal(t, v, anyHashPattern.FindStringSubmatch(k)[1:])
	}

	for _, v := range []string{"https://codeberg.org/forgejo/forgejo/attachments/774421a1-b0ae-4501-8fba-983874b76811"} {
		assert.False(t, anyHashPattern.MatchString(v))
	}
}

func TestRegExp_shortLinkPattern(t *testing.T) {
	trueTestCases := []string{
		"[[stuff]]",
		"[[]]",
		"[[stuff|title=Difficult name with spaces*!]]",
	}
	falseTestCases := []string{
		"test",
		"abcdefg",
		"[[]",
		"[[",
		"[]",
		"]]",
		"abcdefghijklmnopqrstuvwxyz",
	}

	for _, testCase := range trueTestCases {
		assert.True(t, shortLinkPattern.MatchString(testCase))
	}
	for _, testCase := range falseTestCases {
		assert.False(t, shortLinkPattern.MatchString(testCase))
	}
}

func TestRender_escapeInlineCodeBlocks(t *testing.T) {
	test := func(input, expected string) {
		result := escapeInlineCodeBlocks(input)
		assert.Equal(t, expected, result)
	}
	test("`<test>`",
		"`&lt;test&gt;`")
	test("<test>",
		"<test>")
	test("`<foo>` <bar> `<baz>`",
		"`&lt;foo&gt;` <bar> `&lt;baz&gt;`")
	test("<foo> `<bar>` <baz>",
		"<foo> `&lt;bar&gt;` <baz>")
	test("<foo> `<bar> <baz>",
		"<foo> `<bar> <baz>")
	test("<foo> `<bar>` `<baz>",
		"<foo> `&lt;bar&gt;` `<baz>")
	test("<foo> `<bar>` `",
		"<foo> `&lt;bar&gt;` `")
	test("<foo> `<bar>` ``",
		"<foo> `&lt;bar&gt;` ``")
	test("```",
		"```")
	test("``<`",
		"``&lt;`")
}
