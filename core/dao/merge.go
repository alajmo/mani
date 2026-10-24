package dao

import (
	"errors"
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

// mergeWhole merges resources that are used as a whole, such as specs, when they're defined in
// multiple config files. The definition from the config that imports the configs of the others
// is used, and definitions from configs that don't import each other must have the same value.
// A resource may only be defined once per config file.
func mergeWhole[T any, P interface {
	*T
	Resource
}](
	resources []T,
	kind string,
	name func(resource P) string,
	value func(resource P) string,
	imports func(a string, b string) bool,
) ([]T, []ResourceErrors[T]) {
	return mergeByName(resources, kind, name, func(defs []P) (T, int, []ResourceErrors[T]) {
		remaining := notOverridden(defs, func(def P) string { return def.GetContext() }, imports)

		var mergeErrors []ResourceErrors[T]
		winner := remaining[0]
		winnerValue := value(winner)
		for _, def := range remaining[1:] {
			if v := value(def); v != winnerValue {
				err := &core.Conflict{Kind: kind, Name: name(def), Field: "definitions", Value: v, OtherValue: winnerValue, Other: location(winner)}
				mergeErrors = append(mergeErrors, ResourceErrors[T]{Resource: def, Errors: []error{err}})
			}
		}

		return *winner, slices.Index(defs, winner), mergeErrors
	})
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

	// Doesn't happen, since imports that close a cycle are left out of the import graph
	if len(remaining) == 0 {
		return items
	}

	return remaining
}

// keyValues checks that each key of an env, remotes or worktrees field is only set once in a
// definition, or set again to the same value
type keyValues struct {
	field string
	first map[string]keyValue
}

type keyValue struct {
	value string
	line  int
}

func newKeyValues(field string) keyValues {
	return keyValues{field: field, first: make(map[string]keyValue)}
}

// add returns true if the key isn't set yet, and an error if it's already set to a different value
func (k keyValues) add(key string, value string, line int) (bool, error) {
	first, found := k.first[key]
	if !found {
		k.first[key] = keyValue{value: value, line: line}
		return true, nil
	}

	if first.value != value {
		return false, &core.SetTwice{Field: k.field, Name: key, Value: value, OtherValue: first.value, OtherLine: first.line, Line: line}
	}

	return false, nil
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

// isEmpty returns true for a YAML value that isn't set, is null or is an empty string
func isEmpty(node *yaml.Node) bool {
	return node.Kind == 0 || node.Tag == "!!null" || (node.Kind == yaml.ScalarNode && node.Value == "")
}

// errNotMapping is returned by mappingPairs for a value that isn't a mapping
var errNotMapping = errors.New("not a mapping")

// sectionPairs returns the definitions in a section of a config, such as projects, as pairs of
// names and definitions. Aliases and merge keys (<<) are supported, see mappingPairs.
func sectionPairs(section *yaml.Node, name string) ([]pair, error) {
	pairs, err := mappingPairs(section)
	if errors.Is(err, errNotMapping) {
		return nil, &core.InvalidSection{Name: name, Line: section.Line}
	}

	return pairs, err
}

// mappingPairs returns the key-value pairs of a YAML mapping with aliases resolved, including the
// pairs from merge keys (<<). Like when decoding, pairs set in the mapping take precedence over
// merged pairs, and earlier merged mappings take precedence over later ones. Merged pairs take the
// position of the merge key. An empty value has no pairs, and anything else that isn't a mapping,
// or merges something that isn't a mapping, returns errNotMapping.
func mappingPairs(node *yaml.Node) ([]pair, error) {
	return mappingPairsMerging(node, make(map[*yaml.Node]bool))
}

// mappingPairsMerging returns the pairs of a mapping, see mappingPairs. merging are the mappings
// whose merge keys are being resolved, so a mapping that merges one of them contains itself. They're
// identified by their first key, since a yaml.Node field holds a copy of the mapping it's set to.
func mappingPairsMerging(node *yaml.Node, merging map[*yaml.Node]bool) ([]pair, error) {
	node = resolveAlias(node)
	if isEmpty(node) {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, errNotMapping
	}
	if len(node.Content) == 0 {
		return nil, nil
	}

	merging[node.Content[0]] = true
	defer delete(merging, node.Content[0])

	keys := make([]*yaml.Node, 0, len(node.Content)/2)
	set := make(map[string]bool)
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Kind == yaml.AliasNode {
			// The line of an alias is where it's used
			resolved := *resolveAlias(key)
			resolved.Line = key.Line
			key = &resolved
		}
		keys = append(keys, key)
		if key.Tag != "!!merge" {
			set[key.Value] = true
		}
	}

	var pairs []pair
	for i, key := range keys {
		value := node.Content[2*i+1]
		if key.Tag != "!!merge" {
			pairs = append(pairs, pair{key: key, value: resolveAlias(value)})
			continue
		}

		for _, mapping := range mergedMappings(value) {
			// Like when decoding, merging anything but a mapping is an error, including null
			if mapping.Kind != yaml.MappingNode {
				return nil, errNotMapping
			}
			if len(mapping.Content) > 0 && merging[mapping.Content[0]] {
				return nil, &core.AnchorContainsItself{Anchor: mapping.Anchor, Line: key.Line}
			}

			merged, err := mappingPairsMerging(mapping, merging)
			if err != nil {
				return nil, err
			}
			for _, p := range merged {
				if !set[p.key.Value] {
					set[p.key.Value] = true
					pairs = append(pairs, p)
				}
			}
		}
	}

	return pairs, nil
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
