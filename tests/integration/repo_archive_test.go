// Copyright 2024 The Gitea Authors. All rights reserved.
// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	auth_model "forgejo.org/models/auth"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"
	"forgejo.org/routers"
	"forgejo.org/routers/web"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoDownloadArchive(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	defer test.MockVariableValue(&setting.EnableGzip, true)()
	defer test.MockVariableValue(&web.GzipMinSize, 10)()
	defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

	req := NewRequest(t, "GET", "/user2/repo1/archive/master.zip")
	req.Header.Set("Accept-Encoding", "gzip")
	resp := MakeRequest(t, req, http.StatusOK)
	bs, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Empty(t, resp.Header().Get("Content-Encoding"))
	assert.Len(t, bs, 320)

	// Verify that unrecognized archive type returns 404
	req = NewRequest(t, "GET", "/user2/repo1/archive/master.invalid")
	MakeRequest(t, req, http.StatusNotFound)
}

func TestRepoDownloadArchiveSubdir(t *testing.T) {
	onApplicationRun(t, func(*testing.T, *url.URL) {
		defer test.MockVariableValue(&setting.EnableGzip, true)()
		defer test.MockVariableValue(&web.GzipMinSize, 10)()
		defer test.MockVariableValue(&testWebRoutes, routers.NormalRoutes())()

		repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
		user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

		// Create a subdirectory
		err := createOrReplaceFileInBranch(user, repo, "subdir/test.txt", "master", "Test")
		require.NoError(t, err)

		t.Run("Frontend", func(t *testing.T) {
			resp := MakeRequest(t, NewRequestf(t, "GET", "/%s/src/branch/master/subdir", repo.FullName()), http.StatusOK)
			page := NewHTMLParser(t, resp.Body)

			page.AssertElement(t, fmt.Sprintf(".folder-actions a.archive-link[href='/%s/archive/master:subdir.zip'][type='application/zip']", repo.FullName()), true)
			page.AssertElement(t, fmt.Sprintf(".folder-actions a.archive-link[href='/%s/archive/master:subdir.tar.gz'][type='application/gzip']", repo.FullName()), true)
		})

		t.Run("Backend", func(t *testing.T) {
			resp := MakeRequest(t, NewRequestf(t, "GET", "/%s/archive/master:subdir.tar.gz", repo.FullName()), http.StatusOK)

			uncompressedStream, err := gzip.NewReader(resp.Body)
			require.NoError(t, err)

			tarReader := tar.NewReader(uncompressedStream)

			header, err := tarReader.Next()
			require.NoError(t, err)
			assert.Equal(t, tar.TypeDir, int32(header.Typeflag))
			assert.Equal(t, fmt.Sprintf("%s/", repo.Name), header.Name)

			header, err = tarReader.Next()
			require.NoError(t, err)
			assert.Equal(t, tar.TypeReg, int32(header.Typeflag))
			assert.Equal(t, fmt.Sprintf("%s/test.txt", repo.Name), header.Name)

			_, err = tarReader.Next()
			assert.Equal(t, io.EOF, err)
		})
	})
}

// Access under `/{username}/{repo}/archive/*` is permitted for API tokens.  Those API tokens then need to have the
// read:repository and the correct resource scopes to permit access, though.
func TestRepoDownloadArchiveViaAPITokens(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	t.Run("no read:repository scope", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		session := loginUser(t, "user2")
		allToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadMisc)

		t.Run("denied public repo1", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo1/archive/master.zip").AddTokenAuth(allToken)
			MakeRequest(t, req, http.StatusForbidden)
		})
		t.Run("denied private repo2", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo2/archive/master.zip").AddTokenAuth(allToken)
			MakeRequest(t, req, http.StatusForbidden)
		})
		// repo16 is a second repo used in fine-grain testing below, so we include it in other tests as a baseline
		t.Run("denied private repo16", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo16/archive/master.zip").AddTokenAuth(allToken)
			MakeRequest(t, req, http.StatusForbidden)
		})
	})

	t.Run("all access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()

		session := loginUser(t, "user2")
		allToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)

		t.Run("allowed public repo1", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo1/archive/master.zip").AddTokenAuth(allToken)
			resp := MakeRequest(t, req, http.StatusOK)
			assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
		})
		t.Run("allowed private repo2", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo2/archive/master.zip").AddTokenAuth(allToken)
			resp := MakeRequest(t, req, http.StatusOK)
			assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
		})
		// repo16 is a second repo used in fine-grain testing below, so we include it in other tests as a baseline
		t.Run("allowed private repo16", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo16/archive/master.zip").AddTokenAuth(allToken)
			resp := MakeRequest(t, req, http.StatusOK)
			assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
		})
	})

	t.Run("public-only access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		session := loginUser(t, "user2")
		publicOnlyToken := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopePublicOnly, auth_model.AccessTokenScopeReadRepository)

		t.Run("allowed public repo1", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo1/archive/master.zip").AddTokenAuth(publicOnlyToken)
			resp := MakeRequest(t, req, http.StatusOK)
			assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
		})
		t.Run("denied private repo2", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo2/archive/master.zip").AddTokenAuth(publicOnlyToken)
			MakeRequest(t, req, http.StatusNotFound)
		})
		t.Run("denied private repo16", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo16/archive/master.zip").AddTokenAuth(publicOnlyToken)
			MakeRequest(t, req, http.StatusNotFound)
		})
	})

	t.Run("specific repo access token", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		repo2OnlyToken := createFineGrainedRepoAccessToken(t, "user2",
			[]auth_model.AccessTokenScope{auth_model.AccessTokenScopeReadRepository},
			[]int64{2},
		)

		t.Run("allowed public repo1", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo1/archive/master.zip").AddTokenAuth(repo2OnlyToken)
			resp := MakeRequest(t, req, http.StatusOK)
			assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
		})
		t.Run("allowed inside fine-grain repo2", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo2/archive/master.zip").AddTokenAuth(repo2OnlyToken)
			resp := MakeRequest(t, req, http.StatusOK)
			assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
		})
		t.Run("denied private outside fine-grain repo16", func(t *testing.T) {
			defer tests.PrintCurrentTest(t)()
			req := NewRequest(t, "GET", "/user2/repo16/archive/master.zip").AddTokenAuth(repo2OnlyToken)
			MakeRequest(t, req, http.StatusNotFound)
		})
	})
}

func TestRepoDownloadArchiveViaAuthorizedIntegration(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	ait := newAITester(t, func(ai *auth_model.AuthorizedIntegration) {
		ai.Scope = auth_model.AccessTokenScopeReadRepository
	})
	defer ait.close()
	token := ait.signedJWT()

	// Clone of the "all access token" tests from TestRepoDownloadArchiveViaAPITokens -- not all test conditions are
	// repeated as there's no unique code in archive code paths for authorized integrations other than the
	// authentication method. Scopes and repo-specific reducers are common to both implementations.
	t.Run("allowed public repo1", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		req := NewRequest(t, "GET", "/user2/repo1/archive/master.zip").AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)
		assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
	})
	t.Run("allowed private repo2", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		req := NewRequest(t, "GET", "/user2/repo2/archive/master.zip").AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)
		assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
	})
	t.Run("allowed private repo16", func(t *testing.T) {
		defer tests.PrintCurrentTest(t)()
		req := NewRequest(t, "GET", "/user2/repo16/archive/master.zip").AddTokenAuth(token)
		resp := MakeRequest(t, req, http.StatusOK)
		assert.True(t, bytes.HasPrefix(resp.Body.Bytes(), []byte("PK")), "response body missing prefix PK")
	})
}
