package dao

import (
	"strings"

	"github.com/alajmo/mani/core/provider/github"
)

// ProjectFromGitHubRepo maps a GitHub repo into a dao.Project, synthesizing
// tags per gh-mani's rules.
func ProjectFromGitHubRepo(r github.Repo, opts github.MapOptions) Project {
	p := Project{
		Name: r.Name,
		Path: r.Name,
		Desc: r.Description,
		URL:  r.SSHURL,
	}

	if !opts.NoTags {
		p.Tags = synthesizeGitHubTags(r, opts.NoTopicTags)
	}

	return p
}

// synthesizeGitHubTags converts a GitHub repo's metadata into mani tags.
//
// All synthesized tags are bare labels: `archived` and `fork` are emitted
// only when true; visibility is flattened to its value (`public`/
// `private`/`internal`); topics are added as-is. Duplicates are removed
// — if a repo has a topic literally named `archived` and is archived,
// the tag is emitted once.
//
// We deliberately don't emit the default branch: it's not useful to
// filter on, mani already follows remote HEAD, and `defaultBranch=main`
// would be the lone ugly key=value tag in the set.
func synthesizeGitHubTags(r github.Repo, noTopicTags bool) []string {
	var tags []string
	add := func(t string) {
		for _, existing := range tags {
			if existing == t {
				return
			}
		}
		tags = append(tags, t)
	}

	if r.Archived {
		add("archived")
	}
	if r.Fork {
		add("fork")
	}
	if r.Visibility != "" {
		add(strings.ToLower(r.Visibility))
	}
	if !noTopicTags {
		for _, t := range r.Topics {
			add(t)
		}
	}
	return tags
}
