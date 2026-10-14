// Package goreecloudcontext creates bounded, provider-neutral read-only
// repository snapshots for an AI consumer. This Development library does not
// fetch forge content, authorize users, or expose an HTTP endpoint.
package goreecloudcontext

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const MaxItems = 20

type Repository struct {
	ID            string `json:"id"`
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"defaultBranch"`
	Private       bool   `json:"private"`
	WebURL        string `json:"webUrl"`
}
type Branch struct {
	Name      string `json:"name"`
	SHA       string `json:"sha"`
	Protected bool   `json:"protected"`
}
type Commit struct {
	SHA        string `json:"sha"`
	Message    string `json:"message"`
	AuthoredAt string `json:"authoredAt"`
	WebURL     string `json:"webUrl"`
}
type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	WebURL string `json:"webUrl"`
}
type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Base   string `json:"base"`
	Head   string `json:"head"`
	WebURL string `json:"webUrl"`
}
type Input struct {
	Repository   Repository
	Ref          string
	GeneratedAt  time.Time
	Provider     string
	Healthy      bool
	Capabilities []string
	Branches     []Branch
	Commits      []Commit
	Issues       []Issue
	PullRequests []PullRequest
}

func textOK(s string, limit int) bool {
	if len(s) == 0 || len(s) > limit || !utf8.ValidString(s) || strings.TrimSpace(s) != s {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validURL(raw string) bool {
	if !textOK(raw, 2048) {
		return false
	}
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") &&
		u.Hostname() != "" && u.User == nil && u.Fragment == "" &&
		!strings.ContainsAny(raw, string([]byte{13, 10, 92}))
}

// Only declared read scopes are valid evidence in the AI read-only envelope.
// Capability labels are metadata, not authentication or authorization grants.
func readCapabilityOK(scope string) bool {
	switch scope {
	case "repositories:read", "branches:read", "commits:read", "issues:read", "pullRequests:read":
		return true
	}
	return false
}

func descriptionOK(s string) bool {
	return len(s) <= 600 && utf8.ValidString(s) && !strings.ContainsFunc(s, unicode.IsControl)
}

func validCollectionLengths(in Input) bool {
	return len(in.Branches) <= MaxItems && len(in.Commits) <= MaxItems &&
		len(in.Issues) <= MaxItems && len(in.PullRequests) <= MaxItems
}

func validate(in Input) error {
	r := in.Repository
	if !textOK(r.ID, 128) || !textOK(r.Owner, 128) || !textOK(r.Name, 128) ||
		!textOK(r.DefaultBranch, 255) || !validURL(r.WebURL) ||
		!descriptionOK(r.Description) || !validCollectionLengths(in) ||
		in.GeneratedAt.IsZero() || in.GeneratedAt.After(time.Now()) ||
		time.Since(in.GeneratedAt) > 5*time.Minute {
		return errors.New("invalid_repository_context_metadata")
	}
	if in.Ref != "" && !textOK(in.Ref, 255) {
		return errors.New("invalid_repository_context_ref")
	}
	if in.Provider != "forgejo" || !in.Healthy {
		return errors.New("repository_context_provider_not_healthy")
	}
	hasRead := false
	if len(in.Capabilities) > 12 {
		return errors.New("invalid_repository_context_capabilities")
	}
	for _, c := range in.Capabilities {
		if !textOK(c, 64) || !readCapabilityOK(c) {
			return errors.New("invalid_repository_context_capabilities")
		}
		hasRead = hasRead || c == "repositories:read"
	}
	if !hasRead {
		return errors.New("repository_context_read_capability_required")
	}
	for _, branch := range in.Branches {
		if !textOK(branch.Name, 255) || !textOK(branch.SHA, 80) {
			return errors.New("invalid_repository_context_branch")
		}
	}
	for _, commit := range in.Commits {
		if !textOK(commit.SHA, 80) || !textOK(commit.Message, 600) ||
			!validURL(commit.WebURL) || !textOK(commit.AuthoredAt, 40) {
			return errors.New("invalid_repository_context_commit")
		}
		if _, err := time.Parse(time.RFC3339Nano, commit.AuthoredAt); err != nil {
			return errors.New("invalid_repository_context_commit_time")
		}
	}
	for _, issue := range in.Issues {
		if issue.Number < 1 || !textOK(issue.Title, 600) ||
			(issue.State != "open" && issue.State != "closed") || !validURL(issue.WebURL) {
			return errors.New("invalid_repository_context_issue")
		}
	}
	for _, pr := range in.PullRequests {
		if pr.Number < 1 || !textOK(pr.Title, 600) ||
			(pr.State != "open" && pr.State != "closed" && pr.State != "merged") ||
			!textOK(pr.Base, 255) || !textOK(pr.Head, 255) || !validURL(pr.WebURL) {
			return errors.New("invalid_repository_context_pull_request")
		}
	}
	return nil
}

// MarshalReadOnlySnapshot builds the GoreeCloud AI Development 0.1.0 envelope.
// The caller must authenticate the request and obtain real provider data and
// health evidence; neither action is performed or implied by this function.
func MarshalReadOnlySnapshot(in Input) ([]byte, error) {
	if err := validate(in); err != nil {
		return nil, err
	}
	type limits struct {
		Branches     int `json:"branches"`
		Commits      int `json:"commits"`
		Issues       int `json:"issues"`
		PullRequests int `json:"pullRequests"`
	}
	scope := map[string]any{
		"repository":   map[string]string{"owner": in.Repository.Owner, "name": in.Repository.Name},
		"repositoryId": in.Repository.ID,
	}
	if in.Ref != "" {
		scope["ref"] = in.Ref
	}
	payload := map[string]any{
		"contractVersion": "0.1.0",
		"recordType":      "repository_context",
		"producer":        map[string]any{"service": "goreecloud-code-api", "authoritative": false, "role": "provider-neutral-evidence-broker"},
		"scope":           scope,
		"generatedAt":     in.GeneratedAt.UTC().Format(time.RFC3339Nano),
		"source":          map[string]any{"provider": in.Provider, "providerHealthy": in.Healthy, "capabilities": in.Capabilities},
		"repository":      in.Repository,
		"branches":        append([]Branch{}, in.Branches...),
		"commits":         append([]Commit{}, in.Commits...),
		"issues":          append([]Issue{}, in.Issues...),
		"pullRequests":    append([]PullRequest{}, in.PullRequests...),
		"limits":          limits{MaxItems, MaxItems, MaxItems, MaxItems},
		"evidence":        map[string]any{"webUrls": []string{in.Repository.WebURL}},
	}
	return json.Marshal(payload)
}
