// Package sources owns the `sources:` block written to mani.yaml by
// `mani init --source ...` and consumed by `mani refresh`.
//
// This package deliberately does not import core/dao: dao needs to write
// a `sources:` block on init, so depending on dao here would create a
// cycle. Refresh-time diff logic against []dao.Project lives in the dao
// package (core/dao/refresh.go).
package sources

import (
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/provider/gitea"
	"github.com/alajmo/mani/core/provider/github"
)

// Source is one entry in the YAML `sources:` list. The `Source` field
// holds the verbatim flag value (`github/my-org`, `codeberg/joe`, etc.)
// so the YAML round-trips back to what the user originally typed.
type Source struct {
	Source string `yaml:"source"`

	// Path, when set, is prepended to every project's path so all of a
	// source's repos land under a common directory (e.g. path "acme" maps
	// repo "alpha" to path "acme/alpha"). Empty means flat (path = name).
	Path string `yaml:"path,omitempty"`

	// Archived and Fork are "include?" toggles. Default (false) excludes
	// archived repos / forks; true includes them. Omitted in YAML when false
	// (the default), so a plain config carries neither key.
	Archived bool `yaml:"archived,omitempty"`
	Fork     bool `yaml:"fork,omitempty"`

	Limit       int  `yaml:"limit,omitempty"`
	NoTags      bool `yaml:"no_tags,omitempty"`
	NoTopicTags bool `yaml:"no_topic_tags,omitempty"`
}

// includeFilter maps a source-level "include this kind?" toggle onto the
// provider's tri-state filter (*bool): include (true) means no filter (nil);
// exclude (false) means keep only repos where the attribute is false.
func includeFilter(include bool) *bool {
	if include {
		return nil
	}
	exclude := false
	return &exclude
}

// Parse turns a `--source` flag value into a Source. The flag value is
// stored verbatim — `codeberg/joe` stays `codeberg/joe` in the YAML so
// the user's mental model survives a round-trip.
//
// Accepted providers:
//   - github/<name>
//   - gitea/<name>, forgejo/<name>, codeberg/<name>  (all route to Gitea)
func Parse(flag string) (Source, error) {
	if _, err := parseProviderName(flag); err != nil {
		return Source{}, err
	}
	return Source{Source: flag}, nil
}

// Provider returns the normalized provider name (always "github" or
// "gitea", with forgejo/codeberg collapsed to "gitea"). Empty if the
// source string is malformed.
func (s Source) Provider() string {
	pn, err := parseProviderName(s.Source)
	if err != nil {
		return ""
	}
	return pn.canonical
}

// Name returns just the org/user part of `provider/name`.
func (s Source) Name() string {
	pn, err := parseProviderName(s.Source)
	if err != nil {
		return ""
	}
	return pn.name
}

// Hint returns one of "" (auto), "org", or "user" based on whether the
// flag prefix explicitly disambiguated. `github/X` → "" (auto-detect);
// `github-org/X` → "org"; `github-user/X` → "user".
func (s Source) Hint() string {
	pn, err := parseProviderName(s.Source)
	if err != nil {
		return ""
	}
	return pn.hint
}

// Label returns the verbatim source string for display in errors and
// refresh summaries (e.g. "codeberg/joe", not "gitea/joe").
func (s Source) Label() string {
	return s.Source
}

func (s Source) ToGitHubListOptions() github.ListOptions {
	return github.ListOptions{
		Name:     s.Name(),
		Hint:     github.Hint(s.Hint()),
		Limit:    s.Limit,
		Archived: includeFilter(s.Archived),
		Fork:     includeFilter(s.Fork),
	}
}

func (s Source) ToGitHubMapOptions() github.MapOptions {
	return github.MapOptions{
		NoTags:      s.NoTags,
		NoTopicTags: s.NoTopicTags,
	}
}

func (s Source) ToGiteaListOptions() gitea.ListOptions {
	return gitea.ListOptions{
		Name:     s.Name(),
		Hint:     gitea.Hint(s.Hint()),
		Limit:    s.Limit,
		Archived: includeFilter(s.Archived),
		Fork:     includeFilter(s.Fork),
	}
}

func (s Source) ToGiteaMapOptions() gitea.MapOptions {
	return gitea.MapOptions{
		NoTags:      s.NoTags,
		NoTopicTags: s.NoTopicTags,
	}
}

type providerName struct {
	raw       string // "codeberg-org"
	canonical string // "github" | "gitea"
	hint      string // "" | "org" | "user"
	name      string // "joe"
}

// Recognized prefixes. The base name (`github`, `gitea`, …) means
// auto-detect; the `-org` / `-user` suffix forces a hint.
var providerPrefixes = map[string]struct{ canonical, hint string }{
	"github":        {"github", ""},
	"github-org":    {"github", "org"},
	"github-user":   {"github", "user"},
	"gitea":         {"gitea", ""},
	"gitea-org":     {"gitea", "org"},
	"gitea-user":    {"gitea", "user"},
	"forgejo":       {"gitea", ""},
	"forgejo-org":   {"gitea", "org"},
	"forgejo-user":  {"gitea", "user"},
	"codeberg":      {"gitea", ""},
	"codeberg-org":  {"gitea", "org"},
	"codeberg-user": {"gitea", "user"},
}

func parseProviderName(flag string) (providerName, error) {
	provider, name, ok := strings.Cut(flag, "/")
	if !ok || provider == "" || name == "" {
		return providerName{}, fmt.Errorf("invalid source %q: expected <provider>/<name>, e.g. github/my-org", flag)
	}
	p, ok := providerPrefixes[provider]
	if !ok {
		return providerName{}, fmt.Errorf("unsupported source provider %q (supported: github[-org|-user], gitea[-org|-user], forgejo[-org|-user], codeberg[-org|-user])", provider)
	}
	return providerName{raw: provider, canonical: p.canonical, hint: p.hint, name: name}, nil
}
