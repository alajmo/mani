package dao

import (
	"strings"
)

type Network struct {
	Name        string   `yaml:"name"`
	Hosts       []string `yaml:"hosts"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
}

func (d Network) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return d.Name
	case "Hosts", "hosts":
		return d.Hosts
	case "Description", "description":
		return d.Description
	case "Tags", "tags":
		return strings.Join(d.Tags, ", ")
	}

	return ""
}
