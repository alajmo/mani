package dao

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
)

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
		return nil, fmt.Errorf("line %d: `%s` must be a mapping of names to definitions", section.Line, name)
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
				return nil, fmt.Errorf("line %d: anchor `%s` value contains itself", key.Line, mapping.Anchor)
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
