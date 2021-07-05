package core

import (
	"strings"
)

type Config struct {
	Shell    string    `yaml:"shell"`
	Projects []Project `yaml:"projects"`
	Commands []Command `yaml:"commands"`
}

type Project struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Url         string   `yaml:"url"`
	Tags        []string `yaml:"tags"`

	RelPath     string
}

func (p Project) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return p.Name
	case "Path", "path":
		return p.Path
	case "RelPath", "relpath":
		return p.RelPath
	case "Description", "description":
		return p.Description
	case "Url", "url":
		return p.Url
	case "Tags", "tags":
		return strings.Join(p.Tags, ", ")
	}

	return ""
}

type Command struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Args        map[string]string `yaml:"args"`
	Shell		string            `yaml:"shell"`
	Command     string            `yaml:"command"`
}

func (c Command) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return c.Name
	case "Description", "description":
		return c.Description
	case "Shell", "shell":
		return c.Shell
	case "Command", "command":
		return c.Command
	}

	return ""
}
