// Copyright Earl Warren <contact@earl-warren.org>
// Copyright Loïc Dachary <loic@dachary.org>
// SPDX-License-Identifier: MIT

package driver

import (
	"context"

	"code.forgejo.org/f3/gof3/v3/f3"
	f3_tree "code.forgejo.org/f3/gof3/v3/tree/f3"
	f3_tree_generic "code.forgejo.org/f3/gof3/v3/tree/generic"
)

type repositories struct {
	container
}

func (o *repositories) ListPage(ctx context.Context, node f3_tree_generic.NodeInterface, _ f3_tree_generic.ListOptions, page int) f3_tree_generic.ChildrenList {
	children := f3_tree_generic.NewChildrenList(0)
	if page > 1 {
		return children
	}

	names := []string{f3.RepositoryNameDefault}
	project := f3_tree.GetProject(node).ToFormat().(*f3.Project)
	if project.HasWiki {
		names = append(names, RepositoryNameWiki)
	}

	return f3_tree.ConvertListed(ctx, node, f3_tree.ConvertToAny(names...)...)
}

func newRepositories() f3_tree_generic.NodeDriverInterface {
	return &repositories{}
}
