package goreecloudcontext

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func exampleInput() Input {
	now := time.Now().UTC().Add(-time.Minute)
	return Input{
		Repository: Repository{
			ID: "42", Owner: "GoreeCloud", Name: "code",
			Description: "Development forge", DefaultBranch: "main",
			WebURL: "https://git.example.test/GoreeCloud/code",
		},
		Ref: "main", GeneratedAt: now, Provider: "forgejo", Healthy: true,
		Capabilities: []string{"repositories:read", "issues:read", "pullRequests:read"},
		Branches: []Branch{{Name: "main", SHA: "abc", Protected: true}},
		Commits: []Commit{{
			SHA: "abc", Message: "Fix session",
			AuthoredAt: now.Format(time.RFC3339), WebURL: "https://git.example.test/commit/abc",
		}},
		Issues: []Issue{{Number: 2, Title: "Fix session", State: "open", WebURL: "https://git.example.test/issues/2"}},
		PullRequests: []PullRequest{{
			Number: 3, Title: "Better sessions", State: "open",
			Base: "main", Head: "fix/session", WebURL: "https://git.example.test/pulls/3",
		}},
	}
}

func TestSnapshotMatchesReadOnlyContract(t *testing.T) {
	out, err := MarshalReadOnlySnapshot(exampleInput())
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatal(err)
	}
	if result["contractVersion"] != "0.1.0" || result["recordType"] != "repository_context" {
		t.Fatalf("unexpected envelope: %s", out)
	}
	producer := result["producer"].(map[string]any)
	if producer["authoritative"] != false || producer["service"] != "goreecloud-code-api" {
		t.Fatalf("producer must remain nonauthoritative: %s", out)
	}
	scope := result["scope"].(map[string]any)
	if scope["repositoryId"] != "42" || scope["ref"] != "main" {
		t.Fatalf("unexpected scope: %s", out)
	}
	limits := result["limits"].(map[string]any)
	if limits["issues"] != float64(MaxItems) {
		t.Fatalf("unexpected limits: %s", out)
	}
	if strings.Contains(string(out), "authorization") || strings.Contains(string(out), "cloneUrl") {
		t.Fatalf("secret or provider-specific field leaked: %s", out)
	}
}

func TestSnapshotRejectsUnsafeInput(t *testing.T) {
	for name, change := range map[string]func(*Input){
		"missing identity": func(x *Input) { x.Repository.ID = "" },
		"embedded credentials": func(x *Input) { x.Repository.WebURL = "https://user:pass@example.test/repo" },
		"no read permission": func(x *Input) { x.Capabilities = []string{"issues:read"} },
		"source unhealthy": func(x *Input) { x.Healthy = false },
		"missing provider": func(x *Input) { x.Provider = "" },
		"invalid commit URL": func(x *Input) { x.Commits[0].WebURL = "file:///etc/passwd" },
		"invalid timestamp": func(x *Input) { x.GeneratedAt = time.Time{} },
		"future timestamp": func(x *Input) { x.GeneratedAt = time.Now().Add(time.Second) },
		"expired snapshot": func(x *Input) { x.GeneratedAt = time.Now().Add(-6 * time.Minute) },
		"invalid issue": func(x *Input) { x.Issues[0].Number = -1 },
		"invalid branch": func(x *Input) { x.Branches[0].Name = "bad\nname" },
		"too many items": func(x *Input) { x.Issues = make([]Issue, MaxItems+1) },
	} {
		t.Run(name, func(t *testing.T) {
			input := exampleInput()
			change(&input)
			b, err := MarshalReadOnlySnapshot(input)
			if err == nil || b != nil {
				t.Fatalf("unsafe input was accepted: %v, %s", err, b)
			}
		})
	}
}
