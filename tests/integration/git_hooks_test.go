// Copyright 2026 The Forgejo Authors c/o Codeberg e.V.. All rights reserved.
// SPDX-License-Identifier: MIT
package integration

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"

	"forgejo.org/models/auth"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/git"

	"github.com/stretchr/testify/require"
)

func TestCustomGitHooks(t *testing.T) {
	onApplicationRun(t, func(t *testing.T, u *url.URL) {
		repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
		owner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

		httpContext := NewAPITestContext(t, owner.Name, repo.Name, auth.AccessTokenScopeReadRepository)

		dstPath := t.TempDir()

		u.Path = httpContext.GitPath()
		u.User = url.UserPassword(owner.Name, userPassword)

		doGitClone(dstPath, u)(t)

		customHooksDir := path.Join(repo.RepoPath(), "hooks")

		hookNames := []string{"pre-receive", "update", "post-receive"}

		for _, hookName := range hookNames {
			customPath := path.Join(customHooksDir, hookName+".d")
			err := os.MkdirAll(customPath, 0o755)
			require.NoError(t, err)

			err = os.WriteFile(path.Join(customPath, "append-proof"), customGitHookTpl(hookName), 0x755)
			require.NoError(t, err)

			// The legacy, already existing gitea script might be there in the hooks directory in old installations,
			// here it's ensured that these scripts filtered out when custom hooks run
			err = os.WriteFile(path.Join(customPath, "gitea"), customGitHookGiteaTpl(), 0x755)
			require.NoError(t, err)
		}

		fd, err := os.Create(path.Join(dstPath, "hooks-test.txt"))
		require.NoError(t, err)

		err = fd.Close()
		require.NoError(t, err)

		_, _, err = git.NewCommand(git.DefaultContext, "checkout", "master").RunStdString(&git.RunOpts{Dir: dstPath})
		require.NoError(t, err)

		err = os.WriteFile(path.Join(dstPath, "hooks-test.txt"), []byte("test"), 0x644)
		require.NoError(t, err)

		_, _, err = git.NewCommand(git.DefaultContext, "add", "hooks-test.txt").RunStdString(&git.RunOpts{Dir: dstPath})
		require.NoError(t, err)

		_, _, err = git.NewCommand(git.DefaultContext, "commit", "-m", "Add hooks-test.txt").RunStdString(&git.RunOpts{Dir: dstPath})
		require.NoError(t, err)

		_, _, err = git.NewCommand(git.DefaultContext, "push", "origin", "master").RunStdString(&git.RunOpts{Dir: dstPath})
		require.NoError(t, err)

		data, err := os.ReadFile(path.Join(customHooksDir, "hooks-proof.txt"))
		require.NoError(t, err)

		require.Equal(t, `pre-receive
update
post-receive
`, string(data))
	})
}

func customGitHookTpl(hookName string) []byte {
	hookStr := fmt.Sprintf(`#!/usr/bin/env sh
echo "%s" >> $(dirname $0)/../hooks-proof.txt
`, hookName)

	return []byte(hookStr)
}

func customGitHookGiteaTpl() []byte {
	hookStr := `#!/usr/bin/env sh
echo "legacy gitea script shouldn't be called!"
exit 1
`

	return []byte(hookStr)
}
