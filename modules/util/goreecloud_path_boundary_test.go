// SPDX-License-Identifier: MIT
package util

import "testing"

// Regression guard for mixed Windows/POSIX separators in portable repository
// path building. Every untrusted component must be independently canonicalized.
func TestGoreeCloudPathJoinRelXTraversalBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{"windows parent", []string{"repository", `..\private`}, "repository/private"},
		{"mixed separators", []string{`team\docs`, `..\private`}, "team/docs/private"},
		{"embedded parent", []string{"repository", "folder/../../secret"}, "repository/secret"},
		{"normal file", []string{"repository", `folder\file.txt`}, "repository/folder/file.txt"},
		{"leading slash", []string{"repository", "/leading"}, "repository/leading"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := PathJoinRelX(tc.parts...); got != tc.want {
				t.Fatalf("PathJoinRelX(%q) = %q, want %q", tc.parts, got, tc.want)
			}
		})
	}
}
