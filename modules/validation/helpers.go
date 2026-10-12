// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package validation

import (
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"forgejo.org/modules/setting"
)

var externalTrackerRegex = regexp.MustCompile(`({?)(?:user|repo|index)+?(}?)`)

func isLoopbackIP(ip string) bool {
	return net.ParseIP(ip).IsLoopback()
}

// IsValidURL checks if URL is valid
func IsValidURL(uri string) bool {
	if u, err := url.ParseRequestURI(uri); err != nil ||
		(u.Scheme != "http" && u.Scheme != "https") ||
		!validPort(portOnly(u.Host)) {
		return false
	}

	return true
}

// IsValidSiteURL checks if URL is valid
func IsValidSiteURL(uri string) bool {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return false
	}

	if !validPort(portOnly(u.Host)) {
		return false
	}

	return slices.Contains(setting.Service.ValidSiteURLSchemes, u.Scheme)
}

// IsAPIURL checks if URL is current Gitea instance API URL
func IsAPIURL(uri string) bool {
	return strings.HasPrefix(strings.ToLower(uri), strings.ToLower(setting.AppURL+"api"))
}

// IsValidExternalURL checks if URL is valid external URL
func IsValidExternalURL(uri string) bool {
	if !IsValidURL(uri) || IsAPIURL(uri) {
		return false
	}

	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return false
	}

	// Currently check only if not loopback IP is provided to keep compatibility
	if isLoopbackIP(u.Hostname()) || strings.ToLower(u.Hostname()) == "localhost" {
		return false
	}

	// TODO: Later it should be added to allow local network IP addresses
	//       only if allowed by special setting

	return true
}

// IsValidReleaseAssetURL checks if the URL is valid for external release assets
func IsValidReleaseAssetURL(uri string) bool {
	return IsValidURL(uri)
}

// IsValidExternalTrackerURLFormat checks if URL matches required syntax for external trackers
func IsValidExternalTrackerURLFormat(uri string) bool {
	if !IsValidExternalURL(uri) {
		return false
	}

	// check for typoed variables like /{index/ or /[repo}
	for _, match := range externalTrackerRegex.FindAllStringSubmatch(uri, -1) {
		if (match[1] == "{" || match[2] == "}") && (match[1] != "{" || match[2] != "}") {
			return false
		}
	}

	return true
}

var (
	validUsernamePatternWithDots    = regexp.MustCompile(`^[\da-zA-Z][-.\w]*$`)
	validUsernamePatternWithoutDots = regexp.MustCompile(`^[\da-zA-Z][-\w]*$`)

	// No consecutive or trailing non-alphanumeric chars, catches both cases
	invalidUsernamePattern = regexp.MustCompile(`[-._]{2,}|[-._]$`)

	// This is intended to accept any character, in any language, with accent symbols,
	// as well as an arbitrary amount of subdomains and an optional port number defined
	// through `:12345`.
	//
	// This is intended to cover username cases from distant servers in the fediverse, which
	// can have much laxer requirements than those of Forgejo. It is not intended to check for
	// invalid, non-standard compliant domains.
	//
	// For instance, the following should work:
	// @user.όνομαß_21__@subdomain1.subdomain2.example.tld:65536
	// @42@42.example.tld
	// @user@example.tld:99999 (presumed to be an impossible case)
	// @-@-.tld (also impossible)
	validFediverseUsernamePattern = regexp.MustCompile(`^(@[\p{L}\p{M}0-9_\.\-]{1,})(@[\p{L}\p{M}0-9_\.\-]{1,})(:[1-9][0-9]{0,4})?$`)
)

// IsValidUsername checks if username is valid
func IsValidUsername(name string) bool {
	// It is difficult to find a single pattern that is both readable and effective,
	// but it's easier to use positive and negative checks.
	if setting.Service.AllowDotsInUsernames {
		return validUsernamePatternWithDots.MatchString(name) && !invalidUsernamePattern.MatchString(name)
	}

	return validUsernamePatternWithoutDots.MatchString(name) && !invalidUsernamePattern.MatchString(name)
}

// IsValidActivityPubUsername checks whether the username can be a valid ActivityPub handle.
//
// Username refers to the Forgejo user account's username for consistency, and not
// e.g. "username" in @username@example.tld.
func IsValidActivityPubUsername(name string) bool {
	return validFediverseUsernamePattern.MatchString(name)
}
