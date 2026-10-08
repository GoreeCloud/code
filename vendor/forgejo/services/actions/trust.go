// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package actions

import (
	"context"
	"errors"
	"fmt"

	actions_model "forgejo.org/models/actions"
	issues_model "forgejo.org/models/issues"
	access_model "forgejo.org/models/perm/access"
	repo_model "forgejo.org/models/repo"
	unit_model "forgejo.org/models/unit"
	user_model "forgejo.org/models/user"
	actions_module "forgejo.org/modules/actions"
	"forgejo.org/modules/log"
	webhook_module "forgejo.org/modules/webhook"
)

type TrustUpdate string

const (
	UserTrustDenied   = TrustUpdate("deny")
	UserAlwaysTrusted = TrustUpdate("always")
	UserTrustedOnce   = TrustUpdate("once")
	UserTrustRevoked  = TrustUpdate("revoke")
)

func CleanupActionUser(ctx context.Context) error {
	return actions_model.RevokeInactiveActionUser(ctx)
}

func loadPullRequestAttributes(ctx context.Context, pr *issues_model.PullRequest) error {
	if err := pr.LoadIssue(ctx); err != nil {
		return err
	}

	return pr.Issue.LoadRepo(ctx)
}

func getIssuePoster(ctx context.Context, issue *issues_model.Issue) (*user_model.User, error) {
	if issue.Poster != nil {
		return issue.Poster, nil
	}
	if issue.PosterID == 0 {
		return nil, nil
	}

	poster, err := user_model.GetPossibleUserByID(ctx, issue.PosterID)
	if err != nil {
		if user_model.IsErrUserNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("getIssuePoster [%d]: %w", issue.PosterID, err)
	}
	issue.Poster = poster
	return poster, nil
}

func mustGetIssuePoster(ctx context.Context, issue *issues_model.Issue) (*user_model.User, error) {
	poster, err := getIssuePoster(ctx, issue)
	if err != nil {
		return nil, err
	}
	if poster == nil {
		return nil, user_model.ErrUserNotExist{UID: issue.PosterID}
	}
	return poster, nil
}

type useHeadOrBaseCommit int

const (
	useHeadCommit = 1 << iota
	useBaseCommit
)

func getPullRequestCommitAndApproval(ctx context.Context, pr *issues_model.PullRequest, doer *user_model.User, event webhook_module.HookEventType) (useHeadOrBaseCommit, actions_model.ApprovalType, error) {
	if pr == nil || actions_module.IsDefaultBranchWorkflow(event) || !pr.IsForkPullRequest() {
		return useHeadCommit, actions_model.DoesNotNeedApproval, nil
	}

	posterTrust, err := GetPullRequestPosterIsTrustedWithActions(ctx, pr)
	if err != nil {
		return useHeadCommit, actions_model.UndefinedApproval, err
	}

	if posterTrust.IsTrusted() {
		return useHeadCommit, actions_model.DoesNotNeedApproval, nil
	}

	doerTrust, err := getPullRequestUserIsTrustedWithActions(ctx, pr, doer)
	if err != nil {
		return useHeadCommit, actions_model.UndefinedApproval, err
	}

	if doerTrust.IsTrusted() {
		if event == webhook_module.HookEventPullRequestSync {
			// a synchronized event action (i.e. the doer pushed a commit to the pull request)
			// can run from the head
			return useHeadCommit, actions_model.DoesNotNeedApproval, nil
		}
		// other events run from workflows found in the base, not
		// from possibly modified workflows found in the head
		return useBaseCommit, actions_model.DoesNotNeedApproval, nil
	}
	// the poster and the doer are not trusted, approval is needed
	return useHeadCommit, actions_model.NeedApproval, nil
}

// cancels or approves runs and keep track of posters that are to always be trusted
func UpdateTrustedWithPullRequest(ctx context.Context, doerID int64, pr *issues_model.PullRequest, trusted TrustUpdate) error {
	if err := loadPullRequestAttributes(ctx, pr); err != nil {
		return err
	}

	switch trusted {
	case UserAlwaysTrusted:
		poster, err := mustGetIssuePoster(ctx, pr.Issue)
		if err != nil {
			return err
		}
		return AlwaysTrust(ctx, doerID, pr.Issue.RepoID, poster.ID)
	case UserTrustedOnce:
		return pullRequestApprove(ctx, doerID, pr.Issue.RepoID, pr.ID)
	case UserTrustRevoked:
		poster, err := mustGetIssuePoster(ctx, pr.Issue)
		if err != nil {
			return err
		}
		return RevokeTrust(ctx, pr.Issue.RepoID, poster.ID)
	case UserTrustDenied:
		return pullRequestCancel(ctx, pr.Issue.RepoID, pr.ID)
	default:
		return fmt.Errorf("UpdateTrustedWithPullRequest: unknown trust %v", trusted)
	}
}

func setRunTrustForPullRequest(ctx context.Context, run *actions_model.ActionRun, pr *issues_model.PullRequest, needApproval actions_model.ApprovalType) error {
	if pr == nil {
		return nil
	}

	if err := loadPullRequestAttributes(ctx, pr); err != nil {
		return err
	}

	run.IsForkPullRequest = pr.IsForkPullRequest()
	run.PullRequestPosterID = pr.Issue.PosterID
	run.PullRequestID = pr.ID
	run.NeedApproval = bool(needApproval)

	return nil
}

type UserTrust string

const (
	UserTrustIsNotRelevant             = UserTrust("irrelevant")
	UserIsNotTrustedWithActions        = UserTrust("no")
	UserIsExplicitlyTrustedWithActions = UserTrust("explicitly")
	UserIsImplicitlyTrustedWithActions = UserTrust("implicitly")
)

func (t UserTrust) IsTrusted() bool {
	return t != UserIsNotTrustedWithActions
}

func GetPullRequestPosterIsTrustedWithActions(ctx context.Context, pr *issues_model.PullRequest) (UserTrust, error) {
	if err := loadPullRequestAttributes(ctx, pr); err != nil {
		return "", err
	}
	poster, err := mustGetIssuePoster(ctx, pr.Issue)
	if err != nil {
		return UserIsNotTrustedWithActions, err
	}

	return getPullRequestUserIsTrustedWithActions(ctx, pr, poster)
}

func getPullRequestUserIsTrustedWithActions(ctx context.Context, pr *issues_model.PullRequest, user *user_model.User) (UserTrust, error) {
	if err := loadPullRequestAttributes(ctx, pr); err != nil {
		return "", err
	}

	return userIsTrustedWithPullRequest(ctx, pr, user)
}

func userIsTrustedWithPullRequest(ctx context.Context, pr *issues_model.PullRequest, user *user_model.User) (UserTrust, error) {
	implicitlyTrusted, err := userIsImplicitlyTrustedWithPullRequest(ctx, pr, user)
	if err != nil {
		return "", err
	}
	if implicitlyTrusted {
		log.Trace("%s is implicitly trusted to run actions in repository %s", user, pr.Issue.Repo)
		return UserIsImplicitlyTrustedWithActions, nil
	}

	explicitlyTrusted, err := userIsExplicitlyTrustedWithPullRequest(ctx, pr, user)
	if err != nil {
		return "", err
	}
	if explicitlyTrusted {
		log.Trace("%s is explicitly trusted to run actions in repository %s", user, pr.Issue.Repo)
		return UserIsExplicitlyTrustedWithActions, nil
	}

	log.Trace("%s is not trusted to run actions in repository %s", user, pr.Issue.Repo)
	return UserIsNotTrustedWithActions, nil
}

func userIsImplicitlyTrustedWithPullRequest(ctx context.Context, pr *issues_model.PullRequest, user *user_model.User) (bool, error) {
	// users that are trusted to create a pull request that is not from a fork
	// are also implicitly trusted to run workflows
	if !pr.IsForkPullRequest() {
		log.Trace("a pull request that is not from a fork nor AGit is implicitly trusted to run actions")
		return true, nil
	}

	return userCanWriteActionsOnRepo(ctx, pr.Issue.Repo, user)
}

func userCanWriteActionsOnRepo(ctx context.Context, repo *repo_model.Repository, user *user_model.User) (bool, error) {
	// users with write permission to the actions unit are trusted to
	// run actions
	permission, err := access_model.GetUserRepoPermission(ctx, repo, user)
	if err != nil {
		return false, err
	}
	if permission.CanWrite(unit_model.TypeActions) {
		log.Trace("%s has write permissions to the Action unit on %s", user, repo)
		return true, nil
	}

	return false, nil
}

func userIsExplicitlyTrustedWithPullRequest(ctx context.Context, pr *issues_model.PullRequest, user *user_model.User) (bool, error) {
	// there is no need to check if the user is blocked because it is not
	// allowed to create a pull request
	if user.IsRestricted {
		log.Trace("%v is restricted and cannot be trusted with pull requests", user)
		return false, nil
	}

	actionUser, err := actions_model.GetActionUserByUserIDAndRepoIDAndUpdateAccess(ctx, user.ID, pr.Issue.Repo.ID)
	if err != nil {
		log.Trace("%v is not explicitly trusted with pull requests on repository %v", user, pr.Issue.Repo)
		if actions_model.IsErrUserNotExist(err) {
			return false, nil
		}
		return false, err
	}

	log.Trace("%v is explicitly trusted with pull requests on repository %v", user, pr.Issue.Repo)
	return actionUser.TrustedWithPullRequests, nil
}

func RevokeTrust(ctx context.Context, repoID, posterID int64) error {
	if err := actions_model.DeleteActionUserByUserIDAndRepoID(ctx, posterID, repoID); err != nil {
		return err
	}

	runs, err := actions_model.GetRunsNotDoneByRepoIDAndPullRequestPosterID(ctx, repoID, posterID)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if err := CancelRun(ctx, run); err != nil {
			return err
		}
	}
	return nil
}

func AlwaysTrust(ctx context.Context, doerID, repoID, posterID int64) error {
	if err := actions_model.InsertActionUser(ctx, &actions_model.ActionUser{
		UserID:                  posterID,
		RepoID:                  repoID,
		TrustedWithPullRequests: true,
	}); err != nil {
		return err
	}

	runs, err := actions_model.GetRunsNotDoneByRepoIDAndPullRequestPosterID(ctx, repoID, posterID)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if err := ApproveRun(ctx, run, doerID); err != nil {
			return err
		}
	}
	return nil
}

func pullRequestCancel(ctx context.Context, repoID, pullRequestID int64) error {
	runs, err := actions_model.GetRunsNotDoneByRepoIDAndPullRequestID(ctx, repoID, pullRequestID)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if err := CancelRun(ctx, run); err != nil {
			return err
		}
	}
	return nil
}

func pullRequestApprove(ctx context.Context, doerID, repoID, pullRequestID int64) error {
	runs, err := actions_model.GetRunsThatNeedApprovalByRepoIDAndPullRequestID(ctx, repoID, pullRequestID)
	if err != nil {
		return err
	}

	for _, run := range runs {
		if err := ApproveRun(ctx, run, doerID); err != nil {
			return err
		}
	}
	return nil
}

func cleanupPullRequestUnapprovedRuns(ctx context.Context, repoID, pullRequestID int64) error {
	runs, err := actions_model.GetRunsThatNeedApprovalByRepoIDAndPullRequestID(ctx, repoID, pullRequestID)
	if err != nil {
		return err
	}

	errorSlice := []error{}
	for _, run := range runs {
		if err := CancelRun(ctx, run); err != nil {
			errorSlice = append(errorSlice, err)
		}
	}
	if len(errorSlice) > 0 {
		return errors.Join(errorSlice...)
	}
	return nil
}
