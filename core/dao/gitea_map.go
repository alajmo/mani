package dao

import (
	"github.com/alajmo/mani/core/provider/gitea"
)

// ProjectFromGiteaRepo maps a Gitea repo into a dao.Project. Same shape
// as ProjectFromGitHubRepo — the synthesized tag schema is identical.
func ProjectFromGiteaRepo(r gitea.Repo, opts gitea.MapOptions) Project {
	p := Project{
		Name: r.Name,
		Path: r.Name,
		Desc: r.Description,
		URL:  r.SSHURL,
	}
	if !opts.NoTags {
		p.Tags = synthesizeGiteaTags(r, opts.NoTopicTags)
	}
	return p
}

func synthesizeGiteaTags(r gitea.Repo, noTopicTags bool) []string {
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
	add(r.Visibility())
	if !noTopicTags {
		for _, t := range r.Topics {
			add(t)
		}
	}
	return tags
}
