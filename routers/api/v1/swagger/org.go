// Copyright 2017 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package swagger

import (
	api "forgejo.org/modules/structs"
)

// Organization
// swagger:response Organization
type swaggerResponseOrganization struct {
	// in:body
	Body api.Organization `json:"body"`
}

// OrganizationList
// swagger:response OrganizationList
type swaggerResponseOrganizationList struct {
	// in:body
	Body []api.Organization `json:"body"`

	// The total number of organizations
	TotalCount int64 `json:"X-Total-Count"`
}

// OrganizationListWithoutPagination - Organizations without pagination headers
// swagger:response OrganizationListWithoutPagination
type swaggerOrganizationListWithoutPagination struct {
	// in:body
	Body []api.Organization `json:"body"`
}

// Team
// swagger:response Team
type swaggerResponseTeam struct {
	// in:body
	Body api.Team `json:"body"`
}

// TeamList
// swagger:response TeamList
type swaggerResponseTeamList struct {
	// in:body
	Body []api.Team `json:"body"`

	// The total number of teams
	TotalCount int64 `json:"X-Total-Count"`
}

// TeamListWithoutPagination - Teams without pagination headers
// swagger:response TeamListWithoutPagination
type swaggerTeamListWithoutPagination struct {
	// in:body
	Body []api.Team `json:"body"`
}

// OrganizationPermissions
// swagger:response OrganizationPermissions
type swaggerResponseOrganizationPermissions struct {
	// in:body
	Body api.OrganizationPermissions `json:"body"`
}
