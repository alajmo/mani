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
}

func (p Project) GetValue(key string) string {
	switch key {
	case "Name":
		return p.Name
	case "Path":
		return p.Path
	case "Description":
		return p.Description
	case "Url":
		return p.Url
	case "Tags":
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
	case "Name":
		return c.Name
	case "Description":
		return c.Description
	case "Shell":
		return c.Shell
	case "Command":
		return c.Command
	}

	return ""
}
