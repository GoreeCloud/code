// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package moderation

import (
	"context"
	"fmt"
	"strings"

	"forgejo.org/models/db"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/timeutil"

	"xorm.io/builder"
)

type AbuseReportDetailed struct {
	AbuseReport        `xorm:"extends"`
	ReportedTimes      int // only for overview
	ReporterName       string
	ContentReference   string
	ShadowCopyDate     timeutil.TimeStamp // only for details
	ShadowCopyRawValue string             // only for details

	// In case the reported content was deleted before the report was handled, this flag can be set
	// in order to try to determine the abuser/poster ID (and load their details) based on the fields
	// that might have been saved within the shadow copy (for comments and issues/PRs).
	ShouldGetAbuserFromShadowCopy bool `xorm:"-"`
}

func (ard AbuseReportDetailed) ContentTypeIconName() string {
	switch ard.ContentType {
	case ReportedContentTypeUser:
		return "octicon-person"
	case ReportedContentTypeRepository:
		return "octicon-repo"
	case ReportedContentTypeIssue:
		return "octicon-issue-opened"
	case ReportedContentTypeComment:
		return "octicon-comment"
	default:
		return "octicon-question"
	}
}

func (ard AbuseReportDetailed) ContentURL() string {
	switch ard.ContentType {
	case ReportedContentTypeUser:
		return strings.TrimLeft(ard.ContentReference, "@")
	case ReportedContentTypeIssue:
		return strings.ReplaceAll(ard.ContentReference, "#", "/issues/")
	default:
		return ard.ContentReference
	}
}

func (ard AbuseReportDetailed) ReportedContentIsUser() bool {
	return ard.ContentType == ReportedContentTypeUser
}

func (ard AbuseReportDetailed) ReportedContentIsRepo() bool {
	return ard.ContentType == ReportedContentTypeRepository
}

func (ard AbuseReportDetailed) ReportedContentIsIssue() bool {
	return ard.ContentType == ReportedContentTypeIssue
}

func (ard AbuseReportDetailed) ReportedContentIsComment() bool {
	return ard.ContentType == ReportedContentTypeComment
}

func GetOpenReports(ctx context.Context) ([]*AbuseReportDetailed, error) {
	var reports []*AbuseReportDetailed

	// - For PostgreSQL user table name should be escaped.
	//   - Escaping can be done with double quotes (") but this doesn't work for MariaDB.
	// - For SQLite index column name should be escaped.
	//   - Escaping can be done with double quotes (") or backticks (`).
	// - For MariaDB/MySQL there is no need to escape the above.
	// - Therefore we will use double quotes (") but only for PostgreSQL and SQLite.
	// - Also, note that builder.Union() is broken: gitea.com/xorm/builder/issues/71
	identifierEscapeChar := ``
	if setting.Database.Type.IsPostgreSQL() || setting.Database.Type.IsSQLite3() {
		identifierEscapeChar = `"`
	}

	err := db.GetEngine(ctx).SQL(fmt.Sprintf(`SELECT AR.*, ARD.reported_times, U.name AS reporter_name, REFS.ref AS content_reference
		FROM abuse_report AR
		INNER JOIN (
			SELECT min(id) AS id, count(id) AS reported_times
			FROM abuse_report
			WHERE status = %[2]d
			GROUP BY content_type, content_id
		) ARD ON ARD.id = AR.id
		LEFT JOIN %[1]suser%[1]s U ON U.id = AR.reporter_id
		LEFT JOIN (
			SELECT %[3]d AS type, id, concat('@', name) AS "ref"
			FROM %[1]suser%[1]s WHERE id IN (
				SELECT content_id FROM abuse_report WHERE status = %[2]d AND content_type = %[3]d
			)
			UNION
			SELECT %[4]d AS "type", id, concat(owner_name, '/', name) AS "ref"
			FROM repository WHERE id IN (
				SELECT content_id FROM abuse_report WHERE status = %[2]d AND content_type = %[4]d
			)
			UNION
			SELECT %[5]d AS "type", I.id, concat(IR.owner_name, '/', IR.name, '#', I.%[1]sindex%[1]s) AS "ref"
			FROM issue I
			LEFT JOIN repository IR ON IR.id = I.repo_id
			WHERE I.id IN (
				SELECT content_id FROM abuse_report WHERE status = %[2]d AND content_type = %[5]d
			)
			UNION
			SELECT %[6]d AS "type", C.id, concat(CIR.owner_name, '/', CIR.name, '/issues/', CI.%[1]sindex%[1]s, '#issuecomment-', C.id) AS "ref"
			FROM comment C
			LEFT JOIN issue CI ON CI.id = C.issue_id
			LEFT JOIN repository CIR ON CIR.id = CI.repo_id
			WHERE C.id IN (
				SELECT content_id FROM abuse_report WHERE status = %[2]d AND content_type = %[6]d
			)
		) REFS ON REFS.type = AR.content_type AND REFS.id = AR.content_id
		ORDER BY AR.created_unix ASC`, identifierEscapeChar, ReportStatusTypeOpen,
		ReportedContentTypeUser, ReportedContentTypeRepository, ReportedContentTypeIssue, ReportedContentTypeComment)).
		Find(&reports)
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func GetOpenReportsByTypeAndContentID(ctx context.Context, contentType ReportedContentType, contentID int64) ([]*AbuseReportDetailed, error) {
	var reports []*AbuseReportDetailed

	// Some remarks concerning PostgreSQL:
	// - user table should be escaped (e.g. `user`);
	// - tried to use aliases for table names but errors like 'pq: invalid reference to FROM-clause entry'
	//   or 'pq: missing FROM-clause entry' were returned;
	err := db.GetEngine(ctx).
		Select("abuse_report.*, `user`.name AS reporter_name, abuse_report_shadow_copy.created_unix AS shadow_copy_date, abuse_report_shadow_copy.raw_value AS shadow_copy_raw_value").
		Table("abuse_report").
		Join("LEFT", "user", "`user`.id = abuse_report.reporter_id").
		Join("LEFT", "abuse_report_shadow_copy", "abuse_report_shadow_copy.id = abuse_report.shadow_copy_id").
		Where(builder.Eq{
			"content_type": contentType,
			"content_id":   contentID,
			"status":       ReportStatusTypeOpen,
		}).
		Asc("abuse_report.created_unix").
		Find(&reports)
	if err != nil {
		return nil, err
	}

	return reports, nil
}
