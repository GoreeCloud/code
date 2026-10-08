// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package moderation

import (
	"testing"
	"time"

	"forgejo.org/models/db"
	moderation_model "forgejo.org/models/moderation"
	"forgejo.org/models/unittest"
	"forgejo.org/modules/timeutil"

	"github.com/stretchr/testify/require"
)

func TestRemoveResolvedReportsWhenNoTimeSet(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	// reportAge needs to be an int64 to match what timeutil.Day expects so we cast the value
	reportAge := int64(20)
	resolvedReport := &moderation_model.AbuseReport{
		Status:     moderation_model.ReportStatusTypeHandled,
		ReporterID: 1, ContentType: moderation_model.ReportedContentTypeRepository,
		ContentID: 2, Category: moderation_model.AbuseCategoryTypeOther,
		CreatedUnix:  timeutil.TimeStampNow(),
		ResolvedUnix: timeutil.TimeStamp(time.Now().Unix() - timeutil.Day*reportAge),
	}
	_, err := db.GetEngine(db.DefaultContext).NoAutoTime().Insert(resolvedReport)
	require.NoError(t, err)

	// No reports should be deleted when the default time to keep is 0
	err = RemoveResolvedReports(db.DefaultContext, time.Second*0)
	require.NoError(t, err)
	unittest.AssertExistsIf(t, true, resolvedReport)
}

func TestRemoveResolvedReportsWhenMatchTimeSet(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	// keepReportsFor needs to an int64 to match what timeutil.Day expects so we cast the value
	keepReportsFor := int64(4)
	resolvedReport := &moderation_model.AbuseReport{
		Status:     moderation_model.ReportStatusTypeHandled,
		ReporterID: 1, ContentType: moderation_model.ReportedContentTypeRepository,
		ContentID: 2, Category: moderation_model.AbuseCategoryTypeOther,
		CreatedUnix:  timeutil.TimeStampNow(),
		ResolvedUnix: timeutil.TimeStamp(time.Now().Unix() - timeutil.Day*keepReportsFor),
	}

	_, err := db.GetEngine(db.DefaultContext).NoAutoTime().Insert(resolvedReport)
	require.NoError(t, err)

	// Report should be deleted when older than the default time to keep
	err = RemoveResolvedReports(db.DefaultContext, time.Second*4)
	require.NoError(t, err)
	unittest.AssertExistsIf(t, false, resolvedReport)
}

func TestRemoveResolvedReportsWhenTimeSetButReportNew(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	resolvedReport := &moderation_model.AbuseReport{
		Status:     moderation_model.ReportStatusTypeHandled,
		ReporterID: 1, ContentType: moderation_model.ReportedContentTypeRepository,
		ContentID: 2, Category: moderation_model.AbuseCategoryTypeOther,
		CreatedUnix:  timeutil.TimeStampNow(),
		ResolvedUnix: timeutil.TimeStampNow(),
	}
	_, err := db.GetEngine(db.DefaultContext).NoAutoTime().Insert(resolvedReport)
	require.NoError(t, err)

	// Report should not be deleted when newer than the default time to keep
	err = RemoveResolvedReports(db.DefaultContext, time.Second*4)
	require.NoError(t, err)
	unittest.AssertExistsIf(t, true, resolvedReport)
}

func TestDoesNotRemoveOpenReports(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())
	// keepReportsFor needs to an int64 to match what timeutil.Day expects so we cast the value
	keepReportsFor := int64(4)
	resolvedReport := &moderation_model.AbuseReport{
		Status:     moderation_model.ReportStatusTypeOpen,
		ReporterID: 1, ContentType: moderation_model.ReportedContentTypeRepository,
		ContentID: 2, Category: moderation_model.AbuseCategoryTypeOther,
		CreatedUnix:  timeutil.TimeStampNow(),
		ResolvedUnix: timeutil.TimeStamp(time.Now().Unix() - timeutil.Day*keepReportsFor),
	}

	_, err := db.GetEngine(db.DefaultContext).NoAutoTime().Insert(resolvedReport)
	require.NoError(t, err)

	// Report should not be deleted when open
	// and older than the default time to keep
	err = RemoveResolvedReports(db.DefaultContext, time.Second*4)
	require.NoError(t, err)
	unittest.AssertExistsIf(t, true, resolvedReport)
}
