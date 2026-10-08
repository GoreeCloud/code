// Copyright 2025 The Forgejo Contributors. All rights reserved.
// SPDX-License-Identifier: MIT

package source

import (
	"testing"

	"forgejo.org/models/db"
	quota_model "forgejo.org/models/quota"
	"forgejo.org/models/unittest"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/container"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncGroupsToQuotaGroupsCached(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	ctx := db.DefaultContext
	user := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 1})

	tests := []struct {
		name            string
		qgroups         container.Set[string]
		action          syncType
		setupGroups     []string
		setupMembers    map[string]bool
		expectedAdds    []string
		expectedRemoves []string
		expectedError   bool
	}{
		{
			name:        "syncAdd - user not in group should be added",
			qgroups:     container.SetOf("test-group-1"),
			action:      syncAdd,
			setupGroups: []string{"test-group-1"},
			setupMembers: map[string]bool{
				"test-group-1": false,
			},
			expectedAdds: []string{"test-group-1"},
		},
		{
			name:        "syncAdd - user already in group should not be added again",
			qgroups:     container.SetOf("test-group-2"),
			action:      syncAdd,
			setupGroups: []string{"test-group-2"},
			setupMembers: map[string]bool{
				"test-group-2": true,
			},
		},
		{
			name:        "syncRemove - user in group should be removed",
			qgroups:     container.SetOf("test-group-3"),
			action:      syncRemove,
			setupGroups: []string{"test-group-3"},
			setupMembers: map[string]bool{
				"test-group-3": true,
			},
			expectedRemoves: []string{"test-group-3"},
		},
		{
			name:        "syncRemove - user not in group should not cause error",
			qgroups:     container.SetOf("test-group-4"),
			action:      syncRemove,
			setupGroups: []string{"test-group-4"},
			setupMembers: map[string]bool{
				"test-group-4": false,
			},
		},
		{
			name:        "multiple groups - mixed operations",
			qgroups:     container.SetOf("test-group-5", "test-group-6"),
			action:      syncAdd,
			setupGroups: []string{"test-group-5", "test-group-6"},
			setupMembers: map[string]bool{
				"test-group-5": false,
				"test-group-6": true,
			},
			expectedAdds: []string{"test-group-5"},
		},
		{
			name:          "nonexistent group should log warning and continue",
			qgroups:       container.SetOf("nonexistent-group"),
			action:        syncAdd,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qgroupCache := make(map[string]*quota_model.Group)

			for _, groupName := range tt.setupGroups {
				group := &quota_model.Group{Name: groupName}
				_, err := db.GetEngine(ctx).Insert(group)
				require.NoError(t, err, "Failed to setup test group")

				if isMember, exists := tt.setupMembers[groupName]; exists && isMember {
					err = group.AddUserByID(ctx, user.ID)
					require.NoError(t, err, "Failed to setup initial membership")
				}
			}

			err := syncGroupsToQuotaGroupsCached(ctx, user, tt.qgroups, tt.action, qgroupCache)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			for _, groupName := range tt.expectedAdds {
				group := qgroupCache[groupName]
				if group == nil {
					group, err = quota_model.GetGroupByName(ctx, groupName)
					require.NoError(t, err)
				}
				isMember, err := group.IsUserInGroup(ctx, user.ID)
				require.NoError(t, err)
				assert.True(t, isMember, "User should be added to group %s", groupName)
			}

			for _, groupName := range tt.expectedRemoves {
				group := qgroupCache[groupName]
				if group == nil {
					group, err = quota_model.GetGroupByName(ctx, groupName)
					require.NoError(t, err)
				}
				isMember, err := group.IsUserInGroup(ctx, user.ID)
				require.NoError(t, err)
				assert.False(t, isMember, "User should be removed from group %s", groupName)
			}

			for _, groupName := range tt.setupGroups {
				unittest.AssertSuccessfulDelete(t, &quota_model.GroupMapping{
					Kind:      quota_model.KindUser,
					MappedID:  user.ID,
					GroupName: groupName,
				})
				unittest.AssertSuccessfulDelete(t, &quota_model.Group{Name: groupName})
			}
		})
	}
}
