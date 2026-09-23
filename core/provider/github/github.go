// Package github fetches repository metadata from GitHub by shelling out
// to the `gh` CLI. We don't take a Go SDK dependency: `gh` already handles
// auth, pagination, and rate limiting.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Repo struct {
	Name          string   `json:"name"`
	SSHURL        string   `json:"ssh_url"`
	CloneURL      string   `json:"clone_url"`
	Description   string   `json:"description"`
	DefaultBranch string   `json:"default_branch"`
	Visibility    string   `json:"visibility"`
	Archived      bool     `json:"archived"`
	Fork          bool     `json:"fork"`
	Topics        []string `json:"topics"`
}

// Hint tells List whether the name refers to an org, a user, or
// (default) something to auto-detect.
type Hint string

const (
	HintAuto Hint = ""
	HintOrg  Hint = "org"
	HintUser Hint = "user"
)

type ListOptions struct {
	Name     string // org or user
	Hint     Hint   // "" = auto-detect, "org" = skip user probe, "user" = skip org probe
	Limit    int    // 0 means no limit
	Archived *bool  // nil = no filter
	Fork     *bool  // nil = no filter
}

// MapOptions controls how a Repo is rendered into a project entry.
// The mapping itself lives in the dao package (it produces dao.Project,
// which we can't import from here without a cycle).
type MapOptions struct {
	NoTags      bool
	NoTopicTags bool
}

// Runner shells out to `gh`. Tests inject a fake.
type Runner func(ctx context.Context, args ...string) ([]byte, error)

// DefaultRunner invokes the real `gh` binary.
func DefaultRunner(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("gh %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("gh %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// CheckGhAvailable returns nil if `gh` is on PATH, else a friendly error.
func CheckGhAvailable() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("`gh` CLI is required for GitHub sources but was not found on PATH.\nInstall it from https://cli.github.com and run `gh auth login`")
	}
	return nil
}

// List fetches repos for opts.Name. With the default Hint (auto), it
// tries the org endpoint first and falls back to a user endpoint on 404.
// Explicit hints skip the probe:
//
//   - HintOrg: only /orgs/<name>/repos.
//   - HintUser: only the user endpoint. If opts.Name matches the
//     currently authenticated `gh` user, we use /user/repos?affiliation=owner
//     (returns private repos too); otherwise /users/<name>/repos (public
//     repos only, by GitHub API design).
func List(ctx context.Context, opts ListOptions, run Runner) ([]Repo, error) {
	if opts.Name == "" {
		return nil, errors.New("github: source name is required")
	}
	if run == nil {
		run = DefaultRunner
	}

	var repos []Repo
	var err error

	switch opts.Hint {
	case HintOrg:
		repos, err = fetch(ctx, fmt.Sprintf("/orgs/%s/repos?per_page=100", opts.Name), run)
	case HintUser:
		repos, err = fetchUserRepos(ctx, opts.Name, run)
	default: // HintAuto
		repos, err = fetch(ctx, fmt.Sprintf("/orgs/%s/repos?per_page=100", opts.Name), run)
		if err != nil && isNotFound(err) {
			repos, err = fetchUserRepos(ctx, opts.Name, run)
		}
	}
	if err != nil {
		return nil, err
	}

	repos = filter(repos, opts)
	if opts.Limit > 0 && len(repos) > opts.Limit {
		repos = repos[:opts.Limit]
	}
	return repos, nil
}

// fetchUserRepos picks the right user endpoint based on whether the name
// matches the auth'd user. Caller uses this whenever we've decided it's
// a user-shaped query (either by hint or by org-fallback).
func fetchUserRepos(ctx context.Context, name string, run Runner) ([]Repo, error) {
	me, _ := authenticatedUser(ctx, run)
	if me != "" && strings.EqualFold(me, name) {
		return fetch(ctx, "/user/repos?per_page=100&affiliation=owner", run)
	}
	return fetch(ctx, fmt.Sprintf("/users/%s/repos?per_page=100", name), run)
}

// authenticatedUser returns the login of the user `gh` is currently
// authenticated as. Errors are swallowed (an unauthenticated or
// misconfigured `gh` is handled later by the actual list call).
func authenticatedUser(ctx context.Context, run Runner) (string, error) {
	out, err := run(ctx, "api", "user", "-q", ".login")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func fetch(ctx context.Context, path string, run Runner) ([]Repo, error) {
	out, err := run(ctx, "api", "--paginate", path)
	if err != nil {
		return nil, err
	}
	// `gh api --paginate` concatenates JSON arrays from each page by
	// stripping the closing/opening brackets between pages, producing a
	// single valid JSON array. Decode it directly.
	var repos []Repo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("github: decode response: %w", err)
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

// isNotFound checks if a `gh api` error came from a 404. `gh` writes the
// HTTP status into stderr, which DefaultRunner has already folded into the
// error message.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "HTTP 404") || strings.Contains(msg, "Not Found")
}
