// Copyright 2018 The Gitea Authors. All rights reserved.
// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package validation

import (
	"testing"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/test"

	"github.com/stretchr/testify/assert"
)

func Test_IsValidURL(t *testing.T) {
	cases := []struct {
		description string
		url         string
		valid       bool
	}{
		{
			description: "Empty URL",
			url:         "",
			valid:       false,
		},
		{
			description: "Loopback IPv4 URL",
			url:         "http://127.0.1.1:5678/",
			valid:       true,
		},
		{
			description: "Loopback IPv6 URL",
			url:         "https://[::1]/",
			valid:       true,
		},
		{
			description: "Missing semicolon after schema",
			url:         "http//meh/",
			valid:       false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, testCase.valid, IsValidURL(testCase.url))
		})
	}
}

func Test_IsValidExternalURL(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, "https://code.forgejo.org/")()

	cases := []struct {
		description string
		url         string
		valid       bool
	}{
		{
			description: "Current instance URL",
			url:         "https://code.forgejo.org/test",
			valid:       true,
		},
		{
			description: "Loopback IPv4 URL",
			url:         "http://127.0.1.1:5678/",
			valid:       false,
		},
		{
			description: "Current instance API URL",
			url:         "https://code.forgejo.org/api/v1/user/follow",
			valid:       false,
		},
		{
			description: "Local network URL",
			url:         "http://192.168.1.2/api/v1/user/follow",
			valid:       true,
		},
		{
			description: "Local URL",
			url:         "http://LOCALHOST:1234/whatever",
			valid:       false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, testCase.valid, IsValidExternalURL(testCase.url))
		})
	}
}

func Test_IsValidExternalTrackerURLFormat(t *testing.T) {
	defer test.MockVariableValue(&setting.AppURL, "https://code.forgejo.org/")()

	cases := []struct {
		description string
		url         string
		valid       bool
	}{
		{
			description: "Correct external tracker URL with all placeholders",
			url:         "https://github.com/{user}/{repo}/issues/{index}",
			valid:       true,
		},
		{
			description: "Local external tracker URL with all placeholders",
			url:         "https://127.0.0.1/{user}/{repo}/issues/{index}",
			valid:       false,
		},
		{
			description: "External tracker URL with typo placeholder",
			url:         "https://github.com/{user}/{repo/issues/{index}",
			valid:       false,
		},
		{
			description: "External tracker URL with typo placeholder",
			url:         "https://github.com/[user}/{repo/issues/{index}",
			valid:       false,
		},
		{
			description: "External tracker URL with typo placeholder",
			url:         "https://github.com/{user}/repo}/issues/{index}",
			valid:       false,
		},
		{
			description: "External tracker URL missing optional placeholder",
			url:         "https://github.com/{user}/issues/{index}",
			valid:       true,
		},
		{
			description: "External tracker URL missing optional placeholder",
			url:         "https://github.com/{repo}/issues/{index}",
			valid:       true,
		},
		{
			description: "External tracker URL missing optional placeholder",
			url:         "https://github.com/issues/{index}",
			valid:       true,
		},
		{
			description: "External tracker URL missing optional placeholder",
			url:         "https://github.com/issues/{user}",
			valid:       true,
		},
		{
			description: "External tracker URL with similar placeholder names test",
			url:         "https://github.com/user/repo/issues/{index}",
			valid:       true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, testCase.valid, IsValidExternalTrackerURLFormat(testCase.url))
		})
	}
}

func TestIsValidUsernameAllowDots(t *testing.T) {
	defer test.MockVariableValue(&setting.Service.AllowDotsInUsernames, true)()

	tests := []struct {
		arg  string
		want bool
	}{
		{arg: "a", want: true},
		{arg: "abc", want: true},
		{arg: "0.b-c", want: true},
		{arg: "a.b-c_d", want: true},
		{arg: "", want: false},
		{arg: ".abc", want: false},
		{arg: "abc.", want: false},
		{arg: "a..bc", want: false},
		{arg: "a...bc", want: false},
		{arg: "a.-bc", want: false},
		{arg: "a._bc", want: false},
		{arg: "a_-bc", want: false},
		{arg: "a/bc", want: false},
		{arg: "☁️", want: false},
		{arg: "-", want: false},
		{arg: "--diff", want: false},
		{arg: "-im-here", want: false},
		{arg: "a space", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsValidUsername(tt.arg), "IsValidUsername(%v)", tt.arg)
		})
	}
}

func TestIsValidUsernameBanDots(t *testing.T) {
	defer test.MockVariableValue(&setting.Service.AllowDotsInUsernames, false)()

	tests := []struct {
		arg  string
		want bool
	}{
		{arg: "a", want: true},
		{arg: "abc", want: true},
		{arg: "0.b-c", want: false},
		{arg: "a.b-c_d", want: false},
		{arg: ".abc", want: false},
		{arg: "abc.", want: false},
		{arg: "a..bc", want: false},
		{arg: "a...bc", want: false},
		{arg: "a.-bc", want: false},
		{arg: "a._bc", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsValidUsername(tt.arg), "IsValidUsername[AllowDotsInUsernames=false](%v)", tt.arg)
		})
	}
}

func TestIsValidActivityPubUsername(t *testing.T) {
	cases := []struct {
		description string
		username    string
		valid       bool
	}{
		{
			description: "Username without domain",
			username:    "@user",
			valid:       false,
		},
		{
			description: "Username with domain",
			username:    "@user@example.tld",
			valid:       true,
		},
		{
			description: "Numeric username with subdomain",
			username:    "@42@42.example.tld",
			valid:       true,
		},
		{
			description: "Username with two subdomains",
			username:    "@user@forgejo.activitypub.example.tld",
			valid:       true,
		},
		{
			description: "Username with domain and without port",
			username:    "@user@social.example.tld:",
			valid:       false,
		},
		{
			description: "Username with domain and invalid port 0",
			username:    "@user@social.example.tld:0",
			valid:       false,
		},
		{
			// We do not validate the port and assume that federationHost.HostPort
			// cannot present such invalid ports. That also makes the previous case
			// (port: 0) redundant, but it doesn't hurt.
			description: "Username with domain and valid port",
			username:    "@user@social.example.tld:65536",
			valid:       true,
		},
		{
			description: "Username with Latin letters and special symbols",
			username:    "@$username$@example.tld",
			valid:       false,
		},
		{
			description: "Strictly numeric handle, domain, TLD",
			username:    "@0123456789@0123456789.0123456789.123",
			valid:       true,
		},
		{
			description: "Handle with Latin characters and dashes",
			username:    "@0-O@O-O.tld",
			valid:       true,
		},
		// This is an impossible case, but we assume that this will never happen
		// to begin with.
		{
			description: "Handle that only has dashes",
			username:    "@-@-.-",
			valid:       true,
		},
		{
			description: "Username with a mix of Latin and non-Latin letters containing accents",
			username:    "@usernäme.όνομαß_21__@example.tld",
			valid:       true,
		},
		// Note: Our regex should accept any character, in any language and with accent symbols.
		// The list is neither exhaustive, nor does it represent all possible cases.
		// I chose some TLDs from https://en.wikipedia.org/wiki/Country_code_top-level_domain,
		// although only one test case should suffice in theory. Nevertheless, to play it safe,
		// I included four from different geographic regions whose scripts were legible using my
		// IDE's default font to play it safe.
		{
			description: "Username, domain and ccTLD in Greek",
			username:    "@ευ@ευ.ευ",
			valid:       true,
		},
		{
			description: "Username, domain and ccTLD in Georgian (Mkhedruli)",
			username:    "@გე@გე.გე",
			valid:       true,
		},
		{
			description: "Username, domain and ccTLD of Malaysia (Arabic Jawi)",
			username:    "@مليسيا@ລمليسيا.مليسيا",
			valid:       true,
		},
		{
			description: "Username, domain and ccTLD of China (Simplified)",
			username:    "@中国@中国.中国",
			valid:       true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, testCase.valid, IsValidActivityPubUsername(testCase.username))
		})
	}
}
