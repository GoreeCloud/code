// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package convert

import (
	"context"

	actions_model "forgejo.org/models/actions"
	access_model "forgejo.org/models/perm/access"
	user_model "forgejo.org/models/user"
	api "forgejo.org/modules/structs"
)

// ToActionRun convert actions_model.User to api.ActionRun
// the run needs all attributes loaded
func ToActionRun(ctx context.Context, run *actions_model.ActionRun, doer *user_model.User) *api.ActionRun {
	if run == nil {
		return nil
	}

	permissionInRepo, _ := access_model.GetUserRepoPermission(ctx, run.Repo, doer)

	return &api.ActionRun{
		ID:                run.ID,
		Title:             run.Title,
		Repo:              ToRepo(ctx, run.Repo, permissionInRepo),
		WorkflowID:        run.WorkflowID,
		Index:             run.Index,
		TriggerUser:       ToUser(ctx, run.TriggerUser, doer),
		ScheduleID:        run.ScheduleID,
		PrettyRef:         run.PrettyRef(),
		IsRefDeleted:      run.IsRefDeleted,
		CommitSHA:         run.CommitSHA,
		IsForkPullRequest: run.IsForkPullRequest,
		NeedApproval:      run.NeedApproval,
		ApprovedBy:        run.ApprovedBy,
		Event:             run.Event.Event(),
		EventPayload:      run.EventPayload,
		TriggerEvent:      run.TriggerEvent,
		Status:            run.Status.String(),
		Started:           run.Started.AsTime(),
		Stopped:           run.Stopped.AsTime(),
		Created:           run.Created.AsTime(),
		Updated:           run.Updated.AsTime(),
		Duration:          run.Duration(),
		HTMLURL:           run.HTMLURL(),
	}
}

// ToActionArtifact converts an AggregatedArtifact to an API ActionArtifact.
// repoAPIURL is the API URL prefix for the repository (e.g. from Repository.APIURL()).
func ToActionArtifact(repoAPIURL string, art *actions_model.AggregatedArtifact) *api.ActionArtifact {
	return &api.ActionArtifact{
		ID:                 art.ID,
		Name:               art.ArtifactName,
		SizeInBytes:        art.FileSize,
		ArchiveDownloadURL: art.APIDownloadURL(repoAPIURL),
		Expired:            art.Status == actions_model.ArtifactStatusExpired,
		RunID:              art.RunID,
		CreatedAt:          art.CreatedUnix.AsTime(),
		UpdatedAt:          art.UpdatedUnix.AsTime(),
		ExpiresAt:          art.ExpiredUnix.AsTime(),
	}
}

func ToActionRunJob(job *actions_model.ActionRunJob) *api.ActionRunJob {
	if job == nil {
		return nil
	}

	return &api.ActionRunJob{
		ID:      job.ID,
		RunID:   job.RunID,
		Attempt: job.Attempt,
		Handle:  job.Handle,
		RepoID:  job.RepoID,
		OwnerID: job.OwnerID,
		Name:    job.Name,
		Needs:   job.Needs,
		RunsOn:  job.RunsOn,
		TaskID:  job.TaskID,
		Status:  job.Status.String(),
	}
}
