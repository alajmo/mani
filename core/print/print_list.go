package print

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// ProjectOutput represents a project in JSON/YAML output
type ProjectOutput struct {
	Name        string   `json:"name" yaml:"name"`
	Path        string   `json:"path" yaml:"path"`
	RelPath     string   `json:"rel_path" yaml:"rel_path"`
	Description string   `json:"description" yaml:"description"`
	URL         string   `json:"url" yaml:"url"`
	Tags        []string `json:"tags" yaml:"tags"`
}

// TaskOutput represents a task in JSON/YAML output
type TaskOutput struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Spec        string `json:"spec" yaml:"spec"`
	Target      string `json:"target" yaml:"target"`
}

// TagOutput represents a tag in JSON/YAML output
type TagOutput struct {
	Name     string   `json:"name" yaml:"name"`
	Projects []string `json:"projects" yaml:"projects"`
}

// PrintListJSON outputs a list as JSON
func PrintListJSON[T any](items []T, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

// PrintListYAML outputs a list as YAML with document separators
func PrintListYAML[T any](items []T, writer io.Writer) error {
	for i, item := range items {
		if i > 0 {
			fmt.Fprintf(writer, "---\n")
		}
		encoder := yaml.NewEncoder(writer)
		encoder.SetIndent(2)
		if err := encoder.Encode(item); err != nil {
			return err
		}
		encoder.Close()
	}
	return nil
}
