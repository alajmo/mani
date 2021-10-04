package dao

import (
	"github.com/alajmo/mani/core"
)

func (c Config) GetTagsByProject(projectNames []string) []string {
	tags := []string{}
	for _, project := range c.Projects {
		if core.StringInSlice(project.Name, projectNames) {
			tags = append(tags, project.Tags...)
		}
	}

	return tags
}

func (c Config) GetTagsByDir(names []string) []string {
	tags := []string{}
	for _, dir := range c.Dirs {
		if core.StringInSlice(dir.Name, names) {
			tags = append(tags, dir.Tags...)
		}
	}

	return tags
}

func (c Config) GetTags() []string {
	tags := []string{}
	for _, project := range c.Projects {
		for _, tag := range project.Tags {
			if !core.StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	for _, dir := range c.Dirs {
		for _, tag := range dir.Tags {
			if !core.StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	for _, network := range c.NetworkList {
		for _, tag := range network.Tags {
			if !core.StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}
