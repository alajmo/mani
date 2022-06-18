package dao

import (
	"strings"

	"github.com/alajmo/mani/core"
)

type Tag struct {
	Name     string
	Projects []string
}

func (t Tag) GetValue(key string, _ int) string {
	switch key {
	case "Tag", "tag":
		return t.Name
	case "Project", "project":
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

func (c Config) GetTagAssocations(tags []string) ([]Tag, error) {
	t := []Tag{}

	for _, tag := range tags {
		projects, err := c.GetProjectsByTags([]string{tag})
		if err != nil {
			return []Tag{}, err
		}

		var projectNames []string
		for _, p := range projects {
			projectNames = append(projectNames, p.Name)
		}

		t = append(t, Tag{Name: tag, Projects: projectNames})
	}

	return t, nil
}
