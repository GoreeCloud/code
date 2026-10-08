// Copyright Earl Warren <contact@earl-warren.org>
// Copyright Loïc Dachary <loic@dachary.org>
// SPDX-License-Identifier: MIT

package driver

import (
	"context"
	"fmt"

	"forgejo.org/models/db"
	org_model "forgejo.org/models/organization"
	user_model "forgejo.org/models/user"

	"code.forgejo.org/f3/gof3/v3/f3"
	f3_id "code.forgejo.org/f3/gof3/v3/id"
	f3_tree "code.forgejo.org/f3/gof3/v3/tree/f3"
	f3_tree_generic "code.forgejo.org/f3/gof3/v3/tree/generic"
)

type organizations struct {
	container
}

func (o *organizations) ListPage(ctx context.Context, node f3_tree_generic.NodeInterface, _ f3_tree_generic.ListOptions, page int) f3_tree_generic.ChildrenList {
	sess := db.GetEngine(ctx)
	if page != 0 {
		sess = db.SetSessionPagination(sess, &db.ListOptions{Page: page, PageSize: o.getPageSize()})
	}
	sess = sess.Select("`user`.*").
		Where("`type`=?", user_model.UserTypeOrganization)
	organizations := make([]*org_model.Organization, 0, o.getPageSize())

	if err := sess.Find(&organizations); err != nil {
		panic(fmt.Errorf("error while listing organizations: %v", err))
	}

	return f3_tree.ConvertListed(ctx, node, f3_tree.ConvertToAny(organizations...)...)
}

func (o *organizations) LookupMappedID(ctx context.Context, id f3_id.NodeID, f f3.Interface) f3_id.NodeID {
	organization := f.(*f3.Organization)
	return o.GetIDFromName(ctx, organization.Name)
}

func (o *organizations) GetIDFromName(ctx context.Context, name string) f3_id.NodeID {
	organization, err := org_model.GetOrgByName(ctx, name)
	if org_model.IsErrOrgNotExist(err) {
		return f3_id.NilID
	}
	if err != nil {
		panic(fmt.Errorf("GetOrgByName(%s): %v", name, err))
	}

	return f3_id.NewNodeID(organization.ID)
}

func newOrganizations() f3_tree_generic.NodeDriverInterface {
	return &organizations{}
}
