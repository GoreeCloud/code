// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package feed

import (
	"context"
	"fmt"
	"path"
	"strings"

	activities_model "forgejo.org/models/activities"
	issues_model "forgejo.org/models/issues"
	repo_model "forgejo.org/models/repo"
	user_model "forgejo.org/models/user"
	"forgejo.org/modules/git"
	"forgejo.org/modules/json"
	"forgejo.org/modules/log"
	"forgejo.org/modules/repository"
	"forgejo.org/modules/setting"
	"forgejo.org/modules/util"
	federation_service "forgejo.org/services/federation"
	notify_service "forgejo.org/services/notify"
)

type actionNotifier struct {
	notify_service.NullNotifier
}

var _ notify_service.Notifier = &actionNotifier{}

func Init() error {
	notify_service.RegisterNotifier(NewNotifier())

	return nil
}

// NewNotifier create a new actionNotifier notifier
func NewNotifier() notify_service.Notifier {
	return &actionNotifier{}
}

func notifyAll(ctx context.Context, action *activities_model.Action) error {
	out, err := activities_model.NotifyWatchers(ctx, action)
	if err != nil {
		return err
	}
	return federation_service.NotifyActivityPubFollowers(ctx, out)
}

func notifyAllActions(ctx context.Context, acts []*activities_model.Action) error {
	out, err := activities_model.NotifyWatchersActions(ctx, acts)
	if err != nil {
		return err
	}
	return federation_service.NotifyActivityPubFollowers(ctx, out)
}

func (a *actionNotifier) NewIssue(ctx context.Context, issue *issues_model.Issue, mentions []*user_model.User) {
	if err := issue.LoadPoster(ctx); err != nil {
		log.Error("issue.LoadPoster: %v", err)
		return
	}
	if err := issue.LoadRepo(ctx); err != nil {
		log.Error("issue.LoadRepo: %v", err)
		return
	}
	repo := issue.Repo

	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: issue.Poster.ID,
		ActUser:   issue.Poster,
		OpType:    activities_model.ActionCreateIssue,
		Content:   encodeContent(fmt.Sprintf("%d", issue.Index), issue.Title),
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

// IssueChangeStatus notifies close or reopen issue to notifiers
func (a *actionNotifier) IssueChangeStatus(ctx context.Context, doer *user_model.User, commitID string, issue *issues_model.Issue, actionComment *issues_model.Comment, closeOrReopen bool) {
	// Compose comment action, could be plain comment, close or reopen issue/pull request.
	// This object will be used to notify watchers in the end of function.
	act := &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		Content:   encodeContent(fmt.Sprintf("%d", issue.Index), ""),
		RepoID:    issue.Repo.ID,
		Repo:      issue.Repo,
		Comment:   actionComment,
		CommentID: actionComment.ID,
		IsPrivate: issue.Repo.IsPrivate,
	}
	// Check comment type.
	if closeOrReopen {
		act.OpType = activities_model.ActionCloseIssue
		if issue.IsPull {
			act.OpType = activities_model.ActionClosePullRequest
		}
	} else {
		act.OpType = activities_model.ActionReopenIssue
		if issue.IsPull {
			act.OpType = activities_model.ActionReopenPullRequest
		}
	}

	// Notify watchers for whatever action comes in, ignore if no action type.
	if err := notifyAll(ctx, act); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

// CreateIssueComment notifies comment on an issue to notifiers
func (a *actionNotifier) CreateIssueComment(ctx context.Context, doer *user_model.User, repo *repo_model.Repository,
	issue *issues_model.Issue, comment *issues_model.Comment, mentions []*user_model.User,
) {
	act := &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		RepoID:    issue.Repo.ID,
		Repo:      issue.Repo,
		Comment:   comment,
		CommentID: comment.ID,
		IsPrivate: issue.Repo.IsPrivate,
		Content:   encodeContent(fmt.Sprintf("%d", issue.Index), abbreviatedComment(comment.Content)),
	}

	if issue.IsPull {
		act.OpType = activities_model.ActionCommentPull
	} else {
		act.OpType = activities_model.ActionCommentIssue
	}

	// Notify watchers for whatever action comes in, ignore if no action type.
	if err := notifyAll(ctx, act); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) NewPullRequest(ctx context.Context, pull *issues_model.PullRequest, mentions []*user_model.User) {
	if err := pull.LoadIssue(ctx); err != nil {
		log.Error("pull.LoadIssue: %v", err)
		return
	}
	if err := pull.Issue.LoadRepo(ctx); err != nil {
		log.Error("pull.Issue.LoadRepo: %v", err)
		return
	}
	if err := pull.Issue.LoadPoster(ctx); err != nil {
		log.Error("pull.Issue.LoadPoster: %v", err)
		return
	}

	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: pull.Issue.Poster.ID,
		ActUser:   pull.Issue.Poster,
		OpType:    activities_model.ActionCreatePullRequest,
		Content:   encodeContent(fmt.Sprintf("%d", pull.Issue.Index), pull.Issue.Title),
		RepoID:    pull.Issue.Repo.ID,
		Repo:      pull.Issue.Repo,
		IsPrivate: pull.Issue.Repo.IsPrivate,
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) RenameRepository(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, oldRepoName string) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    activities_model.ActionRenameRepo,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		Content:   oldRepoName,
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) TransferRepository(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, oldOwnerName string) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    activities_model.ActionTransferRepo,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		Content:   path.Join(oldOwnerName, repo.Name),
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) CreateRepository(ctx context.Context, doer, u *user_model.User, repo *repo_model.Repository) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    activities_model.ActionCreateRepo,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
	}); err != nil {
		log.Error("notify watchers '%d/%d': %v", doer.ID, repo.ID, err)
	}
}

func (a *actionNotifier) ForkRepository(ctx context.Context, doer *user_model.User, oldRepo, repo *repo_model.Repository) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    activities_model.ActionCreateRepo,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
	}); err != nil {
		log.Error("notify watchers '%d/%d': %v", doer.ID, repo.ID, err)
	}
}

func (a *actionNotifier) PullRequestReview(ctx context.Context, pr *issues_model.PullRequest, review *issues_model.Review, comment *issues_model.Comment, mentions []*user_model.User) {
	if err := review.LoadReviewer(ctx); err != nil {
		log.Error("LoadReviewer '%d/%d': %v", review.ID, review.ReviewerID, err)
		return
	}
	if err := review.LoadCodeComments(ctx); err != nil {
		log.Error("LoadCodeComments '%d/%d': %v", review.Reviewer.ID, review.ID, err)
		return
	}

	actions := make([]*activities_model.Action, 0, 10)
	for _, lines := range review.CodeComments {
		for _, comments := range lines {
			for _, comm := range comments {
				actions = append(actions, &activities_model.Action{
					ActUserID: review.Reviewer.ID,
					ActUser:   review.Reviewer,
					Content:   encodeContent(fmt.Sprintf("%d", review.Issue.Index), abbreviatedComment(comm.Content)),
					OpType:    activities_model.ActionCommentPull,
					RepoID:    review.Issue.RepoID,
					Repo:      review.Issue.Repo,
					IsPrivate: review.Issue.Repo.IsPrivate,
					Comment:   comm,
					CommentID: comm.ID,
				})
			}
		}
	}

	if review.Type != issues_model.ReviewTypeComment || strings.TrimSpace(comment.Content) != "" {
		action := &activities_model.Action{
			ActUserID: review.Reviewer.ID,
			ActUser:   review.Reviewer,
			Content:   encodeContent(fmt.Sprintf("%d", review.Issue.Index), abbreviatedComment(comment.Content)),
			RepoID:    review.Issue.RepoID,
			Repo:      review.Issue.Repo,
			IsPrivate: review.Issue.Repo.IsPrivate,
			Comment:   comment,
			CommentID: comment.ID,
		}

		switch review.Type {
		case issues_model.ReviewTypeApprove:
			action.OpType = activities_model.ActionApprovePullRequest
		case issues_model.ReviewTypeReject:
			action.OpType = activities_model.ActionRejectPullRequest
		default:
			action.OpType = activities_model.ActionCommentPull
		}

		actions = append(actions, action)
	}

	if err := notifyAllActions(ctx, actions); err != nil {
		log.Error("notify watchers '%d/%d': %v", review.Reviewer.ID, review.Issue.RepoID, err)
	}
}

func (*actionNotifier) MergePullRequest(ctx context.Context, doer *user_model.User, pr *issues_model.PullRequest) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    activities_model.ActionMergePullRequest,
		Content:   encodeContent(fmt.Sprintf("%d", pr.Issue.Index), pr.Issue.Title),
		RepoID:    pr.Issue.Repo.ID,
		Repo:      pr.Issue.Repo,
		IsPrivate: pr.Issue.Repo.IsPrivate,
	}); err != nil {
		log.Error("NotifyWatchers [%d]: %v", pr.ID, err)
	}
}

func (*actionNotifier) AutoMergePullRequest(ctx context.Context, doer *user_model.User, pr *issues_model.PullRequest) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    activities_model.ActionAutoMergePullRequest,
		Content:   encodeContent(fmt.Sprintf("%d", pr.Issue.Index), pr.Issue.Title),
		RepoID:    pr.Issue.Repo.ID,
		Repo:      pr.Issue.Repo,
		IsPrivate: pr.Issue.Repo.IsPrivate,
	}); err != nil {
		log.Error("NotifyWatchers [%d]: %v", pr.ID, err)
	}
}

func (*actionNotifier) PullReviewDismiss(ctx context.Context, doer *user_model.User, review *issues_model.Review, comment *issues_model.Comment) {
	if err := review.LoadReviewer(ctx); err != nil {
		log.Error("LoadReviewer '%d/%d': %v", review.ID, review.ReviewerID, err)
		return
	}
	reviewerName := review.Reviewer.Name
	if len(review.OriginalAuthor) > 0 {
		reviewerName = review.OriginalAuthor
	}
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    activities_model.ActionPullReviewDismissed,
		Content:   encodeContent(fmt.Sprintf("%d", review.Issue.Index), reviewerName, abbreviatedComment(comment.Content)),
		RepoID:    review.Issue.Repo.ID,
		Repo:      review.Issue.Repo,
		IsPrivate: review.Issue.Repo.IsPrivate,
		CommentID: comment.ID,
		Comment:   comment,
	}); err != nil {
		log.Error("NotifyWatchers [%d]: %v", review.Issue.ID, err)
	}
}

func (a *actionNotifier) PushCommits(ctx context.Context, pusher *user_model.User, repo *repo_model.Repository, opts *repository.PushUpdateOptions, commits *repository.PushCommits) {
	commits = prepareCommitsForFeed(commits)

	data, err := json.Marshal(commits)
	if err != nil {
		log.Error("Marshal: %v", err)
		return
	}

	opType := activities_model.ActionCommitRepo

	// Check it's tag push or branch.
	if opts.RefFullName.IsTag() {
		opType = activities_model.ActionPushTag
		if opts.IsDelRef() {
			opType = activities_model.ActionDeleteTag
		}
	} else if opts.IsDelRef() {
		opType = activities_model.ActionDeleteBranch
	}

	if err = notifyAll(ctx, &activities_model.Action{
		ActUserID: pusher.ID,
		ActUser:   pusher,
		OpType:    opType,
		Content:   string(data),
		RepoID:    repo.ID,
		Repo:      repo,
		RefName:   opts.RefFullName.String(),
		IsPrivate: repo.IsPrivate,
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) CreateRef(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, refFullName git.RefName, refID string) {
	opType := activities_model.ActionCommitRepo
	if refFullName.IsTag() {
		// has sent same action in `PushCommits`, so skip it.
		return
	}
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    opType,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		RefName:   refFullName.String(),
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) DeleteRef(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, refFullName git.RefName) {
	opType := activities_model.ActionDeleteBranch
	if refFullName.IsTag() {
		// has sent same action in `PushCommits`, so skip it.
		return
	}
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: doer.ID,
		ActUser:   doer,
		OpType:    opType,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		RefName:   refFullName.String(),
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) SyncPushCommits(ctx context.Context, pusher *user_model.User, repo *repo_model.Repository, opts *repository.PushUpdateOptions, commits *repository.PushCommits) {
	commits = prepareCommitsForFeed(commits)

	data, err := json.Marshal(commits)
	if err != nil {
		log.Error("json.Marshal: %v", err)
		return
	}

	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: repo.OwnerID,
		ActUser:   repo.MustOwner(ctx),
		OpType:    activities_model.ActionMirrorSyncPush,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		RefName:   opts.RefFullName.String(),
		Content:   string(data),
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) SyncCreateRef(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, refFullName git.RefName, refID string) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: repo.OwnerID,
		ActUser:   repo.MustOwner(ctx),
		OpType:    activities_model.ActionMirrorSyncCreate,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		RefName:   refFullName.String(),
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) SyncDeleteRef(ctx context.Context, doer *user_model.User, repo *repo_model.Repository, refFullName git.RefName) {
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: repo.OwnerID,
		ActUser:   repo.MustOwner(ctx),
		OpType:    activities_model.ActionMirrorSyncDelete,
		RepoID:    repo.ID,
		Repo:      repo,
		IsPrivate: repo.IsPrivate,
		RefName:   refFullName.String(),
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

func (a *actionNotifier) NewRelease(ctx context.Context, rel *repo_model.Release) {
	if err := rel.LoadAttributes(ctx); err != nil {
		log.Error("LoadAttributes: %v", err)
		return
	}
	if err := notifyAll(ctx, &activities_model.Action{
		ActUserID: rel.PublisherID,
		ActUser:   rel.Publisher,
		OpType:    activities_model.ActionPublishRelease,
		RepoID:    rel.RepoID,
		Repo:      rel.Repo,
		IsPrivate: rel.Repo.IsPrivate,
		Content:   rel.Title,
		RefName:   rel.TagName, // FIXME: use a full ref name?
	}); err != nil {
		log.Error("NotifyWatchers: %v", err)
	}
}

// ... later decoded in models/activities/action.go:GetIssueInfos
func encodeContent(params ...string) string {
	contentEncoded, err := json.Marshal(params)
	if err != nil {
		log.Error("encodeContent: Unexpected json encoding error: %v", err)
	}
	return string(contentEncoded)
}

// Given a comment of arbitrary-length Markdown text, create an abbreviated Markdown text appropriate for the
// activity feed.
func abbreviatedComment(comment string) string {
	firstLine := strings.Split(comment, "\n")[0]

	if strings.HasPrefix(firstLine, "```") {
		// First line is is a fenced code block... with no special abbreviate we would display a blank block, or in the
		// worst-case a ```mermaid would display an error. Better to omit the comment.
		return ""
	}

	truncatedContent, truncatedRight := util.SplitStringAtByteN(firstLine, 200)
	if truncatedRight != "" {
		// in case the content is in a Latin family language, we remove the last broken word.
		lastSpaceIdx := strings.LastIndex(truncatedContent, " ")
		if lastSpaceIdx != -1 && (len(truncatedContent)-lastSpaceIdx < 15) {
			truncatedContent = truncatedContent[:lastSpaceIdx] + "…"
		}
	}

	return truncatedContent
}

// Return a clone of the incoming repository.PushCommits that is appropriately tweaked for the activity feed. The struct
// is cloned rather than modified in-place because the same data will be sent to multiple notifiers. Transformations
// applied are: # of commits are limited to FeedMaxCommitNum, commit messages are trimmed to just the content displayed
// in the activity feed.
func prepareCommitsForFeed(commits *repository.PushCommits) *repository.PushCommits {
	numCommits := min(len(commits.Commits), setting.UI.FeedMaxCommitNum)
	retval := repository.PushCommits{
		Commits:    make([]*repository.PushCommit, 0, numCommits),
		HeadCommit: nil,
		CompareURL: commits.CompareURL,
		Len:        commits.Len,
	}
	if commits.HeadCommit != nil {
		retval.HeadCommit = prepareCommitForFeed(commits.HeadCommit)
	}
	for i, commit := range commits.Commits {
		if i == numCommits {
			break
		}
		retval.Commits = append(retval.Commits, prepareCommitForFeed(commit))
	}
	return &retval
}

func prepareCommitForFeed(commit *repository.PushCommit) *repository.PushCommit {
	return &repository.PushCommit{
		Sha1:           commit.Sha1,
		Message:        abbreviatedComment(commit.Message),
		AuthorEmail:    commit.AuthorEmail,
		AuthorName:     commit.AuthorName,
		CommitterEmail: commit.CommitterEmail,
		CommitterName:  commit.CommitterName,
		Signature:      commit.Signature,
		Verification:   commit.Verification,
		Timestamp:      commit.Timestamp,
	}
}
