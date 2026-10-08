// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package auth

import (
	"forgejo.org/modules/container"
	"forgejo.org/modules/json"
	"forgejo.org/modules/log"
)

func UnmarshalGroupTeamMapping(raw string) (map[string]map[string][]string, error) {
	groupTeamMapping := make(map[string]map[string][]string)
	if raw == "" {
		return groupTeamMapping, nil
	}
	err := json.Unmarshal([]byte(raw), &groupTeamMapping)
	if err != nil {
		log.Error("Failed to unmarshal group team mapping: %v", err)
		return nil, err
	}
	return groupTeamMapping, nil
}

func UnmarshalQuotaGroupMapping(raw string) (map[string]container.Set[string], error) {
	quotaGroupMapping := make(map[string]container.Set[string])
	if raw == "" {
		return quotaGroupMapping, nil
	}

	rawMapping := make(map[string][]string)
	err := json.Unmarshal([]byte(raw), &rawMapping)
	if err != nil {
		log.Error("Failed to unmarshal group quota group mapping: %v", err)
		return nil, err
	}

	for key, values := range rawMapping {
		set := make(container.Set[string])
		set.AddMultiple(values...)
		quotaGroupMapping[key] = set
	}

	return quotaGroupMapping, nil
}
