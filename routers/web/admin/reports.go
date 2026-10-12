// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package admin

import (
	"fmt"
	"net/http"

	"forgejo.org/models/issues"
	"forgejo.org/models/moderation"
	"forgejo.org/models/organization"
	repo_model "forgejo.org/models/repo"
	"forgejo.org/models/user"
	"forgejo.org/modules/base"
	"forgejo.org/modules/log"
	"forgejo.org/modules/optional"
	"forgejo.org/services/context"
	issue_service "forgejo.org/services/issue"
	moderation_service "forgejo.org/services/moderation"
	org_service "forgejo.org/services/org"
	repo_service "forgejo.org/services/repository"
	user_service "forgejo.org/services/user"
)

const (
	tplModerationReports       base.TplName = "admin/moderation/reports"
	tplModerationReportDetails base.TplName = "admin/moderation/report_details"
	tplAlert                   base.TplName = "base/alert"
)

// AbuseReports renders the reports overview page from admin moderation section.
func AbuseReports(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.moderation.reports")
	ctx.Data["PageIsAdminModerationReports"] = true

	reports, err := moderation.GetOpenReports(ctx)
	if err != nil {
		ctx.ServerError("Failed to load abuse reports", err)
		return
	}

	ctx.Data["Reports"] = reports
	ctx.Data["AbuseCategories"] = moderation.AbuseCategoriesTranslationKeys
	ctx.Data["GhostUserName"] = user.GhostUserName

	// available actions that can be done for reports
	ctx.Data["MarkAsHandled"] = int(moderation_service.ReportActionMarkAsHandled)
	ctx.Data["MarkAsIgnored"] = int(moderation_service.ReportActionMarkAsIgnored)

	// available actions that can be done for reported content
	ctx.Data["ActionSuspendAccount"] = int(moderation_service.ContentActionSuspendAccount)
	ctx.Data["ActionDeleteAccount"] = int(moderation_service.ContentActionDeleteAccount)
	ctx.Data["ActionDeleteRepo"] = int(moderation_service.ContentActionDeleteRepo)
	ctx.Data["ActionDeleteIssue"] = int(moderation_service.ContentActionDeleteIssue)
	ctx.Data["ActionDeleteComment"] = int(moderation_service.ContentActionDeleteComment)

	ctx.HTML(http.StatusOK, tplModerationReports)
}

// AbuseReportDetails renders a report details page opened from the reports overview from admin moderation section.
func AbuseReportDetails(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("admin.moderation.reports")
	ctx.Data["PageIsAdminModerationReports"] = true

	ctx.Data["Type"] = ctx.ParamsInt64(":type")
	ctx.Data["ID"] = ctx.ParamsInt64(":id")

	contentType := moderation.ReportedContentType(ctx.ParamsInt64(":type"))

	if !contentType.IsValid() {
		ctx.Flash.Error("Invalid content type")
		return
	}

	reports, err := moderation.GetOpenReportsByTypeAndContentID(ctx, contentType, ctx.ParamsInt64(":id"))
	if err != nil {
		ctx.ServerError("Failed to load reports", err)
		return
	}
	if len(reports) == 0 {
		// something is wrong
		ctx.HTML(http.StatusOK, tplModerationReportDetails)
		return
	}

	ctx.Data["Reports"] = reports
	ctx.Data["AbuseCategories"] = moderation.AbuseCategoriesTranslationKeys
	ctx.Data["GhostUserName"] = user.GhostUserName

	ctx.Data["GetShadowCopyMap"] = moderation_service.GetShadowCopyMap

	if err = setReportedContentDetails(ctx, reports[0]); err != nil {
		if user.IsErrUserNotExist(err) || issues.IsErrCommentNotExist(err) || issues.IsErrIssueNotExist(err) || repo_model.IsErrRepoNotExist(err) {
			ctx.Data["ContentReference"] = ctx.Tr("admin.moderation.deleted_content_ref", reports[0].ContentType, reports[0].ContentID)
			if contentType == moderation.ReportedContentTypeComment || contentType == moderation.ReportedContentTypeIssue {
				reports[0].ShouldGetAbuserFromShadowCopy = true
			}
		} else {
			ctx.ServerError("Failed to load reported content details", err)
			return
		}
	}

	ctx.HTML(http.StatusOK, tplModerationReportDetails)
}

// setReportedContentDetails adds some values into context data for the given report
// (icon name, a reference, the URL and in case of issues and comments also the poster name and URL).
func setReportedContentDetails(ctx *context.Context, report *moderation.AbuseReportDetailed) error {
	contentReference := ""
	var contentURL string
	var poster string
	var posterURL string
	contentType := report.ContentType
	contentID := report.ContentID

	ctx.Data["ContentTypeIconName"] = report.ContentTypeIconName()

	switch contentType {
	case moderation.ReportedContentTypeUser:
		reportedUser, err := user.GetUserByID(ctx, contentID)
		if err != nil {
			return err
		}

		contentReference = reportedUser.Name
		contentURL = reportedUser.HomeLink()
	case moderation.ReportedContentTypeRepository:
		repo, err := repo_model.GetRepositoryByID(ctx, contentID)
		if err != nil {
			return err
		}

		contentReference = repo.FullName()
		contentURL = repo.Link()
	case moderation.ReportedContentTypeIssue:
		issue, err := issues.GetIssueByID(ctx, contentID)
		if err != nil {
			return err
		}
		if err = issue.LoadRepo(ctx); err != nil {
			return err
		}
		if err = issue.LoadPoster(ctx); err != nil {
			return err
		}
		if issue.Poster != nil {
			poster = issue.Poster.Name
			posterURL = issue.Poster.HomeLink()
		}

		contentReference = fmt.Sprintf("%s#%d", issue.Repo.FullName(), issue.Index)
		contentURL = issue.Link()
	case moderation.ReportedContentTypeComment:
		comment, err := issues.GetCommentByID(ctx, contentID)
		if err != nil {
			return err
		}
		if err = comment.LoadIssue(ctx); err != nil {
			return err
		}
		if err = comment.Issue.LoadRepo(ctx); err != nil {
			return err
		}
		if err = comment.LoadPoster(ctx); err != nil && !user.IsErrUserNotExist(err) {
			return err
		}
		if comment.Poster != nil {
			poster = comment.Poster.Name
			posterURL = comment.Poster.HomeLink()
		}

		contentURL = comment.Link(ctx)
		contentReference = contentURL
	}

	ctx.Data["ContentReference"] = contentReference
	ctx.Data["ContentURL"] = contentURL
	ctx.Data["Poster"] = poster
	ctx.Data["PosterURL"] = posterURL
	return nil
}

func PerformAction(ctx *context.Context) {
	var contentID int64
	var contentType moderation.ReportedContentType

	contentID = ctx.FormInt64("content_id")
	if contentID <= 0 {
		ctx.Error(http.StatusBadRequest, "Invalid parameter: content_id")
		return
	}

	contentType = moderation.ReportedContentType(ctx.FormInt64("content_type"))
	if !contentType.IsValid() {
		ctx.Error(http.StatusBadRequest, "Invalid parameter: content_type")
		return
	}

	reportAction := moderation_service.ReportAction(ctx.FormInt64("report_action"))
	if !reportAction.IsValid() {
		ctx.Error(http.StatusBadRequest, "Invalid parameter: report_action")
		return
	}

	contentAction := moderation_service.ContentAction(ctx.FormInt64("content_action"))
	if !contentAction.IsValid() {
		ctx.Error(http.StatusBadRequest, "Invalid parameter: content_action")
		return
	}

	if contentAction == moderation_service.ContentActionNone && reportAction == moderation_service.ReportActionNone {
		ctx.Error(http.StatusBadRequest, "Invalid combination of content_action and report_action parameters")
		return
	}

	switch contentAction {
	case moderation_service.ContentActionNone:
		updateReportStatus(ctx, contentType, contentID, reportAction)
	case moderation_service.ContentActionSuspendAccount:
		suspendAccount(ctx, contentType, contentID, reportAction)
	case moderation_service.ContentActionDeleteAccount:
		deleteAccount(ctx, contentType, contentID, reportAction)
	case moderation_service.ContentActionDeleteRepo:
		deleteRepository(ctx, contentType, contentID, reportAction)
	case moderation_service.ContentActionDeleteIssue:
		deleteIssue(ctx, contentType, contentID, reportAction)
	case moderation_service.ContentActionDeleteComment:
		deleteComment(ctx, contentType, contentID, reportAction)
	default:
		ctx.Flash.Warning(ctx.Tr("moderation.unknown_action"), true)
		ctx.HTML(http.StatusOK, tplAlert)
	}
}

func updateReportStatus(ctx *context.Context, contentType moderation.ReportedContentType, contentID int64, reportAction moderation_service.ReportAction) {
	var err error

	switch reportAction {
	case moderation_service.ReportActionMarkAsHandled:
		err = moderation.MarkAsHandled(ctx, contentType, contentID)
	case moderation_service.ReportActionMarkAsIgnored:
		err = moderation.MarkAsIgnored(ctx, contentType, contentID)
	default:
		return
	}

	if err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to update the status of the report: %s", err.Error()))
		return
	}

	// TODO: translate and maybe use a more specific message (e.g. saying that the status was changed to 'Handled' or 'Ignored')?
	ctx.Flash.Success(fmt.Sprintf("Status updated for report(s) with type #%d and id #%d", contentType, contentID), true)
	ctx.HTML(http.StatusOK, tplAlert)
}

func suspendAccount(ctx *context.Context, contentType moderation.ReportedContentType, contentID int64, reportAction moderation_service.ReportAction) {
	if contentID == ctx.Doer.ID {
		ctx.Flash.Warning(ctx.Tr("moderation.users.cannot_suspend_self"), true)
		ctx.HTML(http.StatusOK, tplAlert)
		return
	}

	reportedUser, err := user.GetUserByID(ctx, contentID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve the user: %s", err.Error()))
		return
	}

	if reportedUser.IsAdmin {
		ctx.Flash.Warning(ctx.Tr("moderation.users.cannot_suspend_admins"), true)
		ctx.HTML(http.StatusOK, tplAlert)
		return
	}

	if reportedUser.IsOrganization() {
		ctx.Flash.Warning(ctx.Tr("moderation.users.cannot_suspend_org"), true)
		ctx.HTML(http.StatusOK, tplAlert)
		return
	}

	if reportedUser.ProhibitLogin {
		ctx.Flash.Info(ctx.Tr("moderation.users.already_suspended"), true)
		ctx.HTML(http.StatusOK, tplAlert)
		return
	}

	authOpts := &user_service.UpdateAuthOptions{ProhibitLogin: optional.Some(true)}
	// TODO: should we implement a new, simpler, SuspendAccount() method?!
	if err = user_service.UpdateAuth(ctx, reportedUser, authOpts); err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to suspend the user: %s", err.Error()))
		return
	}

	if reportAction != moderation_service.ReportActionNone {
		// TODO: currently not implemented
		updateReportStatus(ctx, contentType, contentID, reportAction)
	}

	ctx.Flash.Success(ctx.Tr("moderation.users.suspend_success"), true)
	ctx.HTML(http.StatusOK, tplAlert)
}

func deleteAccount(ctx *context.Context, contentType moderation.ReportedContentType, contentID int64, reportAction moderation_service.ReportAction) {
	if contentID == ctx.Doer.ID {
		ctx.Resp.Header().Add("HX-Reswap", "none") // prevent removing the report from the list
		ctx.Flash.Warning(ctx.Tr("admin.users.cannot_delete_self"), true)
		ctx.HTML(http.StatusOK, tplAlert)
		return
	}

	reportedUser, err := user.GetUserByID(ctx, contentID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve the user: %s", err.Error()))
		return
	}

	if reportedUser.IsAdmin {
		ctx.Resp.Header().Add("HX-Reswap", "none") // prevent removing the report from the list
		ctx.Flash.Warning(ctx.Tr("moderation.users.cannot_delete_admins"), true)
		ctx.HTML(http.StatusOK, tplAlert)
		return
	}

	if reportedUser.IsOrganization() {
		reportedOrg := organization.OrgFromUser(reportedUser)
		if err = org_service.DeleteOrganization(ctx, reportedOrg, true); err != nil {
			ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to delete the organization: %s", err.Error()))
			return
		}
		log.Trace("Organization deleted by admin (%s): %s", ctx.Doer.Name, reportedOrg.Name)
	} else {
		if err = user_service.DeleteUser(ctx, reportedUser, true); err != nil {
			ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to delete the user: %s", err.Error()))
			return
		}
		log.Trace("Account deleted by admin (%s): %s", ctx.Doer.Name, reportedUser.Name)
	}

	// TODO: when deleting content maybe we should always mark the reports as handled (does it makes sense to keep them open?!)
	updateReportStatus(ctx, contentType, contentID, reportAction) // TODO: combine success messages

	ctx.Flash.Success(ctx.Tr("admin.users.deletion_success"), true)
	ctx.HTML(http.StatusOK, tplAlert)
}

func deleteRepository(ctx *context.Context, contentType moderation.ReportedContentType, contentID int64, reportAction moderation_service.ReportAction) {
	repo, err := repo_model.GetRepositoryByID(ctx, contentID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve the repository: %s", err.Error()))
		return
	}

	if err = repo_service.DeleteRepository(ctx, ctx.Doer, repo, true); err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to delete the repository: %s", err.Error()))
		return
	}
	log.Trace("Repository deleted: %s", repo.FullName())

	updateReportStatus(ctx, contentType, contentID, reportAction)

	ctx.Flash.Success(ctx.Tr("repo.settings.deletion_success"), true)
	ctx.HTML(http.StatusOK, tplAlert)
}

func deleteIssue(ctx *context.Context, contentType moderation.ReportedContentType, contentID int64, reportAction moderation_service.ReportAction) {
	issue, err := issues.GetIssueByID(ctx, contentID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve the issue: %s", err.Error()))
		return
	}

	if err = issue_service.DeleteIssue(ctx, ctx.Doer, nil, issue); err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to delete the issue: %s", err.Error()))
		return
	}

	updateReportStatus(ctx, contentType, contentID, reportAction)

	ctx.Flash.Success(ctx.Tr("moderation.issue.deletion_success"), true)
	ctx.HTML(http.StatusOK, tplAlert)
}

func deleteComment(ctx *context.Context, contentType moderation.ReportedContentType, contentID int64, reportAction moderation_service.ReportAction) {
	comment, err := issues.GetCommentByID(ctx, contentID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve the comment: %s", err.Error()))
		return
	}

	if err = issue_service.DeleteComment(ctx, ctx.Doer, comment); err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to delete the comment: %s", err.Error()))
		return
	}

	updateReportStatus(ctx, contentType, contentID, reportAction)

	ctx.Flash.Success(ctx.Tr("moderation.comment.deletion_success"), true)
	ctx.HTML(http.StatusOK, tplAlert)
}
