// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swagger

import (
	api "forgejo.org/modules/structs"
)

// NotificationThread
// swagger:response NotificationThread
type swaggerNotificationThread struct {
	// in:body
	Body api.NotificationThread `json:"body"`
}

// NotificationThreadList
// swagger:response NotificationThreadList
type swaggerNotificationThreadList struct {
	// in:body
	Body []api.NotificationThread `json:"body"`

	// The total number of notification threads
	TotalCount int64 `json:"X-Total-Count"`
}

// NotificationThreadListWithoutPagination - Notification threads without pagination headers
// swagger:response NotificationThreadListWithoutPagination
type swaggerNotificationThreadListWithoutPagination struct {
	// in:body
	Body []api.NotificationThread `json:"body"`
}

// Number of unread notifications
// swagger:response NotificationCount
type swaggerNotificationCount struct {
	// in:body
	Body api.NotificationCount `json:"body"`
}
