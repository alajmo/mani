package dao

import (
	"reflect"
	"sort"
)

// Helper functions

func getProjectNames(projects []Project) []string {
	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.Name
	}
	sort.Strings(names)
	return names
}

func getTreePaths(nodes []TreeNode) []string {
	paths := make([]string, len(nodes))
	for i, node := range nodes {
		paths[i] = node.Path
	}
	sort.Strings(paths)
	return paths
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	return reflect.DeepEqual(aCopy, bCopy)
}
