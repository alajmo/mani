package dao

import (
	"strings"
	"path/filepath"
)

type Dir struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`

	RelPath     string
}

func (d Dir) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return d.Name
	case "Path", "path":
		return d.Path
	case "RelPath", "relpath":
		return d.RelPath
	case "Description", "description":
		return d.Description
	case "Tags", "tags":
		return strings.Join(d.Tags, ", ")
	}

	return ""
}

func GetDirRelPath(configPath string, path string) (string, error) {
	baseDir := filepath.Dir(configPath)
	relPath, err := filepath.Rel(baseDir, path)

	return relPath, err
}
