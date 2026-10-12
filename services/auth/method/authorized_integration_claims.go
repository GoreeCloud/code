// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package method

import (
	"fmt"
	"maps"

	"forgejo.org/modules/json"

	"github.com/golang-jwt/jwt/v5"
)

// Structure for inspecting the standard claims of a JWT (jwt.RegisteredClaims), which also stores any provided
// service-defined claims in an unstructured map[string]any.
type flexibleClaims struct {
	jwt.RegisteredClaims
	other map[string]any
}

// Populate a [flexibleClaims] from JSON data, implementing [json.Unmarshaler].
func (a *flexibleClaims) UnmarshalJSON(b []byte) error {
	var s map[string]any
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	var rc jwt.RegisteredClaims
	other := map[string]any{}
	for k, v := range s {
		switch k {
		case "iss":
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("expected `iss` to be string, but was %v", v)
			}
			rc.Issuer = str
		case "sub":
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("expected `sub` to be string, but was %v", v)
			}
			rc.Subject = str
		case "aud":
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("uanble to return `aud` to []byte: %w", err)
			}
			if err := json.Unmarshal(b, &rc.Audience); err != nil {
				return fmt.Errorf("uanble to decode `aud: %w", err)
			}
		case "exp":
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("uanble to return `exp` to []byte: %w", err)
			}
			if err := json.Unmarshal(b, &rc.ExpiresAt); err != nil {
				return fmt.Errorf("uanble to decode `exp: %w", err)
			}
		case "nbf":
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("uanble to return `nbf` to []byte: %w", err)
			}
			if err := json.Unmarshal(b, &rc.NotBefore); err != nil {
				return fmt.Errorf("uanble to decode `nbf: %w", err)
			}
		case "iat":
			b, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("uanble to return `iat` to []byte: %w", err)
			}
			if err := json.Unmarshal(b, &rc.IssuedAt); err != nil {
				return fmt.Errorf("uanble to decode `iat: %w", err)
			}
		case "jti":
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("expected `jti` to be string, but was %v", v)
			}
			rc.ID = str
		default:
			other[k] = v
		}
	}

	a.RegisteredClaims = rc
	a.other = other

	return nil
}

// Marshal flexibleClaims to JSON, merging both the registered claims and the additional claims into a map.
func (a flexibleClaims) MarshalJSON() ([]byte, error) {
	rcJSON, err := json.Marshal(a.RegisteredClaims)
	if err != nil {
		return nil, err
	}

	var fullMap map[string]any
	err = json.Unmarshal(rcJSON, &fullMap)
	if err != nil {
		return nil, err
	}
	maps.Copy(fullMap, a.other)

	return json.Marshal(fullMap)
}
