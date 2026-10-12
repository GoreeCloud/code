// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "forgejo.org/models/auth"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIListActionArtifacts(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 4})
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})
	token := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeReadRepository)

	t.Run("ListRepoArtifacts", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)

		res := MakeRequest(t, req, http.StatusOK)
		var entries []*api.ActionArtifact
		DecodeJSON(t, res, &entries)

		// Fixture has 2 logical artifacts with confirmed status:
		// "multi-file-download" (run 791, ids 19+20) and "artifact-v4-download" (run 792, id 22)
		assert.Equal(t, "2", res.Header().Get("X-Total-Count"))
		require.Len(t, entries, 2)

		names := make([]string, len(entries))
		for i, a := range entries {
			names[i] = a.Name
			assert.False(t, a.Expired)
			assert.NotZero(t, a.SizeInBytes)
			assert.Contains(t, a.ArchiveDownloadURL, "/actions/artifacts/")
			assert.Contains(t, a.ArchiveDownloadURL, "/zip")
		}
		assert.ElementsMatch(t, []string{"multi-file-download", "artifact-v4-download"}, names)
	})

	t.Run("ListRepoArtifactsWithNameFilter", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts?name=multi-file-download", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)

		res := MakeRequest(t, req, http.StatusOK)
		var entries []*api.ActionArtifact
		DecodeJSON(t, res, &entries)

		assert.Equal(t, "1", res.Header().Get("X-Total-Count"))
		require.Len(t, entries, 1)
		assert.Equal(t, "multi-file-download", entries[0].Name)
		// multi-file-download has 2 rows of 1024 bytes each
		assert.Equal(t, int64(2048), entries[0].SizeInBytes)
	})

	t.Run("ListRunArtifacts", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/runs/791/artifacts", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)

		res := MakeRequest(t, req, http.StatusOK)
		var entries []*api.ActionArtifact
		DecodeJSON(t, res, &entries)

		// run 791 has only "multi-file-download" (id=1 is pending, not listed)
		assert.Equal(t, "1", res.Header().Get("X-Total-Count"))
		require.Len(t, entries, 1)
		assert.Equal(t, "multi-file-download", entries[0].Name)
	})

	t.Run("ListRunArtifactsNotFound", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/runs/99999/artifacts", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})
}

func TestAPIGetActionArtifact(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 4})
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})
	token := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeReadRepository)

	t.Run("GetV1V3Artifact", func(t *testing.T) {
		// id=19 is the MIN(id) for "multi-file-download" group
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/19", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)

		res := MakeRequest(t, req, http.StatusOK)
		var art api.ActionArtifact
		DecodeJSON(t, res, &art)

		assert.Equal(t, int64(19), art.ID)
		assert.Equal(t, "multi-file-download", art.Name)
		assert.Equal(t, int64(2048), art.SizeInBytes)
		assert.Equal(t, int64(791), art.RunID)
		assert.False(t, art.Expired)
	})

	t.Run("GetV4Artifact", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/22", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)

		res := MakeRequest(t, req, http.StatusOK)
		var art api.ActionArtifact
		DecodeJSON(t, res, &art)

		assert.Equal(t, int64(22), art.ID)
		assert.Equal(t, "artifact-v4-download", art.Name)
		assert.Equal(t, int64(1024), art.SizeInBytes)
		assert.Equal(t, int64(792), art.RunID)
		assert.False(t, art.Expired)
	})

	t.Run("GetNonCanonicalID", func(t *testing.T) {
		// id=20 is part of "multi-file-download" but is NOT the MIN(id), so it should 404
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/20", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("GetPendingArtifact", func(t *testing.T) {
		// id=1 has status=1 (upload pending), should not be accessible
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/1", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("GetNonExistent", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/99999", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("GetFromWrongRepository", func(t *testing.T) {
		// artifact id=22 belongs to user5/repo4; requesting it through a repo
		// the caller can access but that doesn't own the artifact must 404 —
		// this is the load-bearing check that caller-side RepoID was replaced
		// by a query-side constraint.
		otherRepo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
		otherUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: otherRepo.OwnerID})
		otherToken := getUserToken(t, otherUser.LowerName, auth_model.AccessTokenScopeReadRepository)

		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/22", otherRepo.OwnerName, otherRepo.Name),
		)
		req.AddTokenAuth(otherToken)
		MakeRequest(t, req, http.StatusNotFound)
	})
}

func TestAPIActionArtifactsRequireRepoScope(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 4})
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})
	wrongScopeToken := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeReadNotification)

	endpoints := []struct {
		name string
		path string
	}{
		{"list repo", fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts", repo.OwnerName, repo.Name)},
		{"list run", fmt.Sprintf("/api/v1/repos/%s/%s/actions/runs/791/artifacts", repo.OwnerName, repo.Name)},
		{"get", fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/19", repo.OwnerName, repo.Name)},
		{"download", fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/19/zip", repo.OwnerName, repo.Name)},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req := NewRequest(t, http.MethodGet, ep.path)
			req.AddTokenAuth(wrongScopeToken)
			MakeRequest(t, req, http.StatusForbidden)
		})
	}
}

func TestAPIDownloadActionArtifact(t *testing.T) {
	defer tests.PrepareTestEnv(t)()
	tests.PrepareArtifactsStorage(t)

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 4})
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})
	token := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeReadRepository)

	t.Run("DownloadV1V3Artifact", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/19/zip", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)

		res := MakeRequest(t, req, http.StatusOK)
		assert.Contains(t, res.Header().Get("Content-Disposition"), "multi-file-download.zip")
	})

	t.Run("DownloadV4Artifact", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/22/zip", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)

		res := MakeRequest(t, req, http.StatusOK)
		assert.Equal(t, "bytes", res.Header().Get("Accept-Ranges"))
	})

	t.Run("DownloadPendingArtifact", func(t *testing.T) {
		req := NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/1/zip", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(token)
		MakeRequest(t, req, http.StatusNotFound)
	})
}

func TestAPIDeleteActionArtifact(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 4})
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	t.Run("DeleteRequiresWritePermission", func(t *testing.T) {
		readToken := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeReadRepository)
		req := NewRequest(t, http.MethodDelete,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/22", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(readToken)
		MakeRequest(t, req, http.StatusForbidden)
	})

	t.Run("DeleteArtifact", func(t *testing.T) {
		writeToken := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeWriteRepository)
		req := NewRequest(t, http.MethodDelete,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/22", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(writeToken)
		MakeRequest(t, req, http.StatusNoContent)

		// Verify the artifact is no longer accessible
		readToken := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeReadRepository)
		req = NewRequest(t, http.MethodGet,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/22", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(readToken)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("DeleteNonExistent", func(t *testing.T) {
		writeToken := getUserToken(t, user.LowerName, auth_model.AccessTokenScopeWriteRepository)
		req := NewRequest(t, http.MethodDelete,
			fmt.Sprintf("/api/v1/repos/%s/%s/actions/artifacts/99999", repo.OwnerName, repo.Name),
		)
		req.AddTokenAuth(writeToken)
		MakeRequest(t, req, http.StatusNotFound)
	})
}
