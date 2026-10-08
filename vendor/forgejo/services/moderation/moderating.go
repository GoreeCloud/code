// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package moderation

import (
	"slices"

	"forgejo.org/models/issues"
	"forgejo.org/models/moderation"
	"forgejo.org/models/repo"
	"forgejo.org/models/user"
	"forgejo.org/modules/json"
	"forgejo.org/modules/log"
	"forgejo.org/services/context"
)

type ReportAction int

const (
	ReportActionNone ReportAction = iota
	ReportActionMarkAsHandled
	ReportActionMarkAsIgnored
)

var allReportActions = []ReportAction{
	ReportActionNone,
	ReportActionMarkAsHandled,
	ReportActionMarkAsIgnored,
}

func (ra ReportAction) IsValid() bool {
	return slices.Contains(allReportActions, ra)
}

type ContentAction int

const (
	// ContentActionNone means that no action should be done for the reported content itself;
	// is should be used when needed to just update the status of the report.
	ContentActionNone ContentAction = iota
	ContentActionSuspendAccount
	ContentActionDeleteAccount
	ContentActionDeleteRepo
	ContentActionDeleteIssue
	ContentActionDeleteComment
)

var allContentActions = []ContentAction{
	ContentActionNone,
	ContentActionSuspendAccount,
	ContentActionDeleteAccount,
	ContentActionDeleteRepo,
	ContentActionDeleteIssue,
	ContentActionDeleteComment,
}

func (ca ContentAction) IsValid() bool {
	return slices.Contains(allContentActions, ca)
}

// GetShadowCopyMap unmarshals the shadow copy raw value of the given abuse report and returns a list of <key, value> pairs
// (to be rendered when the report is reviewed by an admin).
// It also checks whether the ShouldGetAbuserFromShadowCopy runtime flag is set on the report and if so will try to
// retrieve the abusive user (when their ID can be found within the shadow copy) in order to set some details
// (name and profile link) as context data.
// If the report does not have a shadow copy ID or the raw value is empty, returns nil.
// If the unmarshal fails a warning is added in the logs and returns nil.
func GetShadowCopyMap(ctx *context.Context, ard *moderation.AbuseReportDetailed) []moderation.ShadowCopyField {
	if ard.ShadowCopyID.Valid && len(ard.ShadowCopyRawValue) > 0 {
		var data moderation.ShadowCopyData

		switch ard.ContentType {
		case moderation.ReportedContentTypeUser:
			data = new(user.UserData)
		case moderation.ReportedContentTypeRepository:
			data = new(repo.RepositoryData)
		case moderation.ReportedContentTypeIssue:
			data = new(issues.IssueData)
		case moderation.ReportedContentTypeComment:
			data = new(issues.CommentData)
		}
		if err := json.Unmarshal([]byte(ard.ShadowCopyRawValue), &data); err != nil {
			log.Warn("Unmarshal failed for shadow copy #%d. %v", ard.ShadowCopyID.Int64, err)
			return nil
		}

		if ard.ShouldGetAbuserFromShadowCopy {
			abuserID, isValidID := data.GetAbuserID()
			if isValidID {
				setAbuserDetails(ctx, abuserID)
			}
		}

		return data.GetFieldsMap()
	}
	return nil
}

// setAbuserDetails tries to retrieve a user with the given ID and in case
// a user is found it will set their name and profile URL into ctx.Data.
func setAbuserDetails(ctx *context.Context, abuserID int64) {
	abuser, err := user.GetPossibleUserByID(ctx, abuserID)
	if err == nil {
		ctx.Data["AbuserName"] = abuser.Name
		ctx.Data["AbuserURL"] = abuser.HomeLink()
	}
}
