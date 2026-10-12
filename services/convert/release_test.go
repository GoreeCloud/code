// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	"testing"

	"forgejo.org/models/db"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelease_ToRelease(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	repo1 := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	release1 := unittest.AssertExistsAndLoadBean(t, &repo_model.Release{ID: 1})
	release1.LoadAttributes(db.DefaultContext)

	t.Run("Normal", func(t *testing.T) {
		apiRelease := ToAPIRelease(t.Context(), repo1, release1, false)
		assert.NotNil(t, apiRelease)
		assert.EqualValues(t, 1, apiRelease.ID)
		assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/releases/1", apiRelease.URL)
		assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/releases/1/assets", apiRelease.UploadURL)
	})

	t.Run("Github format", func(t *testing.T) {
		apiRelease := ToAPIRelease(t.Context(), repo1, release1, true)
		assert.NotNil(t, apiRelease)
		assert.EqualValues(t, 1, apiRelease.ID)
		assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/releases/1", apiRelease.URL)
		assert.Equal(t, "https://try.gitea.io/api/v1/repos/user2/repo1/releases/1/assets{?name,label}", apiRelease.UploadURL)
	})
}
