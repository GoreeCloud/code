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

type issues struct {
	container
}

func (o *issues) ListPage(ctx context.Context, node f3_tree_generic.NodeInterface, _ f3_tree_generic.ListOptions, page int) f3_tree_generic.ChildrenList {
	pageSize := o.getPageSize()

	project := f3_tree.GetProjectID(node)

	forgejoIssues, err := issues_model.Issues(ctx, &issues_model.IssuesOptions{
		Paginator: &db.ListOptions{Page: page, PageSize: pageSize},
		RepoIDs:   []int64{project},
	})
	if err != nil {
		panic(fmt.Errorf("error while listing issues: %v", err))
	}

	return f3_tree.ConvertListed(ctx, node, f3_tree.ConvertToAny(forgejoIssues...)...)
}

func newIssues() f3_tree_generic.NodeDriverInterface {
	return &issues{}
}
