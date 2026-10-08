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

type labels struct {
	container
}

func (o *labels) ListPage(ctx context.Context, node f3_tree_generic.NodeInterface, _ f3_tree_generic.ListOptions, page int) f3_tree_generic.ChildrenList {
	pageSize := o.getPageSize()

	project := f3_tree.GetProjectID(node)

	forgejoLabels, err := issues_model.GetLabelsByRepoID(ctx, project, "", db.ListOptions{Page: page, PageSize: pageSize})
	if err != nil {
		panic(fmt.Errorf("error while listing labels: %v", err))
	}

	return f3_tree.ConvertListed(ctx, node, f3_tree.ConvertToAny(forgejoLabels...)...)
}

func newLabels() f3_tree_generic.NodeDriverInterface {
	return &labels{}
}
