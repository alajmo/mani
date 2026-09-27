package dao

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

// Helpers for combining projects, tasks and env that are defined in multiple config files.

// mergeByName groups resources by name and merges the ones that are defined in multiple config
// files. merge returns the merged resource and the index of the definition that owns it, and the
// merged resource takes the position of that definition. A resource may only be defined once per
// config file, later definitions in the same file are reported and left out.
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
		var defs []P
		var indices []int
		for _, index := range definitions[n] {
			var def P = &resources[index]
			first := slices.IndexFunc(defs, func(prev P) bool { return prev.GetContext() == def.GetContext() })
			if first >= 0 {
				err := &core.DefinedTwice{Kind: kind, Name: n, Other: location(defs[first])}
				mergeErrors = append(mergeErrors, ResourceErrors[T]{Resource: def, Errors: []error{err}})
				continue
			}
			defs = append(defs, def)
			indices = append(indices, index)
		}

		if len(defs) == 1 {
			merged[indices[0]] = resources[indices[0]]
			continue
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

// noImports is used to merge the entries of a single definition, where a key that's set twice with
// different values is a conflict
func noImports(string, string) bool {
	return false
}

func location(resource Resource) string {
	return fmt.Sprintf("%s:%d", resource.GetContext(), resource.GetContextLine())
}

// source is the location of the definition a value of a merged resource comes from, used to
// report errors in the value at the config that sets it
type source struct {
	context     string
	contextLine int
}

func sourceOf(resource Resource) source {
	return source{context: resource.GetContext(), contextLine: resource.GetContextLine()}
}

// nodeString returns a single-line representation of a YAML value, used to compare and print values.
// The value is decoded and encoded again, so quoting, comments, key order, aliases and merge keys
// don't affect it.
func nodeString(node *yaml.Node) string {
	node = resolveAlias(node)
	if node.Kind == yaml.ScalarNode {
		return node.Value
	}

	var value any
	if err := node.Decode(&value); err != nil {
		return node.Value
	}

	var flow yaml.Node
	if err := flow.Encode(value); err != nil {
		return node.Value
	}
	flow.Style = yaml.FlowStyle
	out, err := yaml.Marshal(&flow)
	if err != nil {
		return node.Value
	}

	return strings.TrimSpace(string(out))
}

func resolveAlias(node *yaml.Node) *yaml.Node {
	for node.Kind == yaml.AliasNode && node.Alias != nil {
		node = node.Alias
	}
	return node
}

// pair is a key and its value in a YAML mapping
type pair struct {
	key   *yaml.Node
	value *yaml.Node
}

// mappingPairs returns the key-value pairs of a YAML mapping with aliases resolved, including the
// pairs from merge keys (<<). Like when decoding, pairs set in the mapping take precedence over
// merged pairs, and earlier merged mappings take precedence over later ones. Merged pairs take the
// position of the merge key. An empty value has no pairs, and anything else that isn't a mapping,
// or merges something that isn't a mapping, returns false.
func mappingPairs(node *yaml.Node) ([]pair, bool) {
	node = resolveAlias(node)
	if node.Kind == 0 || node.Tag == "!!null" {
		return nil, true
	}
	if node.Kind != yaml.MappingNode {
		return nil, false
	}

	set := make(map[string]bool)
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Tag != "!!merge" {
			set[node.Content[i].Value] = true
		}
	}

	var pairs []pair
	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Tag != "!!merge" {
			pairs = append(pairs, pair{key: key, value: resolveAlias(value)})
			continue
		}

		for _, mapping := range mergedMappings(value) {
			merged, ok := mappingPairs(mapping)
			if !ok {
				return nil, false
			}
			for _, p := range merged {
				if !set[p.key.Value] {
					set[p.key.Value] = true
					pairs = append(pairs, p)
				}
			}
		}
	}

	return pairs, true
}

// mergedMappings returns the mappings the value of a merge key (<<) merges, in order of precedence
func mergedMappings(value *yaml.Node) []*yaml.Node {
	value = resolveAlias(value)
	if value.Kind != yaml.SequenceNode {
		return []*yaml.Node{value}
	}

	var mappings []*yaml.Node
	for _, node := range value.Content {
		mappings = append(mappings, resolveAlias(node))
	}
	return mappings
}
