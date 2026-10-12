// Copyright 2024 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgefed

import (
	"context"
	"fmt"

	"forgejo.org/models/db"
	"forgejo.org/modules/log"
	"forgejo.org/modules/validation"
)

func init() {
	db.RegisterModel(new(FederationHost))
}

func CountFederationHosts(ctx context.Context) (int64, error) {
	return db.GetEngine(ctx).Count(FederationHost{})
}

func FindFederationHosts(ctx context.Context, opts db.ListOptions) (hosts []*FederationHost, err error) {
	sess := db.GetEngine(ctx)

	if opts.PageSize > 0 {
		sess = db.SetSessionPagination(sess, &opts)
	}

	err = sess.Find(&hosts)
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if res, err := validation.IsValid(host); !res {
			return nil, err
		}
	}

	return hosts, nil
}

func GetFederationHost(ctx context.Context, ID int64) (*FederationHost, error) {
	log.Trace("GetFederationHost: %v", ID)
	host := new(FederationHost)
	has, err := db.GetEngine(ctx).Where("id=?", ID).Get(host)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, fmt.Errorf("FederationInfo record %v does not exist", ID)
	}
	if res, err := validation.IsValid(host); !res {
		return nil, err
	}
	log.Trace("GetFederationHost: %v, got host %v", ID, host)
	return host, nil
}

func findFederationHostFromDB(ctx context.Context, searchKey string, searchValue ...any) (*FederationHost, error) {
	host := new(FederationHost)
	has, err := db.GetEngine(ctx).Where(searchKey, searchValue...).Get(host)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrFederationHostNotFound{SearchKey: searchKey, SearchValue: fmt.Sprintf("%v", searchValue)}
	}
	if res, err := validation.IsValid(host); !res {
		return nil, err
	}
	return host, nil
}

func FindFederationHostByFqdnAndPort(ctx context.Context, fqdn string, port uint16) (*FederationHost, error) {
	return findFederationHostFromDB(ctx, "host_fqdn=? AND host_port=?", fqdn, port)
}

func FindFederationHostByKeyID(ctx context.Context, keyID string) (*FederationHost, error) {
	return findFederationHostFromDB(ctx, "key_id=?", keyID)
}

func CreateFederationHost(ctx context.Context, host *FederationHost) error {
	if res, err := validation.IsValid(host); !res {
		return err
	}
	_, err := db.GetEngine(ctx).Insert(host)
	return err
}

func UpdateFederationHost(ctx context.Context, host *FederationHost) error {
	if res, err := validation.IsValid(host); !res {
		return err
	}
	_, err := db.GetEngine(ctx).ID(host.ID).Update(host)
	return err
}
