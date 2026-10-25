package dao

import (
	"fmt"
	"slices"
)

// projectCandidates is the one precedence rule for fields, map entries and
// path ownership. Only explicitly set values participate; a config overrides
// values from its transitive imports. Independent survivors must agree.
func projectCandidates[T any](values []projectValue[T], overrides func(string, string) bool) []projectValue[T] {
	var remaining []projectValue[T]
	for _, value := range values {
		if !value.set {
			continue
		}
		if slices.ContainsFunc(values, func(other projectValue[T]) bool {
			return other.set && overrides(other.source.file, value.source.file)
		}) {
			continue
		}
		remaining = append(remaining, value)
	}
	return remaining
}

func projectConflict[T any](name, field string, value, other projectValue[T]) ResourceErrors[projectSource] {
	return value.source.problem(fmt.Errorf("project `%s` has conflicting %s: `%v` here and `%v` in %s, override it in a config that imports both", name, field, value.value, other.value, other.source.location()))
}

func mergeProjectValue[T comparable](name, field string, values []projectValue[T], overrides func(string, string) bool, problems *[]ResourceErrors[projectSource]) projectValue[T] {
	remaining := projectCandidates(values, overrides)
	if len(remaining) == 0 {
		return projectValue[T]{}
	}
	for _, value := range remaining[1:] {
		if value.value != remaining[0].value {
			*problems = append(*problems, projectConflict(name, field, value, remaining[0]))
		}
	}
	return remaining[0]
}

func projectValues[T any](defs []projectDefinition, get func(projectDefinition) projectValue[T]) []projectValue[T] {
	values := make([]projectValue[T], 0, len(defs))
	for _, def := range defs {
		values = append(values, get(def))
	}
	return values
}

func mergeProjectEntries[T comparable](name, field string, entries []projectEntry[T], overrides func(string, string) bool, problems *[]ResourceErrors[projectSource]) []projectEntry[T] {
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
		value := mergeProjectValue(name, fmt.Sprintf("%s `%s`", field, key), groups[key], overrides, problems)
		merged = append(merged, projectEntry[T]{key: key, projectValue: value})
	}
	return merged
}

// Stage 2 operates only on typed definitions. It keeps unresolved path
// candidates for stage 3, where equivalent paths can be compared absolutely.
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
		paths := projectCandidates(projectValues(defs, func(d projectDefinition) projectValue[string] { return d.path }), overrides)
		if len(paths) == 0 {
			// A tag-only overlay must not move a repository defined elsewhere.
			// User definitions supply defaults only when no workspace definition exists.
			owners := slices.DeleteFunc(slices.Clone(defs), func(d projectDefinition) bool { return userConfigs[d.source.file] })
			if len(owners) == 0 {
				owners = defs
			}
			urls := projectCandidates(projectValues(owners, func(d projectDefinition) projectValue[string] { return d.url }), overrides)
			urls = slices.DeleteFunc(urls, func(v projectValue[string]) bool { return v.value == "" })
			if len(urls) != 0 {
				owners = slices.DeleteFunc(slices.Clone(owners), func(d projectDefinition) bool {
					return !slices.ContainsFunc(urls, func(v projectValue[string]) bool { return v.source.file == d.source.file })
				})
			}
			paths = projectCandidates(projectValues(owners, func(d projectDefinition) projectValue[string] {
				return projectValue[string]{set: true, source: d.source}
			}), overrides)
		}
		owner := slices.IndexFunc(defs, func(d projectDefinition) bool { return d.source.file == paths[0].source.file })
		def := projectDefinition{name: name, source: defs[owner].source, pathCandidates: paths}
		// Keep the owner's position and collection order when adding overlays.
		defs = append([]projectDefinition{defs[owner]}, slices.Delete(slices.Clone(defs), owner, owner+1)...)
		if defs[0].path.set {
			def.path = paths[0]
		}
		def.desc = mergeProjectValue(name, "`desc`", projectValues(defs, func(d projectDefinition) projectValue[string] { return d.desc }), overrides, &problems)
		def.url = mergeProjectValue(name, "`url`", projectValues(defs, func(d projectDefinition) projectValue[string] { return d.url }), overrides, &problems)
		def.clone = mergeProjectValue(name, "`clone`", projectValues(defs, func(d projectDefinition) projectValue[string] { return d.clone }), overrides, &problems)
		def.branch = mergeProjectValue(name, "`branch`", projectValues(defs, func(d projectDefinition) projectValue[string] { return d.branch }), overrides, &problems)
		def.singleBranch = mergeProjectValue(name, "`single_branch`", projectValues(defs, func(d projectDefinition) projectValue[bool] { return d.singleBranch }), overrides, &problems)
		def.sync = mergeProjectValue(name, "`sync`", projectValues(defs, func(d projectDefinition) projectValue[bool] { return d.sync }), overrides, &problems)
		for _, d := range defs {
			if d.tags.set && !def.tags.set {
				def.tags.set = true
				def.tags.source = d.tags.source
			}
			for _, tag := range d.tags.value {
				if !slices.Contains(def.tags.value, tag) {
					def.tags.value = append(def.tags.value, tag)
				}
			}
			appendProjectCollection(&def.env, d.env)
			appendProjectCollection(&def.remotes, d.remotes)
			appendProjectCollection(&def.worktrees, d.worktrees)
		}
		def.env.value = mergeProjectEntries(name, "env", def.env.value, overrides, &problems)
		def.remotes.value = mergeProjectEntries(name, "remote", def.remotes.value, overrides, &problems)
		// Worktree identity is its cleaned path. Preserve the winning spelling.
		def.worktrees.value = mergeProjectWorktreeEntries(name, def.worktrees.value, overrides, &problems)
		merged[name] = def
	}
	var result []projectDefinition
	for _, def := range definitions {
		if value := merged[def.name]; value.source == def.source {
			result = append(result, value)
		}
	}
	return result, problems
}

func mergeProjectWorktreeEntries(name string, entries []projectEntry[Worktree], overrides func(string, string) bool, problems *[]ResourceErrors[projectSource]) []projectEntry[Worktree] {
	// Compare branch values, since equivalent path spellings share the same key.
	var branches []projectEntry[string]
	for _, entry := range entries {
		branches = append(branches, projectEntry[string]{key: entry.key, projectValue: projectValue[string]{value: entry.value.Branch, set: true, source: entry.source}})
	}
	merged := mergeProjectEntries(name, "worktree", branches, overrides, problems)
	var result []projectEntry[Worktree]
	for _, branch := range merged {
		for _, entry := range entries {
			if entry.key == branch.key && entry.source == branch.source {
				result = append(result, entry)
				break
			}
		}
	}
	return result
}

// Empty collections are present but additive: they do not erase inherited entries.
func appendProjectCollection[T any](dst *projectValue[[]T], src projectValue[[]T]) {
	if src.set && !dst.set {
		dst.set, dst.source = true, src.source
	}
	dst.value = append(dst.value, src.value...)
}
