package dao

import (
	"strings"

	"github.com/alajmo/mani/core"
)

type TagAssocations struct {
	Name     string
	Projects []string
}

func (t TagAssocations) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return t.Name
	case "Projects", "projects":
		return strings.Join(t.Projects, "\n")
	}

	return ""
}

func (c Config) GetTags() []string {
	tags := []string{}
	for _, project := range c.ProjectList {
		for _, tag := range project.Tags {
			if !core.StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func (c Config) GetTagAssocations(tags []string) map[string]TagAssocations {
	m := make(map[string]TagAssocations)

	for _, tag := range tags {
		projects := c.GetProjectsByTags([]string{tag})
		var projectNames []string
		for _, p := range projects {
			projectNames = append(projectNames, p.Name)
		}

		m[tag] = TagAssocations{
			Name:     tag,
			Projects: projectNames,
		}
	}

	return m
}
