package dao

import (
	"fmt"
	"slices"
)

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
// merged by name, so the collections of the result hold the entries of all definitions, and only
// the entries have a source.
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
		for _, tag := range d.tags.value {
			if !slices.Contains(def.tags.value, tag) {
				def.tags.value = append(def.tags.value, tag)
			}
		}
		env = append(env, d.env.value...)
		remotes = append(remotes, d.remotes.value...)
		worktrees = append(worktrees, d.worktrees.value...)
	}
	def.env.value = mergeProjectEntries(m, "env", env)
	def.remotes.value = mergeProjectEntries(m, "remote", remotes)
	def.worktrees.value = mergeProjectWorktrees(m, worktrees)

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
