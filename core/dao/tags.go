package dao

import (
	"strings"

	"github.com/alajmo/mani/core"
)

type TagAssocations struct {
	Name     string
	Projects []string
	Dirs     []string
}

func (t TagAssocations) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return t.Name
	case "Projects", "projects":
		return strings.Join(t.Projects, "\n")
	case "Dirs", "dirs", "directories":
		return strings.Join(t.Dirs, "\n")
	}

	return ""
}

func (c Config) GetTagsByProject(projectNames []string) []string {
	tags := []string{}
	for _, project := range c.ProjectList {
		if core.StringInSlice(project.Name, projectNames) {
			tags = append(tags, project.Tags...)
		}
	}

	return tags
}

func (c Config) GetTagsByDir(names []string) []string {
	tags := []string{}
	for _, dir := range c.DirList {
		if core.StringInSlice(dir.Name, names) {
			tags = append(tags, dir.Tags...)
		}
	}

	return tags
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

	for _, dir := range c.DirList {
		for _, tag := range dir.Tags {
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

		dirs := c.GetDirsByTags([]string{tag})
		var dirNames []string
		for _, d := range dirs {
			dirNames = append(dirNames, d.Name)
		}

		m[tag] = TagAssocations{
			Name:     tag,
			Projects: projectNames,
			Dirs:     dirNames,
		}
	}

	return m
}
