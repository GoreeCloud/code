// Copyright Earl Warren <contact@earl-warren.org>
// Copyright Loïc Dachary <loic@dachary.org>
// SPDX-License-Identifier: MIT

package driver

import (
	"context"
	"fmt"

	"forgejo.org/models/db"
	issues_model "forgejo.org/models/issues"

	f3_tree "code.forgejo.org/f3/gof3/v3/tree/f3"
	f3_tree_generic "code.forgejo.org/f3/gof3/v3/tree/generic"
)

type comments struct {
	container
}

func (o *comments) ListPage(ctx context.Context, node f3_tree_generic.NodeInterface, _ f3_tree_generic.ListOptions, page int) f3_tree_generic.ChildrenList {
	pageSize := o.getPageSize()

	project := f3_tree.GetProjectID(node)
	commentable := f3_tree.GetCommentableID(node)

	issue, err := issues_model.GetIssueByIndex(ctx, project, commentable)
	if err != nil {
		panic(fmt.Errorf("GetIssueByIndex %v %w", commentable, err))
	}

	sess := db.GetEngine(ctx).
		Table("comment").
		Where("`issue_id` = ? AND `type` = ?", issue.ID, issues_model.CommentTypeComment)
	if page != 0 {
		sess = db.SetSessionPagination(sess, &db.ListOptions{Page: page, PageSize: pageSize})
	}
	forgejoComments := make([]*issues_model.Comment, 0, pageSize)
	if err := sess.Find(&forgejoComments); err != nil {
		panic(fmt.Errorf("error while listing comments: %v", err))
	}

	for _, forgejoComment := range forgejoComments {
		if err := forgejoComment.LoadPoster(ctx); err != nil {
			panic(fmt.Errorf("LoadPoster %+v %w", *forgejoComment, err))
		}
	}

	return f3_tree.ConvertListed(ctx, node, f3_tree.ConvertToAny(forgejoComments...)...)
}

func newComments() f3_tree_generic.NodeDriverInterface {
	return &comments{}
}
