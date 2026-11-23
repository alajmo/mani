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

// Stage 2 merges the definitions of each project into one definition. It operates only on typed
// definitions: nothing is executed, and paths are kept unresolved for stage 3, where equivalent
// paths can be compared absolutely.

// projectMerge merges the definitions of one project
type projectMerge struct {
	name string

	// The definitions of the project. The one that owns its path is first, so that its tags, env,
	// remotes and worktrees come before the ones the other definitions add.
	defs []projectDefinition

	// overrides returns true if config file a overrides config file b
	overrides func(a string, b string) bool

	problems *[]ResourceErrors[projectSource]
}

func mergeProjectDefinitions(definitions []projectDefinition, overrides func(string, string) bool, userConfigs map[string]bool) ([]projectDefinition, []ResourceErrors[projectSource]) {
	var names []string
	groups := make(map[string][]projectDefinition)
	for _, def := range definitions {
		if _, found := groups[def.name]; !found {
			names = append(names, def.name)
		}
		groups[def.name] = append(groups[def.name], def)
	}

	var problems []ResourceErrors[projectSource]
	merged := make(map[string]projectDefinition)
	for _, name := range names {
		defs := groups[name]
		paths := projectPathCandidates(defs, overrides, userConfigs)
		owner := slices.IndexFunc(defs, func(d projectDefinition) bool { return d.source.file == paths[0].source.file })
		defs = slices.Concat(defs[owner:owner+1], defs[:owner], defs[owner+1:])

		m := &projectMerge{name: name, defs: defs, overrides: overrides, problems: &problems}
		merged[name] = m.merge(paths)
	}

	// A project takes the position of the definition that owns its path
	var result []projectDefinition
	for _, def := range definitions {
		if project := merged[def.name]; project.source == def.source {
			result = append(result, project)
		}
	}
	return result, problems
}

// merge returns the definitions of the project merged into one. Each field takes the value that
// wins precedence, see mergeProjectValue. Tags are combined, and env, remotes and worktrees are
// merged by name.
func (m *projectMerge) merge(paths []projectValue[string]) projectDefinition {
	owner := m.defs[0]
	def := projectDefinition{name: m.name, source: owner.source, path: owner.path, pathCandidates: paths}

	def.desc = mergeProjectField(m, "desc", func(d projectDefinition) projectValue[string] { return d.desc })
	def.url = mergeProjectField(m, "url", func(d projectDefinition) projectValue[string] { return d.url })
	def.clone = mergeProjectField(m, "clone", func(d projectDefinition) projectValue[string] { return d.clone })
	def.branch = mergeProjectField(m, "branch", func(d projectDefinition) projectValue[string] { return d.branch })
	def.singleBranch = mergeProjectField(m, "single_branch", func(d projectDefinition) projectValue[bool] { return d.singleBranch })
	def.sync = mergeProjectField(m, "sync", func(d projectDefinition) projectValue[bool] { return d.sync })

	var env, remotes []projectEntry[string]
	var worktrees []projectEntry[Worktree]
	for _, d := range m.defs {
		for _, tag := range d.tags {
			if !slices.Contains(def.tags, tag) {
				def.tags = append(def.tags, tag)
			}
		}
		env = append(env, d.env...)
		remotes = append(remotes, d.remotes...)
		worktrees = append(worktrees, d.worktrees...)
	}
	def.env = mergeProjectEntries(m, "env", env)
	def.remotes = mergeProjectEntries(m, "remote", remotes)
	def.worktrees = mergeProjectWorktrees(m, worktrees)

	return def
}

// projectPathCandidates returns the values that decide the path of a project, for stage 3 to
// resolve and compare. A path that's set decides it, with the same precedence as other fields.
// When no definition sets a path, the project is placed by its name beside one of its definitions,
// and the candidates are empty paths from the definitions that may place it:
//
//  1. The definitions in the workspace. A user config places a project only when no workspace
//     config defines it.
//  2. Of those, the definitions that supply the url, when the project has one. A definition that
//     only adds tags doesn't move a repository that's defined elsewhere.
//  3. Of those, the definitions that win precedence.
func projectPathCandidates(defs []projectDefinition, overrides func(string, string) bool, userConfigs map[string]bool) []projectValue[string] {
	paths := projectValues(defs, func(d projectDefinition) projectValue[string] { return d.path })
	if paths = projectCandidates(paths, overrides); len(paths) > 0 {
		return paths
	}

	owners := projectDefinitionsWhere(defs, func(d projectDefinition) bool { return !userConfigs[d.source.file] })
	if len(owners) == 0 {
		owners = defs
	}

	urls := projectValues(owners, func(d projectDefinition) projectValue[string] { return d.url })
	urls = projectCandidates(urls, overrides)
	suppliers := projectDefinitionsWhere(owners, func(d projectDefinition) bool {
		return slices.ContainsFunc(urls, func(url projectValue[string]) bool {
			return url.value != "" && url.source.file == d.source.file
		})
	})
	if len(suppliers) > 0 {
		owners = suppliers
	}

	paths = projectValues(owners, func(d projectDefinition) projectValue[string] {
		return projectValue[string]{set: true, source: d.source}
	})
	return projectCandidates(paths, overrides)
}

// projectCandidates is the one precedence rule for fields, map entries and path ownership. Only
// values that are set take part, and a config overrides the values from the configs it imports,
// directly or indirectly. What remains are values from configs that don't override each other,
// which must agree.
func projectCandidates[T any](values []projectValue[T], overrides func(string, string) bool) []projectValue[T] {
	var remaining []projectValue[T]
	for _, value := range values {
		if !value.set {
			continue
		}
		overridden := slices.ContainsFunc(values, func(other projectValue[T]) bool {
			return other.set && overrides(other.source.file, value.source.file)
		})
		if !overridden {
			remaining = append(remaining, value)
		}
	}
	return remaining
}

func projectConflict[T any](name, field string, value, other projectValue[T]) ResourceErrors[projectSource] {
	return value.source.problem(fmt.Errorf("project `%s` has conflicting %s: `%v` here and `%v` in %s, override it in a config that imports both", name, field, value.value, other.value, other.source.location()))
}

// mergeProjectValue returns the value that wins precedence, see projectCandidates, or a value
// that isn't set when none of the values are. Values that remain but differ are a conflict.
func mergeProjectValue[T comparable](m *projectMerge, field string, values []projectValue[T]) projectValue[T] {
	remaining := projectCandidates(values, m.overrides)
	if len(remaining) == 0 {
		return projectValue[T]{}
	}
	for _, value := range remaining[1:] {
		if value.value != remaining[0].value {
			*m.problems = append(*m.problems, projectConflict(m.name, field, value, remaining[0]))
		}
	}
	return remaining[0]
}

// mergeProjectField merges the values the definitions of the project have for a field
func mergeProjectField[T comparable](m *projectMerge, field string, get func(projectDefinition) projectValue[T]) projectValue[T] {
	return mergeProjectValue(m, "`"+field+"`", projectValues(m.defs, get))
}

// mergeProjectEntries merges the entries of a collection, such as env, by their key. The entries
// keep the order their keys first appear in.
func mergeProjectEntries[T comparable](m *projectMerge, field string, entries []projectEntry[T]) []projectEntry[T] {
	var keys []string
	groups := make(map[string][]projectValue[T])
	for _, entry := range entries {
		if _, found := groups[entry.key]; !found {
			keys = append(keys, entry.key)
		}
		groups[entry.key] = append(groups[entry.key], entry.projectValue)
	}

	var merged []projectEntry[T]
	for _, key := range keys {
		value := mergeProjectValue(m, fmt.Sprintf("%s `%s`", field, key), groups[key])
		merged = append(merged, projectEntry[T]{key: key, projectValue: value})
	}
	return merged
}

// mergeProjectWorktrees merges worktrees by their cleaned path. Worktrees with the same path can
// spell it differently, so it's their branches that are merged, and the worktree whose branch wins
// is kept, with its spelling of the path.
func mergeProjectWorktrees(m *projectMerge, worktrees []projectEntry[Worktree]) []projectEntry[Worktree] {
	branches := make([]projectEntry[string], len(worktrees))
	for i, worktree := range worktrees {
		branch := projectValue[string]{value: worktree.value.Branch, set: true, source: worktree.source}
		branches[i] = projectEntry[string]{key: worktree.key, projectValue: branch}
	}

	var merged []projectEntry[Worktree]
	for _, branch := range mergeProjectEntries(m, "worktree", branches) {
		winner := slices.IndexFunc(worktrees, func(worktree projectEntry[Worktree]) bool {
			return worktree.key == branch.key && worktree.source == branch.source
		})
		merged = append(merged, worktrees[winner])
	}
	return merged
}

// projectValues returns the value each definition has for a field
func projectValues[T any](defs []projectDefinition, get func(projectDefinition) projectValue[T]) []projectValue[T] {
	values := make([]projectValue[T], 0, len(defs))
	for _, def := range defs {
		values = append(values, get(def))
	}
	return values
}

// projectDefinitionsWhere returns the definitions that keep returns true for
func projectDefinitionsWhere(defs []projectDefinition, keep func(projectDefinition) bool) []projectDefinition {
	var kept []projectDefinition
	for _, def := range defs {
		if keep(def) {
			kept = append(kept, def)
		}
	}
	return kept
}

// Stage 3 resolves surviving paths and applies defaults. Evaluation is a
// separate final pass so an invalid config never executes project env commands.
func resolveProjects(definitions []projectDefinition, rootDir string) ([]Project, []ResourceErrors[projectSource]) {
	var projects []Project
	var problems []ResourceErrors[projectSource]
	for _, def := range definitions {
		project := Project{Name: def.name, Desc: def.desc.value, URL: def.url.value,
			Clone: def.clone.value, Branch: def.branch.value, Tags: def.tags,
			context: def.source.file, contextLine: def.source.line,
			Sync: core.Ptr(true), SingleBranch: core.Ptr(false)}
		var resolved []projectValue[string]
		for _, candidate := range def.pathCandidates {
			base, err := core.GetAbsolutePath(filepath.Dir(candidate.source.file), candidate.source.projectDir, "")
			if err == nil {
				candidate.value, err = core.GetAbsolutePath(base, candidate.value, def.name)
			}
			if err != nil {
				problems = append(problems, candidate.source.problem(err))
				continue
			}
			candidate.value = filepath.Clean(candidate.value)
			resolved = append(resolved, candidate)
		}
		if len(resolved) == 0 {
			continue
		}
		project.Path = resolved[0].value
		for _, path := range resolved[1:] {
			if path.value == project.Path {
				continue
			}
			if def.path.set {
				problems = append(problems, projectConflict(def.name, "`path`", path, resolved[0]))
			} else {
				problems = append(problems, path.source.problem(fmt.Errorf("project `%s` resolves to `%s` here and `%s` in %s, set `path` in one of the definitions", def.name, path.value, project.Path, resolved[0].source.location())))
			}
		}
		var err error
		project.RelPath, err = core.GetRelativePath(rootDir, project.Path)
		if err != nil {
			problems = append(problems, def.source.problem(err))
		}
		if def.sync.set {
			project.Sync = core.Ptr(def.sync.value)
		}
		if def.singleBranch.set {
			project.SingleBranch = core.Ptr(def.singleBranch.value)
		}
		for _, remote := range def.remotes {
			project.RemoteList = append(project.RemoteList, Remote{Name: remote.key, URL: remote.value})
		}
		for _, entry := range def.worktrees {
			worktree := entry.value
			if worktree.Branch == "" {
				worktree.Branch = filepath.Base(worktree.Path)
			}
			project.WorktreeList = append(project.WorktreeList, worktree)
		}
		projects = append(projects, project)
	}
	return projects, problems
}

func evaluateProjectEnv(projects []Project, definitions []projectDefinition) []ResourceErrors[projectSource] {
	var problems []ResourceErrors[projectSource]
	index := make(map[string]int, len(projects))
	for i, project := range projects {
		index[project.Name] = i
	}
	for _, def := range definitions {
		i, found := index[def.name]
		if !found {
			continue
		}
		for _, env := range def.env {
			values, err := EvaluateEnv([]string{env.key + "=" + env.value})
			if err != nil {
				problems = append(problems, env.source.problem(err))
				continue
			}
			projects[i].EnvList = append(projects[i].EnvList, values[0])
		}
	}
	return problems
}
