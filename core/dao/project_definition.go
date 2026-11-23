package dao

import (
	"cmp"
	"fmt"
	"maps"
	"path/filepath"
	"slices"

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
func (s projectSource) problem(errs ...error) ResourceErrors[projectSource] {
	return ResourceErrors[projectSource]{Resource: &s, Errors: errs}
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
	tags                           []string
	env, remotes                   []projectEntry[string]
	worktrees                      []projectEntry[Worktree]
	pathCandidates                 []projectValue[string] // populated by stage 2
}

// projectYAML is a project as it's written in a config. The YAML library decodes it, which
// resolves aliases and merge keys (<<), and reports keys that are set more than once. The fields
// are nodes, since a node has the line of its value, and a field that isn't set has an empty node,
// unlike a field that's set to false or to an empty string.
type projectYAML struct {
	Path         yaml.Node `yaml:"path"`
	Desc         yaml.Node `yaml:"desc"`
	URL          yaml.Node `yaml:"url"`
	Clone        yaml.Node `yaml:"clone"`
	Branch       yaml.Node `yaml:"branch"`
	SingleBranch yaml.Node `yaml:"single_branch"`
	Sync         yaml.Node `yaml:"sync"`
	Tags         yaml.Node `yaml:"tags"`
	Env          yaml.Node `yaml:"env"`
	Remotes      yaml.Node `yaml:"remotes"`
	Worktrees    yaml.Node `yaml:"worktrees"`
}

func (c Config) parseProjectDefinitions() ([]projectDefinition, []ResourceErrors[projectSource]) {
	source := projectSource{file: c.Path, projectDir: c.ProjectDir, line: c.Projects.Line}

	// Decoded only to have the library check the section: that it's a mapping, and that the names
	// of the projects are single values that are set once
	var section map[string]yaml.Node
	if err := c.Projects.Decode(&section); err != nil {
		return nil, []ResourceErrors[projectSource]{source.problem(err)}
	}

	// The names are read from the section itself, to keep the order of the projects and the line
	// of each name
	var definitions []projectDefinition
	var problems []ResourceErrors[projectSource]
	pairs := resolveAlias(&c.Projects).Content
	for i := 0; i+1 < len(pairs); i += 2 {
		name, node := resolveAlias(pairs[i]), pairs[i+1]
		source.line = pairs[i].Line
		if name.Tag == "!!merge" {
			problems = append(problems, source.problem(fmt.Errorf("a merge key (<<) can be used in a project, but not to add projects")))
			continue
		}

		def, errs := parseProjectDefinition(name.Value, node, source)
		definitions = append(definitions, def)
		if len(errs) > 0 {
			problems = append(problems, source.problem(errs...))
		}
	}
	return definitions, problems
}

func parseProjectDefinition(name string, node *yaml.Node, source projectSource) (projectDefinition, []error) {
	def := projectDefinition{name: name, source: source}
	var fields projectYAML
	if err := node.Decode(&fields); err != nil {
		return def, []error{err}
	}

	var errs []error
	def.path = parseProjectValue[string](fields.Path, source, &errs)
	def.desc = parseProjectValue[string](fields.Desc, source, &errs)
	def.url = parseProjectValue[string](fields.URL, source, &errs)
	def.clone = parseProjectValue[string](fields.Clone, source, &errs)
	def.branch = parseProjectValue[string](fields.Branch, source, &errs)
	def.singleBranch = parseProjectValue[bool](fields.SingleBranch, source, &errs)
	def.sync = parseProjectValue[bool](fields.Sync, source, &errs)
	def.tags = parseProjectValue[[]string](fields.Tags, source, &errs).value
	def.env = parseProjectEntries(fields.Env, source, "env", &errs)
	def.remotes = parseProjectEntries(fields.Remotes, source, "remotes", &errs)
	def.worktrees = parseProjectWorktrees(fields.Worktrees, source, &errs)
	return def, errs
}

// parseProjectValue returns the value of a field. A field that isn't set has no value, and neither
// has a field that's set to null, such as `path:` with nothing after it.
func parseProjectValue[T any](node yaml.Node, source projectSource, errs *[]error) projectValue[T] {
	if isNull(&node) {
		return projectValue[T]{}
	}

	source.line = node.Line
	value := projectValue[T]{set: true, source: source}
	if err := node.Decode(&value.value); err != nil {
		*errs = append(*errs, err)
	}
	return value
}

// parseProjectEntries returns the entries of env or remotes. A value is kept as it's written, so
// `VERSION: 1.0` is 1.0 and `TOKEN:` is empty. The library decodes the mapping into a map, which
// has no order, so the entries are ordered by where their values are in the config.
func parseProjectEntries(node yaml.Node, source projectSource, field string, errs *[]error) []projectEntry[string] {
	var values map[string]yaml.Node
	if err := node.Decode(&values); err != nil {
		*errs = append(*errs, err)
		return nil
	}

	names := slices.SortedFunc(maps.Keys(values), func(a, b string) int {
		return cmp.Or(cmp.Compare(values[a].Line, values[b].Line), cmp.Compare(values[a].Column, values[b].Column))
	})

	var entries []projectEntry[string]
	for _, name := range names {
		node := values[name]
		source.line = node.Line
		value := resolveAlias(&node)
		if value.Kind != yaml.ScalarNode {
			*errs = append(*errs, fmt.Errorf("line %d: %s values must be single values", source.line, field))
			continue
		}
		entries = append(entries, projectEntry[string]{
			key:          name,
			projectValue: projectValue[string]{value: value.Value, set: true, source: source},
		})
	}
	return entries
}

// parseProjectWorktrees returns the worktrees by their cleaned path, which is what identifies a
// worktree, so `./wt` and `wt` are the same worktree and must have the same branch
func parseProjectWorktrees(node yaml.Node, source projectSource, errs *[]error) []projectEntry[Worktree] {
	var nodes []yaml.Node
	if err := node.Decode(&nodes); err != nil {
		*errs = append(*errs, err)
		return nil
	}

	var entries []projectEntry[Worktree]
	for _, node := range nodes {
		source.line = node.Line
		worktree := projectValue[Worktree]{set: true, source: source}
		if err := node.Decode(&worktree.value); err != nil {
			*errs = append(*errs, err)
			continue
		}
		if worktree.value.Path == "" {
			*errs = append(*errs, fmt.Errorf("line %d: %w", source.line, &core.WorktreePathRequired{}))
			continue
		}

		key := filepath.Clean(worktree.value.Path)
		first := slices.IndexFunc(entries, func(entry projectEntry[Worktree]) bool { return entry.key == key })
		if first == -1 {
			entries = append(entries, projectEntry[Worktree]{key: key, projectValue: worktree})
		} else if entries[first].value.Branch != worktree.value.Branch {
			*errs = append(*errs, fmt.Errorf("line %d: worktree `%s` is set more than once with different branches (first at %s)", source.line, key, entries[first].source.location()))
		}
	}
	return entries
}

// resolveAlias returns the node an alias refers to, or the node itself when it isn't an alias
func resolveAlias(node *yaml.Node) *yaml.Node {
	for node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}
	return node
}

// isNull returns true for a node that isn't set or is null
func isNull(node *yaml.Node) bool {
	node = resolveAlias(node)
	return node.Kind == 0 || node.Tag == "!!null"
}
