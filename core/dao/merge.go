package dao

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

// Helpers for combining projects, tasks and env that are defined in multiple config files.

// mergeByName groups resources by name and merges the ones that are defined in multiple config
// files. merge returns the merged resource and the index of the definition that owns it, and the
// merged resource takes the position of that definition. A resource may only be defined once per
// config file.
func mergeByName[T any, P interface {
	*T
	Resource
}](
	resources []T,
	kind string,
	name func(resource P) string,
	merge func(defs []P) (T, int, []ResourceErrors[T]),
) ([]T, []ResourceErrors[T]) {
	var names []string
	definitions := make(map[string][]int)
	for i := range resources {
		n := name(&resources[i])
		if _, found := definitions[n]; !found {
			names = append(names, n)
		}
		definitions[n] = append(definitions[n], i)
	}

	merged := make(map[int]T)
	var mergeErrors []ResourceErrors[T]
	for _, n := range names {
		indices := definitions[n]
		if len(indices) == 1 {
			merged[indices[0]] = resources[indices[0]]
			continue
		}

		defs := make([]P, len(indices))
		for i, index := range indices {
			defs[i] = &resources[index]
		}

		for i, def := range defs {
			for _, prev := range defs[:i] {
				if def.GetContext() == prev.GetContext() {
					err := &core.DefinedTwice{Kind: kind, Name: n, Other: location(prev)}
					mergeErrors = append(mergeErrors, ResourceErrors[T]{Resource: def, Errors: []error{err}})
					break
				}
			}
		}

		resource, owner, errs := merge(defs)
		merged[indices[owner]] = resource
		mergeErrors = append(mergeErrors, errs...)
	}

	var mergedResources []T
	for i := range resources {
		if resource, found := merged[i]; found {
			mergedResources = append(mergedResources, resource)
		}
	}

	return mergedResources, mergeErrors
}

// mergeByKey combines the entries of all definitions. When entries have the same key but
// different values, entries from configs that are imported by another config setting the same
// key are overridden. The remaining entries must have the same value, otherwise it's an error.
func mergeByKey[D Resource, T any](
	defs []D,
	kind string,
	name string,
	field string,
	entries func(def D) []T,
	keyValue func(entry T) (string, string),
	imports func(a string, b string) bool,
	addError func(def D, err error),
) []T {
	type candidate struct {
		entry T
		value string
		def   D
	}

	var keys []string
	candidates := make(map[string][]candidate)
	for _, def := range defs {
		for _, entry := range entries(def) {
			key, value := keyValue(entry)
			if _, found := candidates[key]; !found {
				keys = append(keys, key)
			}
			candidates[key] = append(candidates[key], candidate{entry: entry, value: value, def: def})
		}
	}

	var merged []T
	for _, key := range keys {
		remaining := notOverridden(candidates[key], func(c candidate) string { return c.def.GetContext() }, imports)

		label := fmt.Sprintf("`%s`", key)
		if field != "" {
			label = fmt.Sprintf("%s `%s`", field, key)
		}

		winner := remaining[0]
		for _, c := range remaining[1:] {
			if c.value != winner.value {
				addError(c.def, &core.Conflict{
					Kind:       kind,
					Name:       name,
					Field:      label,
					Value:      c.value,
					OtherValue: winner.value,
					Other:      location(winner.def),
					SameFile:   c.def.GetContext() == winner.def.GetContext(),
				})
			}
		}

		merged = append(merged, winner.entry)
	}

	return merged
}

// notOverridden returns the items whose config isn't imported by the config of another item
func notOverridden[T any](items []T, context func(item T) string, imports func(a string, b string) bool) []T {
	var remaining []T
	for _, item := range items {
		overridden := false
		for _, other := range items {
			if imports(context(other), context(item)) {
				overridden = true
				break
			}
		}
		if !overridden {
			remaining = append(remaining, item)
		}
	}

	// Only happens with cyclic imports, which are reported separately
	if len(remaining) == 0 {
		return items
	}

	return remaining
}

func location(resource Resource) string {
	return fmt.Sprintf("%s:%d", resource.GetContext(), resource.GetContextLine())
}

// nodeString returns a single-line representation of a YAML value, used to compare and print values
func nodeString(node *yaml.Node) string {
	if node.Kind == yaml.ScalarNode {
		return node.Value
	}

	flow := *node
	flow.Style = yaml.FlowStyle
	out, err := yaml.Marshal(&flow)
	if err != nil {
		return node.Value
	}

	return strings.TrimSpace(string(out))
}
