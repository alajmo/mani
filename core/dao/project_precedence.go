package dao

// ImportGraph maps each config file to the config files it imports
type ImportGraph map[string][]string

// importClosure maps each config file to the config files it imports, directly or indirectly
type importClosure map[string]map[string]bool

// imports returns true if config file a imports config file b, directly or indirectly
func (c importClosure) imports(a string, b string) bool {
	return c[a][b]
}

// imported returns the config files that config file path imports, directly or indirectly, except
// through config file skip
func (g ImportGraph) imported(path string, skip string) map[string]bool {
	imported := make(map[string]bool)
	var visit func(path string)
	visit = func(path string) {
		for _, next := range g[path] {
			if next != skip && !imported[next] {
				imported[next] = true
				visit(next)
			}
		}
	}
	visit(path)

	return imported
}

// closure returns the config files each config file imports, directly or indirectly. It's computed
// once, since merging checks it for every pair of definitions of every value.
func (g ImportGraph) closure() importClosure {
	closure := make(importClosure)
	for path := range g {
		closure[path] = g.imported(path, "")
	}

	return closure
}

// userConfigs returns the user config and the configs only it imports, meaning the configs that the
// root config doesn't import other than through the user config. The root config is never one.
func (g ImportGraph) userConfigs(root string, user string) map[string]bool {
	if user == root {
		return nil
	}

	workspace := g.imported(root, user)
	userConfigs := map[string]bool{user: true}
	for path := range g.imported(user, "") {
		if path != root && !workspace[path] {
			userConfigs[path] = true
		}
	}

	return userConfigs
}

// withUserConfigs returns whether config file a imports config file b, directly or indirectly. The
// user config is an invisible import of every other config, so the user configs, meaning the user
// config and the configs only it imports, are imported by every other config, and their own imports
// of other configs don't count.
func (g ImportGraph) withUserConfigs(userConfigs map[string]bool) func(a string, b string) bool {
	graph := ImportGraph{}
	for path, imports := range g {
		for _, imported := range imports {
			if !userConfigs[path] || userConfigs[imported] {
				graph[path] = append(graph[path], imported)
			}
		}
	}
	closure := graph.closure()

	return func(a string, b string) bool {
		if userConfigs[a] != userConfigs[b] {
			return userConfigs[b]
		}
		return closure.imports(a, b)
	}
}
