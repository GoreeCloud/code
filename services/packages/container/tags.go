// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package container

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	packages_model "forgejo.org/models/packages"
	container_model "forgejo.org/models/packages/container"
)

type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func GetLocalTagList(ctx context.Context, ownerLower, image, last string, n int, ownerID int64) (*TagList, *url.Values, error) {
	_, err := packages_model.GetPackageByName(ctx, ownerID, packages_model.TypeContainer, image)
	if err != nil {
		return nil, nil, err
	}
	tags, err := container_model.GetImageTags(ctx, ownerID, image, n, last)
	if err != nil {
		return nil, nil, err
	}
	tagList := &TagList{
		Name: strings.ToLower(ownerLower + "/" + image),
		Tags: tags,
	}
	v := setLinkHeaderValues(tagList, n)
	return tagList, v, nil
}

func setLinkHeaderValues(tagList *TagList, n int) *url.Values {
	v := &url.Values{}
	if len(tagList.Tags) > 0 {
		if n > 0 {
			v.Add("n", strconv.Itoa(n))
		}
		v.Add("last", tagList.Tags[len(tagList.Tags)-1])
	}
	return v
}
