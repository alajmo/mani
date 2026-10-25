package dao

import (
	"fmt"
	"path/filepath"

	"github.com/alajmo/mani/core"
	"gopkg.in/yaml.v3"
)

// projectSource retains the declaring file and its raw project_dir. Paths are
// resolved only after precedence has selected the surviving definitions.
type projectSource struct {
	file, projectDir string
	line             int
}

func (s *projectSource) GetContext() string  { return s.file }
func (s *projectSource) GetContextLine() int { return s.line }
func (s projectSource) location() string     { return fmt.Sprintf("%s:%d", s.file, s.line) }
func (s projectSource) problem(err error) ResourceErrors[projectSource] {
	return ResourceErrors[projectSource]{Resource: &s, Errors: []error{err}}
}

// A zero value is different from an absent field, including false and "".
type projectValue[T any] struct {
	value  T
	set    bool
	source projectSource
}

type projectEntry[T any] struct {
	key string
	projectValue[T]
}

// Stage 1 produces only typed values, presence and provenance. No defaults,
// absolute paths or evaluated shell commands are stored in a definition.
type projectDefinition struct {
	name                           string
	source                         projectSource
	path, desc, url, clone, branch projectValue[string]
	singleBranch, sync             projectValue[bool]
	tags                           projectValue[[]string]
	env, remotes                   projectValue[[]projectEntry[string]]
	worktrees                      projectValue[[]projectEntry[Worktree]]
	pathCandidates                 []projectValue[string] // populated by stage 2
}

func parseProjectValue[T any](node *yaml.Node, source projectSource) (projectValue[T], error) {
	value := projectValue[T]{set: true, source: source}
	err := node.Decode(&value.value)
	return value, err
}

func (c Config) parseProjectDefinitions() ([]projectDefinition, []ResourceErrors[projectSource]) {
	source := projectSource{file: c.Path, projectDir: c.ProjectDir, line: c.Projects.Line}
	pairs, err := sectionPairs(&c.Projects, "projects")
	if err != nil {
		return nil, []ResourceErrors[projectSource]{source.problem(err)}
	}
	var definitions []projectDefinition
	var problems []ResourceErrors[projectSource]
	seen := make(map[string]projectSource)
	for _, pair := range pairs {
		source.line = pair.key.Line
		if pair.key.Kind != yaml.ScalarNode {
			problems = append(problems, source.problem(fmt.Errorf("project name must be a single value")))
			continue
		}
		if first, found := seen[pair.key.Value]; found {
			problems = append(problems, source.problem(fmt.Errorf("project `%s` is defined more than once in the same file, first definition is at %s", pair.key.Value, first.location())))
			continue
		}
		seen[pair.key.Value] = source
		def, errs := parseProjectDefinition(pair.key.Value, pair.value, source)
		definitions = append(definitions, def)
		problems = append(problems, errs...)
	}
	return definitions, problems
}

func parseProjectDefinition(name string, node *yaml.Node, source projectSource) (projectDefinition, []ResourceErrors[projectSource]) {
	def := projectDefinition{name: name, source: source}
	pairs, err := mappingPairs(node)
	if err != nil {
		return def, []ResourceErrors[projectSource]{source.problem(fmt.Errorf("project `%s` must be a mapping: %w", name, err))}
	}
	var problems []ResourceErrors[projectSource]
	seen := make(map[string]bool)
	for _, pair := range pairs {
		source.line = pair.key.Line
		if seen[pair.key.Value] {
			problems = append(problems, source.problem(fmt.Errorf("project `%s` sets `%s` more than once", name, pair.key.Value)))
			continue
		}
		seen[pair.key.Value] = true
		err = nil
		switch pair.key.Value {
		case "path":
			def.path, err = parseProjectValue[string](pair.value, source)
		case "desc":
			def.desc, err = parseProjectValue[string](pair.value, source)
		case "url":
			def.url, err = parseProjectValue[string](pair.value, source)
		case "clone":
			def.clone, err = parseProjectValue[string](pair.value, source)
		case "branch":
			def.branch, err = parseProjectValue[string](pair.value, source)
		case "single_branch":
			def.singleBranch, err = parseProjectValue[bool](pair.value, source)
		case "sync":
			def.sync, err = parseProjectValue[bool](pair.value, source)
		case "tags":
			def.tags, err = parseProjectValue[[]string](pair.value, source)
		case "env", "remotes":
			var entries []projectEntry[string]
			entries, err = parseProjectEntries(pair.value, source, pair.key.Value)
			if pair.key.Value == "env" {
				def.env = projectValue[[]projectEntry[string]]{value: entries, set: true, source: source}
			} else {
				def.remotes = projectValue[[]projectEntry[string]]{value: entries, set: true, source: source}
			}
		case "worktrees":
			def.worktrees.set, def.worktrees.source = true, source
			def.worktrees.value, err = parseProjectWorktrees(pair.value, source)
		}
		if err != nil {
			problems = append(problems, source.problem(err))
		}
	}
	return def, problems
}

func parseProjectEntries(node *yaml.Node, source projectSource, field string) ([]projectEntry[string], error) {
	pairs, err := mappingPairs(node)
	if err != nil {
		return nil, fmt.Errorf("`%s` must be a mapping of names to values: %w", field, err)
	}
	var entries []projectEntry[string]
	seen := make(map[string]projectValue[string])
	for _, pair := range pairs {
		source.line = pair.key.Line
		if pair.key.Kind != yaml.ScalarNode || pair.value.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("line %d: %s names and values must be single values", source.line, field)
		}
		// Preserve scalar spelling: env VERSION: 1.0 is the string "1.0".
		value := projectValue[string]{value: pair.value.Value, set: true, source: source}
		if first, found := seen[pair.key.Value]; found {
			if first.value != value.value {
				return nil, fmt.Errorf("line %d: %s `%s` is set more than once with different values (first at %s)", source.line, field, pair.key.Value, first.source.location())
			}
			continue
		}
		seen[pair.key.Value] = value
		entries = append(entries, projectEntry[string]{key: pair.key.Value, projectValue: value})
	}
	return entries, nil
}

func parseProjectWorktrees(node *yaml.Node, source projectSource) ([]projectEntry[Worktree], error) {
	node = resolveAlias(node)
	if isEmpty(node) {
		return nil, nil
	}
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("`worktrees` must be a list of worktrees")
	}
	var entries []projectEntry[Worktree]
	seen := make(map[string]projectValue[Worktree])
	for _, item := range node.Content {
		source.line = item.Line
		value, err := parseProjectValue[Worktree](item, source)
		if err != nil {
			return nil, err
		}
		if value.value.Path == "" {
			return nil, fmt.Errorf("line %d: %w", source.line, &core.WorktreePathRequired{})
		}
		key := filepath.Clean(value.value.Path)
		if first, found := seen[key]; found {
			if first.value.Branch != value.value.Branch {
				return nil, fmt.Errorf("line %d: worktree `%s` is set more than once with different branches (first at %s)", source.line, key, first.source.location())
			}
			continue
		}
		seen[key] = value
		entries = append(entries, projectEntry[Worktree]{key: key, projectValue: value})
	}
	return entries, nil
}
