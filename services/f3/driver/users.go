// Copyright Earl Warren <contact@earl-warren.org>
// Copyright Loïc Dachary <loic@dachary.org>
// SPDX-License-Identifier: MIT

package driver

import (
	"context"
	"fmt"

	"forgejo.org/models/db"
	user_model "forgejo.org/models/user"

	"code.forgejo.org/f3/gof3/v3/f3"
	f3_id "code.forgejo.org/f3/gof3/v3/id"
	f3_tree "code.forgejo.org/f3/gof3/v3/tree/f3"
	f3_tree_generic "code.forgejo.org/f3/gof3/v3/tree/generic"
)

type users struct {
	container
}

func (o *users) ListPage(ctx context.Context, node f3_tree_generic.NodeInterface, _ f3_tree_generic.ListOptions, page int) f3_tree_generic.ChildrenList {
	sess := db.GetEngine(ctx).In("type", user_model.UserTypeIndividual, user_model.UserTypeRemoteUser)
	if page != 0 {
		sess = db.SetSessionPagination(sess, &db.ListOptions{Page: page, PageSize: o.getPageSize()})
	}
	sess = sess.Select("`user`.*")
	users := make([]*user_model.User, 0, o.getPageSize())

	if err := sess.Find(&users); err != nil {
		panic(fmt.Errorf("error while listing users: %v", err))
	}

	return f3_tree.ConvertListed(ctx, node, f3_tree.ConvertToAny(users...)...)
}

func (o *users) LookupMappedID(ctx context.Context, id f3_id.NodeID, f f3.Interface) f3_id.NodeID {
	user := f.(*f3.User)
	return o.GetIDFromName(ctx, user.UserName)
}

func (o *users) GetIDFromName(ctx context.Context, name string) f3_id.NodeID {
	user, err := user_model.GetUserByName(ctx, name)
	if user_model.IsErrUserNotExist(err) {
		return f3_id.NilID
	}
	if err != nil {
		panic(fmt.Errorf("GetUserByName(%s): %v", name, err))
	}

	return f3_id.NewNodeID(user.ID)
}

func newUsers() f3_tree_generic.NodeDriverInterface {
	return &users{}
}
