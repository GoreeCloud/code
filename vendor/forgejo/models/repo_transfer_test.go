// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package models

import (
	"testing"

	"forgejo.org/models/db"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPendingTransferIDs(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	doer := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 3})
	recipient := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})
	pendingTransfer := unittest.AssertExistsAndLoadBean(t, &RepoTransfer{RecipientID: recipient.ID, DoerID: doer.ID})

	pendingTransferIDs, err := GetPendingTransferIDs(db.DefaultContext, recipient.ID, doer.ID)
	require.NoError(t, err)
	if assert.Len(t, pendingTransferIDs, 1) {
		assert.Equal(t, pendingTransfer.ID, pendingTransferIDs[0])
	}
}
