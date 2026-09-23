package dao

import (
	"context"
	"fmt"
	"path"

	"github.com/alajmo/mani/core/provider/gitea"
	"github.com/alajmo/mani/core/provider/github"
	"github.com/alajmo/mani/core/sources"
)

// CheckProviderCLI returns an error if the provider's required CLI is
// not installed. Called before any fetch so the user gets a clean
// message instead of an exec failure.
func CheckProviderCLI(s sources.Source) error {
	switch s.Provider() {
	case "github":
		return github.CheckGhAvailable()
	case "gitea":
		return gitea.CheckTeaAvailable()
	}
	return fmt.Errorf("source %q has no recognized provider", s.Source)
}

// FetchAndMap fetches all repos for a source and maps them into Projects.
// Single dispatch point so InitMani and RefreshMani don't switch on
// provider type themselves.
func FetchAndMap(ctx context.Context, s sources.Source) ([]Project, error) {
	switch s.Provider() {
	case "github":
		repos, err := github.List(ctx, s.ToGitHubListOptions(), nil)
		if err != nil {
			return nil, err
		}
		mapOpts := s.ToGitHubMapOptions()
		out := make([]Project, 0, len(repos))
		for _, r := range repos {
			out = append(out, applySourcePath(s, ProjectFromGitHubRepo(r, mapOpts)))
		}
		return out, nil

	case "gitea":
		repos, err := gitea.List(ctx, s.ToGiteaListOptions(), nil)
		if err != nil {
			return nil, err
		}
		mapOpts := s.ToGiteaMapOptions()
		out := make([]Project, 0, len(repos))
		for _, r := range repos {
			out = append(out, applySourcePath(s, ProjectFromGiteaRepo(r, mapOpts)))
		}
		return out, nil
	}
	return nil, fmt.Errorf("source %q has no recognized provider", s.Source)
}

// applySourcePath prepends the source's Path to the project's path, when set,
// so all of a source's repos share a common parent directory. The repo name
// stays the project key (and is the path's final segment), so refresh's
// URL-keyed diff and name-based dedup are unaffected — only the emitted path:
// changes. path.Join (not filepath.Join) keeps the config value forward-slashed
// and portable across platforms.
func applySourcePath(s sources.Source, p Project) Project {
	if s.Path != "" {
		p.Path = path.Join(s.Path, p.Path)
	}
	return p
}
