// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRefURL(t *testing.T) {
	kases := []struct {
		refURL     string
		prefixURL  string
		parentPath string
		SSHDomain  string
		expect     string
	}{
		{"git://github.com/user1/repo1", "/", "user1/repo2", "", "http://github.com/user1/repo1"},
		{"https://localhost/user1/repo1.git", "/", "user1/repo2", "", "https://localhost/user1/repo1"},
		{"http://localhost/user1/repo1.git", "/", "owner/reponame", "", "http://localhost/user1/repo1"},
		{"git@github.com:user1/repo1.git", "/", "owner/reponame", "", "http://github.com/user1/repo1"},
		{"ssh://git@git.zefie.net:2222/zefie/lge_g6_kernel_scripts.git", "/", "zefie/lge_g6_kernel", "", "http://git.zefie.net/zefie/lge_g6_kernel_scripts"},
		{"git@git.zefie.net:2222/zefie/lge_g6_kernel_scripts.git", "/", "zefie/lge_g6_kernel", "", "http://git.zefie.net/2222/zefie/lge_g6_kernel_scripts"},
		{"git@try.gitea.io:go-gitea/gitea", "https://try.gitea.io/", "go-gitea/sdk", "", "https://try.gitea.io/go-gitea/gitea"},
		{"ssh://git@try.gitea.io:9999/go-gitea/gitea", "https://try.gitea.io/", "go-gitea/sdk", "", "https://try.gitea.io/go-gitea/gitea"},
		{"git://git@try.gitea.io:9999/go-gitea/gitea", "https://try.gitea.io/", "go-gitea/sdk", "", "https://try.gitea.io/go-gitea/gitea"},
		{"ssh://git@127.0.0.1:9999/go-gitea/gitea", "https://127.0.0.1:3000/", "go-gitea/sdk", "", "https://127.0.0.1:3000/go-gitea/gitea"},
		{"https://gitea.com:3000/user1/repo1.git", "https://127.0.0.1:3000/", "user/repo2", "", "https://gitea.com:3000/user1/repo1"},
		{"https://example.gitea.com/gitea/user1/repo1.git", "https://example.gitea.com/gitea/", "", "user/repo2", "https://example.gitea.com/gitea/user1/repo1"},
		{"https://username:password@github.com/username/repository.git", "/", "username/repository2", "", "https://username:password@github.com/username/repository"},
		{"somethingbad", "https://127.0.0.1:3000/go-gitea/gitea", "/", "", ""},
		{"git@localhost:user/repo", "https://localhost/", "user2/repo1", "", "https://localhost/user/repo"},
		{"../path/to/repo.git/", "https://localhost/", "user/repo2", "", "https://localhost/user/path/to/repo.git"},
		{"ssh://git@ssh.gitea.io:2222/go-gitea/gitea", "https://try.gitea.io/", "go-gitea/sdk", "ssh.gitea.io", "https://try.gitea.io/go-gitea/gitea"},
	}

	for _, kase := range kases {
		assert.Equal(t, kase.expect, getRefURL(kase.refURL, kase.prefixURL, kase.parentPath, kase.SSHDomain))
	}
}

func Test_parseSubmoduleContent(t *testing.T) {
	submoduleFiles := []struct {
		fileContent  string
		expectedPath string
		expected     Submodule
	}{
		{
			fileContent: `[submodule "jakarta-servlet"]
url = ../../ALP-pool/jakarta-servlet
path = jakarta-servlet`,
			expectedPath: "jakarta-servlet",
			expected: Submodule{
				Path: "jakarta-servlet",
				URL:  "../../ALP-pool/jakarta-servlet",
			},
		},
		{
			fileContent: `[submodule "jakarta-servlet"]
path = jakarta-servlet
url = ../../ALP-pool/jakarta-servlet`,
			expectedPath: "jakarta-servlet",
			expected: Submodule{
				Path: "jakarta-servlet",
				URL:  "../../ALP-pool/jakarta-servlet",
			},
		},
		{
			fileContent: `[submodule "about/documents"]
  path = about/documents
  url = git@github.com:example/documents.git
  branch = gh-pages
[submodule "custom-name"]
  path = manifesto
  url = https://github.com/example/manifesto.git
[submodule]
  path = relative/url
  url = ../such-relative.git
`,
			expectedPath: "relative/url",
			expected: Submodule{
				Path: "relative/url",
				URL:  "../such-relative.git",
			},
		},
		{
			fileContent: `# .gitmodules
# Subsection names are case sensitive
[submodule "Seanpm2001/Degoogle-your-life"]
	path = Its-time-to-cut-WideVine-DRM/DeGoogle-Your-Life/submodule.gitmodules
	url = https://github.com/seanpm2001/Degoogle-your-life/

[submodule "seanpm2001/degoogle-your-life"]
	url = https://github.com/seanpm2001/degoogle-your-life/
# This second section should not be merged with the first, because of casing
`,
			expectedPath: "Its-time-to-cut-WideVine-DRM/DeGoogle-Your-Life/submodule.gitmodules",
			expected: Submodule{
				Path: "Its-time-to-cut-WideVine-DRM/DeGoogle-Your-Life/submodule.gitmodules",
				URL:  "https://github.com/seanpm2001/Degoogle-your-life/",
			},
		},
	}
	for _, kase := range submoduleFiles {
		submodule, err := parseSubmoduleContent(strings.NewReader(kase.fileContent))
		require.NoError(t, err)
		v, ok := submodule[kase.expectedPath]
		assert.True(t, ok)
		assert.Equal(t, kase.expected, v)
	}
}
