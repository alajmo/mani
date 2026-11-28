// Package gitea fetches repository metadata from a Gitea instance by
// shelling out to the `tea` CLI. Mirrors core/provider/github in shape;
// no HTTP client or auth handling on our side — `tea` already knows
// which Gitea instance the user is logged into.
//
// Works against any Gitea-compatible host that `tea` supports: Gitea,
// Forgejo, Codeberg, etc.
package gitea

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Repo mirrors the subset of Gitea's API repo object we care about.
// Gitea's REST is GitHub-shaped; field names match deliberately.
type Repo struct {
	Name          string   `json:"name"`
	FullName      string   `json:"full_name"`
	SSHURL        string   `json:"ssh_url"`
	CloneURL      string   `json:"clone_url"`
	Description   string   `json:"description"`
	DefaultBranch string   `json:"default_branch"`
	Private       bool     `json:"private"`
	Internal      bool     `json:"internal"` // Gitea has explicit `internal` visibility
	Archived      bool     `json:"archived"`
	Fork          bool     `json:"fork"`
	Topics        []string `json:"topics"`
}

// Visibility returns "private", "public", or "internal" derived from
// Gitea's flags. Gitea has no single `visibility` field like GitHub
// (added in newer versions, but not universally); we derive it.
func (r Repo) Visibility() string {
	switch {
	case r.Internal:
		return "internal"
	case r.Private:
		return "private"
	default:
		return "public"
	}
}

type Hint string

const (
	HintAuto Hint = ""
	HintOrg  Hint = "org"
	HintUser Hint = "user"
)

type ListOptions struct {
	Name     string
	Hint     Hint
	Limit    int
	Archived *bool
	Fork     *bool
}

type MapOptions struct {
	NoTags      bool
	NoTopicTags bool
}

type Runner func(ctx context.Context, args ...string) ([]byte, error)

func DefaultRunner(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "tea", args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, &cliError{args: args, output: strings.TrimSpace(string(exitErr.Stderr))}
		}
		return nil, fmt.Errorf("tea %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

func CheckTeaAvailable() error {
	if _, err := exec.LookPath("tea"); err != nil {
		return fmt.Errorf("`tea` CLI is required for Gitea sources but was not found on PATH.\nInstall it from https://gitea.com/gitea/tea and run `tea login add` for your Gitea instance")
	}
	return nil
}

// pageSize is per-page when we paginate manually. Gitea's API caps this
// at 50 by default (configurable per instance, but 50 is the safe ceiling).
const pageSize = 50

// ListAll fetches repos for opts.Name. With Hint=auto (default), tries
// /orgs/<name>/repos first and falls back on 404. Explicit hints skip
// the probe.
//
// It also reports whether the listing is partial: whether
// it may leave out repos that exist. That is the case for a user that isn't
// the authenticated one, whose private repos are not listed. A repo that is
// missing from a partial listing can't be taken for gone.
func ListAll(ctx context.Context, opts ListOptions, run Runner) ([]Repo, bool, error) {
	if opts.Name == "" {
		return nil, false, errors.New("gitea: source name is required")
	}
	if run == nil {
		run = DefaultRunner
	}

	var repos []Repo
	var partial bool
	var err error

	switch opts.Hint {
	case HintOrg:
		repos, err = fetchAllPages(ctx, fmt.Sprintf("/orgs/%s/repos", opts.Name), run)
	case HintUser:
		repos, partial, err = fetchUserRepos(ctx, opts.Name, run)
	default:
		repos, err = fetchAllPages(ctx, fmt.Sprintf("/orgs/%s/repos", opts.Name), run)
		if err != nil && isNotFound(err) {
			repos, partial, err = fetchUserRepos(ctx, opts.Name, run)
		}
	}
	if err != nil {
		return nil, false, err
	}

	repos = ownedBy(repos, opts.Name)
	repos = filter(repos, opts)
	if opts.Limit > 0 && len(repos) > opts.Limit {
		repos = repos[:opts.Limit]
	}
	return repos, partial, nil
}

// ownedBy keeps the repos that owner owns. The endpoint for the authenticated
// user (/user/repos) also returns the repos of others that the user has access
// to, as a collaborator or through an organization, and those are not repos of
// the source. A repo without a full name is kept, as its owner isn't known.
func ownedBy(repos []Repo, owner string) []Repo {
	out := repos[:0]
	for _, r := range repos {
		repoOwner, _, found := strings.Cut(r.FullName, "/")
		if found && !strings.EqualFold(repoOwner, owner) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func fetchUserRepos(ctx context.Context, name string, run Runner) ([]Repo, bool, error) {
	me, _ := authenticatedUser(ctx, run)
	if me != "" && strings.EqualFold(me, name) {
		repos, err := fetchAllPages(ctx, "/user/repos", run)
		return repos, false, err
	}
	// Only the public repos of another user are listed
	repos, err := fetchAllPages(ctx, fmt.Sprintf("/users/%s/repos", name), run)
	return repos, true, err
}

func authenticatedUser(ctx context.Context, run Runner) (string, error) {
	out, err := run(ctx, "api", "user")
	if err != nil {
		return "", err
	}
	var u struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(out, &u); err != nil {
		return "", err
	}
	return u.Login, nil
}

// fetchAllPages manually paginates Gitea's `page`/`limit` query params,
// accumulating repos until a page contributes no new ones. We don't rely on a
// `--paginate` flag in `tea` since it isn't reliably available across versions.
//
// We deliberately do NOT stop at "fewer items than the requested limit": an
// instance whose MAX_RESPONSE_ITEMS is configured below pageSize caps every
// page below our ask, so a full page already looks short and the old check
// dropped every repo past the first page. Instead we stop when the server stops
// handing us new repos — which also guarantees termination against an instance
// that ignores the `page` parameter and keeps returning the same page (the old
// check happened to stop there only because that page was short). Repos are
// keyed by full name (falling back to SSH URL, then name); these are unique
// within a single listing, so de-duping never discards a distinct repo.
func fetchAllPages(ctx context.Context, basePath string, run Runner) ([]Repo, error) {
	var all []Repo
	seen := make(map[string]bool)
	for page := 1; ; page++ {
		path := fmt.Sprintf("%s?page=%d&limit=%d", basePath, page, pageSize)
		repos, err := fetchOnce(ctx, path, run)
		if err != nil {
			return nil, err
		}
		added := 0
		for _, r := range repos {
			key := repoKey(r)
			if seen[key] {
				continue
			}
			seen[key] = true
			all = append(all, r)
			added++
		}
		if added == 0 {
			break
		}
	}
	return all, nil
}

// repoKey is a stable identity for de-duplicating a repo across pages.
func repoKey(r Repo) string {
	switch {
	case r.FullName != "":
		return r.FullName
	case r.SSHURL != "":
		return r.SSHURL
	default:
		return r.Name
	}
}

func fetchOnce(ctx context.Context, path string, run Runner) ([]Repo, error) {
	out, err := run(ctx, "api", path)
	if err != nil {
		return nil, err
	}
	var repos []Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("gitea: decode response: %w", err)
	}
	return repos, nil
}

func filter(repos []Repo, opts ListOptions) []Repo {
	if opts.Archived == nil && opts.Fork == nil {
		return repos
	}
	out := repos[:0]
	for _, r := range repos {
		if opts.Archived != nil && r.Archived != *opts.Archived {
			continue
		}
		if opts.Fork != nil && r.Fork != *opts.Fork {
			continue
		}
		out = append(out, r)
	}
	return out
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	// Only what the CLI printed says what went wrong. The call itself holds
	// the name of the source, and a name such as `team404` is not a status.
	msg := err.Error()
	var cliErr *cliError
	if errors.As(err, &cliErr) {
		msg = cliErr.output
	}
	msg = strings.ToLower(msg)

	return strings.Contains(msg, "http 404") || strings.Contains(msg, "404 not found") || strings.Contains(msg, "not found")
}

// cliError is a failed call of the `tea` CLI: what was called, and what it
// printed on stderr.
type cliError struct {
	args   []string
	output string
}

func (e *cliError) Error() string {
	return fmt.Sprintf("tea %s: %s", strings.Join(e.args, " "), e.output)
}
