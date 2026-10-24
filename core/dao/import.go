package dao

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alajmo/mani/core"
	"github.com/gookit/color"
	"gopkg.in/yaml.v3"
)

type Import struct {
	Path string

	context     string
	contextLine int
}

func (i *Import) GetContext() string {
	return i.context
}

func (i *Import) GetContextLine() int {
	return i.contextLine
}

// GetImportList returns the config files the config imports, which must be a list. Aliases are
// supported.
func (c *Config) GetImportList() ([]Import, []ResourceErrors[Import]) {
	node := resolveAlias(&c.Import)
	if isEmpty(node) {
		return nil, nil
	}

	invalid := func(line int) []ResourceErrors[Import] {
		imp := &Import{context: c.Path, contextLine: -1}
		return []ResourceErrors[Import]{{Resource: imp, Errors: []error{&core.InvalidImport{Line: line}}}}
	}
	if node.Kind != yaml.SequenceNode {
		return nil, invalid(node.Line)
	}

	var imports []Import
	for _, item := range node.Content {
		path := resolveAlias(item)
		if path.Kind != yaml.ScalarNode || isEmpty(path) {
			return nil, invalid(item.Line)
		}

		imports = append(imports, Import{Path: path.Value, context: c.Path, contextLine: item.Line})
	}

	return imports, nil
}

// Used for config imports
type ConfigResources struct {
	Themes   []Theme
	Specs    []Spec
	Targets  []Target
	Tasks    []Task
	Projects []Project
	Envs     []string // Merged env of all configs

	envDefinitions []Env
	configFiles    []string // Absolute paths of the imported configs, in the order they're loaded

	ThemeErrors   []ResourceErrors[Theme]
	SpecErrors    []ResourceErrors[Spec]
	TargetErrors  []ResourceErrors[Target]
	TaskErrors    []ResourceErrors[Task]
	ProjectErrors []ResourceErrors[Project]
	ImportErrors  []ResourceErrors[Import]
	EnvErrors     []ResourceErrors[Env]

	ImportGraph ImportGraph
	userConfigs map[string]bool // The user config and the configs only it imports
}

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

type Node struct {
	Path     string
	Imports  []Import
	Visiting bool
	Visited  bool
}

type NodeLink struct {
	A Node
	B Node
}

type FoundCyclicDependency struct {
	Cycles []NodeLink
}

func (c *FoundCyclicDependency) Error() string {
	var msg string

	var errPrefix = color.FgRed.Sprintf("error")
	var ptrPrefix = color.FgBlue.Sprintf("-->")
	msg = fmt.Sprintf("%s: %s\n", errPrefix, "Found direct or indirect circular dependency")
	for i := range c.Cycles {
		msg += fmt.Sprintf("  %s %s\n      %s\n", ptrPrefix, c.Cycles[i].A.Path, c.Cycles[i].B.Path)
	}

	return msg
}

// Given config imports, use a Depth-first-search algorithm to recursively
// check for resources (tasks, projects, dirs, themes, specs, targets).
// A struct is passed around that is populated with resources from each config.
// In case a cyclic dependency is found (a -> b and b -> a), we return early and
// with an error containing the cyclic dependency found.
//
// This is the first parsing, later on we will perform more passes where we check what commands/tasks
// are imported.
func (c Config) importConfigs() (ConfigResources, error) {
	// Main config
	ci := ConfigResources{ImportGraph: ImportGraph{}}
	rootImports := c.loadResources(&ci)

	// Import other configs
	n := Node{
		Path:    c.Path,
		Imports: rootImports,
	}
	m := make(map[string]*Node)
	m[n.Path] = &n
	cycles := []NodeLink{}

	dfs(&n, m, &cycles, &ci)

	// The user config is an invisible import of every other config, see withUserConfigs, so it's
	// left out of the import graph. It's imported after the root config, so that a user config
	// that imports the root config isn't a cycle.
	if c.UserConfigFile != nil {
		userConfig := *c.UserConfigFile
		importConfig(userConfig, &Import{Path: userConfig, context: c.Path, contextLine: -1}, m, &cycles, &ci)
		ci.userConfigs = ci.ImportGraph.userConfigs(c.Path, userConfig)
	}

	// Combine projects, tasks and env defined in multiple configs, and pick the specs, targets and
	// themes to use
	imports := ci.ImportGraph.withUserConfigs(ci.userConfigs)
	c.userConfigs = ci.userConfigs

	projects, projectErrors := MergeProjects(ci.Projects, imports, ci.userConfigs)
	ci.Projects = projects
	ci.ProjectErrors = append(ci.ProjectErrors, projectErrors...)

	tasks, taskErrors := MergeTasks(ci.Tasks, imports, c.pathsDir)
	ci.Tasks = tasks
	ci.TaskErrors = append(ci.TaskErrors, taskErrors...)

	envs, envErrors := MergeEnvList(ci.envDefinitions, imports)
	ci.Envs = envs
	ci.EnvErrors = append(ci.EnvErrors, envErrors...)

	specs, specErrors := mergeWhole(ci.Specs, "spec",
		func(s *Spec) string { return s.Name },
		func(s *Spec) string { return nodeString(s.node) },
		imports,
	)
	ci.Specs = specs
	ci.SpecErrors = append(ci.SpecErrors, specErrors...)

	for i := range ci.Targets {
		ci.Targets[i].dir = c.pathsDir(ci.Targets[i].context)
	}
	targets, targetErrors := mergeWhole(ci.Targets, "target",
		func(t *Target) string { return t.Name },
		func(t *Target) string { return targetValue(t.node, t.dir) },
		imports,
	)
	ci.Targets = targets
	ci.TargetErrors = append(ci.TargetErrors, targetErrors...)

	themes, themeErrors := mergeWhole(ci.Themes, "theme",
		func(t *Theme) string { return t.Name },
		func(t *Theme) string { return nodeString(t.node) },
		imports,
	)
	ci.Themes = themes
	ci.ThemeErrors = append(ci.ThemeErrors, themeErrors...)

	// The definitions are only needed for merging
	for i := range ci.Specs {
		ci.Specs[i].node = nil
	}
	for i := range ci.Targets {
		ci.Targets[i].node = nil
	}
	for i := range ci.Themes {
		ci.Themes[i].node = nil
	}

	// Get errors
	configErr := concatErrors(ci, &cycles)

	if configErr != nil {
		return ci, configErr
	}

	return ci, nil
}

func concatErrors(ci ConfigResources, cycles *[]NodeLink) error {
	var configErr = ""

	if len(*cycles) > 0 {
		err := &FoundCyclicDependency{Cycles: *cycles}
		configErr = fmt.Sprintf("%s%s\n", configErr, err.Error())
	}

	for _, theme := range ci.ThemeErrors {
		if len(theme.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(theme.Resource, theme.Errors))
		}
	}

	for _, spec := range ci.SpecErrors {
		if len(spec.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(spec.Resource, spec.Errors))
		}
	}

	for _, target := range ci.TargetErrors {
		if len(target.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(target.Resource, target.Errors))
		}
	}

	for _, task := range ci.TaskErrors {
		if len(task.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(task.Resource, task.Errors))
		}
	}

	for _, project := range ci.ProjectErrors {
		if len(project.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(project.Resource, project.Errors))
		}
	}

	for _, imp := range ci.ImportErrors {
		if len(imp.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(imp.Resource, imp.Errors))
		}
	}

	for _, env := range ci.EnvErrors {
		if len(env.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(env.Resource, env.Errors))
		}
	}

	if configErr != "" {
		return &core.ConfigErr{Msg: configErr}
	}

	return nil
}

func parseConfig(path string, dat []byte, ci *ConfigResources) ([]Import, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return []Import{}, err
	}

	// Found config, now try to read it
	var config Config
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return []Import{}, err
	}

	config.Path = absPath
	config.Dir = filepath.Dir(absPath)
	imports := config.loadResources(ci)

	return imports, nil
}

func (c Config) loadResources(ci *ConfigResources) []Import {
	imports, importErrors := c.GetImportList()
	ci.ImportErrors = append(ci.ImportErrors, importErrors...)

	tasks, taskErrors := c.GetTaskList()
	ci.TaskErrors = append(ci.TaskErrors, taskErrors...)

	projects, projectErrors := c.GetProjectList()
	ci.ProjectErrors = append(ci.ProjectErrors, projectErrors...)

	themes, themeErrors := c.ParseThemes()
	ci.ThemeErrors = append(ci.ThemeErrors, themeErrors...)

	specs, specErrors := c.GetSpecList()
	ci.SpecErrors = append(ci.SpecErrors, specErrors...)

	targets, targetErrors := c.GetTargetList()
	ci.TargetErrors = append(ci.TargetErrors, targetErrors...)

	envs, envErrors := c.GetEnvList()
	ci.EnvErrors = append(ci.EnvErrors, envErrors...)

	ci.Tasks = append(ci.Tasks, tasks...)
	ci.Projects = append(ci.Projects, projects...)
	ci.Themes = append(ci.Themes, themes...)
	ci.Specs = append(ci.Specs, specs...)
	ci.Targets = append(ci.Targets, targets...)
	ci.envDefinitions = append(ci.envDefinitions, envs...)

	return imports
}

func dfs(n *Node, m map[string]*Node, cycles *[]NodeLink, ci *ConfigResources) {
	n.Visiting = true

	for i := range n.Imports {
		p, err := core.GetAbsolutePath(filepath.Dir(n.Path), n.Imports[i].Path, "")
		if err != nil {
			importError := ResourceErrors[Import]{Resource: &n.Imports[i], Errors: []error{err}}
			ci.ImportErrors = append(ci.ImportErrors, importError)
			continue
		}

		// Found cyclic dependency. The import is left out of the import graph, so that the configs
		// in the cycle don't all override each other.
		if nc, found := m[p]; found && nc.Visiting {
			c := NodeLink{
				A: *n,
				B: *nc,
			}

			*cycles = append(*cycles, c)
			break
		}

		ci.ImportGraph[n.Path] = append(ci.ImportGraph[n.Path], p)
		importConfig(p, &n.Imports[i], m, cycles, ci)
	}

	n.Visiting = false
	n.Visited = true
}

// importConfig imports the config file at path, which imp imports, and the configs it imports,
// unless it's already imported
func importConfig(path string, imp *Import, m map[string]*Node, cycles *[]NodeLink, ci *ConfigResources) {
	n, found := m[path]
	if !found {
		n = &Node{Path: path}
		m[path] = n
	}

	if n.Visited {
		return
	}

	// A config that can't be read is reported at each import of it. An invalid config is reported
	// once, at the config itself, since that's where the lines in YAML errors are.
	dat, err := os.ReadFile(path)
	if err != nil {
		importError := ResourceErrors[Import]{Resource: imp, Errors: []error{err}}
		ci.ImportErrors = append(ci.ImportErrors, importError)
		return
	}
	ci.configFiles = append(ci.configFiles, path)

	imports, err := parseConfig(path, dat, ci)
	if err != nil {
		n.Visited = true
		config := &Import{Path: imp.Path, context: path, contextLine: -1}
		importError := ResourceErrors[Import]{Resource: config, Errors: decodeErrors(err)}
		ci.ImportErrors = append(ci.ImportErrors, importError)
		return
	}

	n.Imports = imports
	dfs(n, m, cycles, ci)
}
