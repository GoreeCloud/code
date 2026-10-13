package util

import (
	"strings"
	"testing"
)

func FuzzGoreeCloudPathJoinRelX(f *testing.F) {
	for _, pair := range [][2]string{
		{"repository", `..\private`},
		{`team\docs`, `..\private`},
		{"repository", "../../etc/passwd"},
		{"repo", `folder\..\file`},
		{"", ".."},
	} {
		f.Add(pair[0], pair[1])
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		got := PathJoinRelX(a, b)
		if strings.Contains(got, `\`) || strings.HasPrefix(got, "/") ||
			got == ".." || strings.HasPrefix(got, "../") || strings.Contains(got, "/../") {
			t.Fatalf("non-relative path: %q", got)
		}
		normalized := PathJoinRelX(strings.ReplaceAll(a, `\`, "/"), strings.ReplaceAll(b, `\`, "/"))
		if got != normalized {
			t.Fatalf("separator equivalence mismatch: %q, %q", got, normalized)
		}
	})
}
