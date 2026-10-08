// Copyright Earl Warren <contact@earl-warren.org>
// Copyright Loïc Dachary <loic@dachary.org>
// SPDX-License-Identifier: MIT

package driver

import (
	"context"

	f3_kind "code.forgejo.org/f3/gof3/v3/kind"
	f3_tree_generic "code.forgejo.org/f3/gof3/v3/tree/generic"
)

type common struct {
	f3_tree_generic.NullDriver
}

func (o *common) GetHelper() any {
	panic("not implemented")
}

func (o *common) ListPage(ctx context.Context, node f3_tree_generic.NodeInterface, _ f3_tree_generic.ListOptions, page int) f3_tree_generic.ChildrenList {
	return f3_tree_generic.NewChildrenList(0)
}

func (o *common) GetNativeID() string {
	return ""
}

func (o *common) SetNative(native any) {
}

func (o *common) getTree() f3_tree_generic.TreeInterface {
	return o.GetNode().GetTree()
}

func (o *common) getPageSize() int {
	return o.getTreeDriver().GetPageSize()
}

func (o *common) getKind() f3_kind.Kind {
	return o.GetNode().GetKind()
}

func (o *common) getTreeDriver() *treeDriver {
	return o.GetTreeDriver().(*treeDriver)
}

func (o *common) IsNull() bool { return false }
