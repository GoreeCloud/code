// Copyright 2019 The Gitea Authors. All rights reserved.
// Copyright 2015 The Gogs Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package git

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"path"
	"regexp"
	"strings"

	"forgejo.org/modules/setting"
	"forgejo.org/modules/util"

	"gopkg.in/ini.v1" //nolint:depguard // used to read .gitmodules
)

const MaxGitmodulesFileSize = 64 * 1024

// GetSubmodule returns the Submodule of a given path
func (c *Commit) GetSubmodule(path string, entry *TreeEntry) (Submodule, error) {
	err := c.readSubmodules()
	if err != nil {
		// the .gitmodules file exists but could not be read or parsed
		return Submodule{}, err
	}

	sm, ok := c.submodules[path]
	if !ok {
		// no info found in .gitmodules: fallback to what we can provide
		return Submodule{
			Path:   path,
			Commit: entry.ID,
		}, nil
	}

	sm.Commit = entry.ID
	return sm, nil
}

// readSubmodules populates the submodules field by reading the .gitmodules file
func (c *Commit) readSubmodules() error {
	if c.submodules != nil {
		return nil
	}

	entry, err := c.GetTreeEntryByPath(".gitmodules")
	if err != nil {
		if IsErrNotExist(err) {
			c.submodules = make(map[string]Submodule)
			return nil
		}
		return err
	}

	rc, _, err := entry.Blob().NewReader(MaxGitmodulesFileSize)
	if err != nil {
		if errors.As(err, &BlobTooLargeError{}) {
			c.submodules = make(map[string]Submodule)
			return nil
		}
		return err
	}
	defer rc.Close()

	c.submodules, err = parseSubmoduleContent(rc)
	return err
}

func parseSubmoduleContent(r io.Reader) (map[string]Submodule, error) {
	// https://git-scm.com/docs/gitmodules#_description
	// The .gitmodules file, located in the top-level directory of a Git working tree
	// is a text file with a syntax matching the requirements of git-config[1].
	// https://git-scm.com/docs/git-config#_configuration_file

	cfg := ini.Empty(ini.LoadOptions{
		InsensitiveKeys: true, // "The variable names are case-insensitive", but "Subsection names are case sensitive"
	})
	err := cfg.Append(r)
	if err != nil {
		return nil, err
	}

	sections := cfg.Sections()
	submodule := make(map[string]Submodule, len(sections))

	for _, s := range sections {
		sm := parseSubmoduleSection(s)
		if sm.Path == "" || sm.URL == "" {
			continue
		}
		submodule[sm.Path] = sm
	}
	return submodule, nil
}

func parseSubmoduleSection(s *ini.Section) Submodule {
	section, name, _ := strings.Cut(s.Name(), " ")
	if !util.ASCIIEqualFold("submodule", section) { // See https://codeberg.org/forgejo/forgejo/pulls/8438#issuecomment-5805251
		return Submodule{}
	}
	_ = name

	sm := Submodule{}
	if key, _ := s.GetKey("path"); key != nil {
		sm.Path = key.Value()
	}
	if key, _ := s.GetKey("url"); key != nil {
		sm.URL = key.Value()
	}
	return sm
}

// Submodule represents a parsed git submodule reference.
type Submodule struct {
	Path   string   // path property
	URL    string   // upstream URL
	Commit ObjectID // upstream Commit-ID
}

// ResolveUpstreamURL resolves the upstream URL relative to the repo URL.
func (sm Submodule) ResolveUpstreamURL(repoURL string) string {
	repoFullName := strings.TrimPrefix(repoURL, setting.AppURL) // currently hacky, but can be dropped when refactoring getRefURL
	return getRefURL(sm.URL, setting.AppURL, repoFullName, setting.SSH.Domain)
}

var scpSyntax = regexp.MustCompile(`^([a-zA-Z0-9_]+@)?([a-zA-Z0-9._-]+):(.*)$`)

func getRefURL(refURL, urlPrefix, repoFullName, sshDomain string) string {
	if refURL == "" {
		return ""
	}

	refURI := strings.TrimSuffix(refURL, ".git")

	prefixURL, _ := url.Parse(urlPrefix)
	urlPrefixHostname, _, err := net.SplitHostPort(prefixURL.Host)
	if err != nil {
		urlPrefixHostname = prefixURL.Host
	}

	urlPrefix = strings.TrimSuffix(urlPrefix, "/")

	// FIXME: Need to consider branch - which will require changes in parseSubmoduleSection
	// Relative url prefix check (according to git submodule documentation)
	if strings.HasPrefix(refURI, "./") || strings.HasPrefix(refURI, "../") {
		return urlPrefix + path.Clean(path.Join("/", repoFullName, refURI))
	}

	if !strings.Contains(refURI, "://") {
		// scp style syntax which contains *no* port number after the : (and is not parsed by net/url)
		// ex: git@try.gitea.io:go-gitea/gitea
		match := scpSyntax.FindAllStringSubmatch(refURI, -1)
		if len(match) > 0 {
			m := match[0]
			refHostname := m[2]
			pth := m[3]

			if !strings.HasPrefix(pth, "/") {
				pth = "/" + pth
			}

			if urlPrefixHostname == refHostname || refHostname == sshDomain {
				return urlPrefix + path.Clean(path.Join("/", pth))
			}
			return "http://" + refHostname + pth
		}
	}

	ref, err := url.Parse(refURI)
	if err != nil {
		return ""
	}

	refHostname, _, err := net.SplitHostPort(ref.Host)
	if err != nil {
		refHostname = ref.Host
	}

	supportedSchemes := []string{"http", "https", "git", "ssh", "git+ssh"}

	for _, scheme := range supportedSchemes {
		if ref.Scheme == scheme {
			if ref.Scheme == "http" || ref.Scheme == "https" {
				if len(ref.User.Username()) > 0 {
					return ref.Scheme + "://" + fmt.Sprintf("%v", ref.User) + "@" + ref.Host + ref.Path
				}
				return ref.Scheme + "://" + ref.Host + ref.Path
			} else if urlPrefixHostname == refHostname || refHostname == sshDomain {
				return urlPrefix + path.Clean(path.Join("/", ref.Path))
			}
			return "http://" + refHostname + ref.Path
		}
	}

	return ""
}
