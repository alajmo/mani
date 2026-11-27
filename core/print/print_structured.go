package print

import (
	"encoding/json"
	"io"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core/dao"
)

// IsStructured reports whether output is a machine-readable format (json, jsonl or yaml).
func IsStructured(output string) bool {
	return output == "json" || output == "jsonl" || output == "yaml"
}

// PrintStructured writes all items at once: an indented JSON array for json,
// and otherwise the same as a StructuredStream.
func PrintStructured[T any](items []T, output string, writer io.Writer) error {
	if output != "json" {
		stream := NewStructuredStream(output, writer)
		for _, item := range items {
			if err := stream.Write(item); err != nil {
				return err
			}
		}
		return stream.Close()
	}

	// A nil slice would be encoded as `null`
	if items == nil {
		items = []T{}
	}

	encoder := newJSONEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

// StructuredStream writes items one at a time, as they become available: one
// compact JSON object per line (jsonl), or one YAML document per item separated
// by `---` (yaml). It is safe for concurrent use.
type StructuredStream struct {
	mu      sync.Mutex
	json    *json.Encoder
	yaml    *yaml.Encoder
	written bool
}

func NewStructuredStream(output string, writer io.Writer) *StructuredStream {
	if output == "yaml" {
		encoder := yaml.NewEncoder(writer)
		encoder.SetIndent(2)
		return &StructuredStream{yaml: encoder}
	}

	return &StructuredStream{json: newJSONEncoder(writer)}
}

func (s *StructuredStream) Write(item any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.written = true

	if s.yaml != nil {
		// The encoder writes the `---` separator between documents
		return s.yaml.Encode(item)
	}

	return s.json.Encode(item)
}

func (s *StructuredStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Closing a yaml encoder that has not encoded anything is an error
	if s.yaml != nil && s.written {
		return s.yaml.Close()
	}

	return nil
}

func newJSONEncoder(writer io.Writer) *json.Encoder {
	encoder := json.NewEncoder(writer)
	// Keep characters such as <, > and & readable, the output is not embedded in HTML
	encoder.SetEscapeHTML(false)
	return encoder
}

// ProjectOutput is a project in json/yaml list output.
type ProjectOutput struct {
	Name        string           `json:"name" yaml:"name"`
	Path        string           `json:"path" yaml:"path"`
	RelPath     string           `json:"rel_path" yaml:"rel_path"`
	Description string           `json:"description" yaml:"description"`
	URL         string           `json:"url" yaml:"url"`
	Tags        []string         `json:"tags" yaml:"tags"`
	Worktrees   []WorktreeOutput `json:"worktrees" yaml:"worktrees"`
}

// WorktreeOutput is a worktree of a project in json/yaml list output.
type WorktreeOutput struct {
	Path   string `json:"path" yaml:"path"`
	Branch string `json:"branch" yaml:"branch"`
}

// TaskOutput is a task in json/yaml list output.
type TaskOutput struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Spec        string `json:"spec" yaml:"spec"`
	Target      string `json:"target" yaml:"target"`
}

// TagOutput is a tag in json/yaml list output.
type TagOutput struct {
	Name     string   `json:"name" yaml:"name"`
	Projects []string `json:"projects" yaml:"projects"`
}

func ProjectsOutput(projects []dao.Project) []ProjectOutput {
	out := make([]ProjectOutput, len(projects))
	for i, p := range projects {
		worktrees := make([]WorktreeOutput, len(p.WorktreeList))
		for j, wt := range p.WorktreeList {
			worktrees[j] = WorktreeOutput{Path: wt.Path, Branch: wt.Branch}
		}

		out[i] = ProjectOutput{
			Name:        p.Name,
			Path:        p.Path,
			RelPath:     p.RelPath,
			Description: p.Desc,
			URL:         p.URL,
			Tags:        nonNil(p.Tags),
			Worktrees:   worktrees,
		}
	}
	return out
}

func TasksOutput(tasks []dao.Task) []TaskOutput {
	out := make([]TaskOutput, len(tasks))
	for i, t := range tasks {
		out[i] = TaskOutput{
			Name:        t.Name,
			Description: t.Desc,
			Spec:        t.SpecData.Name,
			Target:      t.TargetData.Name,
		}
	}
	return out
}

func TagsOutput(tags []dao.Tag) []TagOutput {
	out := make([]TagOutput, len(tags))
	for i, t := range tags {
		out[i] = TagOutput{
			Name:     t.Name,
			Projects: nonNil(t.Projects),
		}
	}
	return out
}

// nonNil makes sure an empty list is encoded as `[]` and not `null`.
func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
