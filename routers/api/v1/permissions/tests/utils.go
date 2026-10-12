// Copyright 2026 The Forgejo Authors.
// SPDX-License-Identifier: GPLv3-or-later

package tests

import (
	"fmt"
	"strings"

	auth_model "forgejo.org/models/auth"
	unit_model "forgejo.org/models/unit"
)

func levelStringToLevel(levelString string) auth_model.AccessTokenScopeLevel {
	level := auth_model.Read
	if levelString != "" {
		switch levelString {
		case "read":
			level = auth_model.Read
		case "write":
			level = auth_model.Write
		case "noaccess":
			level = auth_model.NoAccess
		default:
			panic(fmt.Sprintf("unexpected level '%s'", levelString))
		}
	}
	return level
}

func unitsTypeToString(unitTypes ...unit_model.Type) string {
	var unitStrings []string
	for _, unitType := range unitTypes {
		var unit *unit_model.Unit
		for _, u := range unit_model.Units {
			if u.Type == unitType {
				unit = &u
				break
			}
		}
		if unit == nil {
			panic(fmt.Errorf("unable to find a unit with type %v", unitType))
		}
		unitStrings = append(unitStrings, unit.NameKey)
	}
	return strings.Join(unitStrings, ",")
}

func unitsToScopes(unitTypes []unit_model.Type, levelString string) string {
	var scopeStrings []string
	for _, unitType := range unitTypes {
		unit := strings.TrimPrefix(unitsTypeToString(unitType), "repo.")
		var scope string
		switch unit {
		case "issues":
			scope = "issue"
		case "code", "pulls", "wiki", "project", "actions", "releases":
			scope = "repository"
		case "packages":
			scope = "package"
		default:
			panic(fmt.Errorf("unexpected unit type %v", unitType))
		}
		scopeStrings = append(scopeStrings, fmt.Sprintf("%s:%s", levelString, scope))
	}
	return strings.Join(scopeStrings, ",")
}
