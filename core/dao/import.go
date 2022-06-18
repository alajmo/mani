package dao

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/alajmo/mani/core"
	"github.com/jedib0t/go-pretty/v6/text"
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

// Populates SpecList and creates a default spec if no default spec is set.
func (c *Config) GetImportList() ([]Import, []ResourceErrors[Import]) {
	var imports []Import
	count := len(c.Import.Content)

	importErrors := []ResourceErrors[Import]{}
	foundErrors := false
	for i := 0; i < count; i += 1 {
		imp := &Import{
			Path:        c.Import.Content[i].Value,
			context:     c.Path,
			contextLine: c.Import.Content[i].Line,
		}

		imports = append(imports, *imp)
	}

	if foundErrors {
		return imports, importErrors
	}

	return imports, nil
}

// Used for config imports
type ConfigResources struct {
	Imports  []Import
	Themes   []Theme
	Specs    []Spec
	Targets  []Target
	Tasks    []Task
	Projects []Project
	Envs     []string

	ThemeErrors   []ResourceErrors[Theme]
	SpecErrors    []ResourceErrors[Spec]
	TargetErrors  []ResourceErrors[Target]
	TaskErrors    []ResourceErrors[Task]
	ProjectErrors []ResourceErrors[Project]
	ImportErrors  []ResourceErrors[Import]
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

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
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
	ci := ConfigResources{}
	c.loadResources(&ci)

	if c.UserConfigFile != nil {
		ci.Imports = append(ci.Imports, Import{Path: *c.UserConfigFile, context: c.Path, contextLine: -1})
	}

	// Import other configs
	n := Node{
		Path:    c.Path,
		Imports: ci.Imports,
	}
	m := make(map[string]*Node)
	m[n.Path] = &n
	cycles := []NodeLink{}

	dfs(&n, m, &cycles, &ci)

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

	if configErr != "" {
		return &core.ConfigErr{Msg: configErr}
	}

	return nil
}

func parseConfig(path string, ci *ConfigResources) ([]Import, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return []Import{}, err
	}

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
	for i := range importErrors {
		ci.ImportErrors = append(ci.ImportErrors, importErrors[i])
	}

	tasks, taskErrors := c.GetTaskList()
	for i := range taskErrors {
		ci.TaskErrors = append(ci.TaskErrors, taskErrors[i])
	}

	projects, projectErrors := c.GetProjectList()
	for i := range projectErrors {
		ci.ProjectErrors = append(ci.ProjectErrors, projectErrors[i])
	}

	themes, themeErrors := c.GetThemeList()
	for i := range themeErrors {
		ci.ThemeErrors = append(ci.ThemeErrors, themeErrors[i])
	}

	specs, specErrors := c.GetSpecList()
	for i := range specErrors {
		ci.SpecErrors = append(ci.SpecErrors, specErrors[i])
	}

	targets, targetErrors := c.GetTargetList()
	for i := range targetErrors {
		ci.TargetErrors = append(ci.TargetErrors, targetErrors[i])
	}

	envs := c.GetEnvList()

	ci.Imports = append(ci.Imports, imports...)
	ci.Tasks = append(ci.Tasks, tasks...)
	ci.Projects = append(ci.Projects, projects...)
	ci.Themes = append(ci.Themes, themes...)
	ci.Specs = append(ci.Specs, specs...)
	ci.Targets = append(ci.Targets, targets...)
	ci.Envs = append(ci.Envs, envs...)

	return imports
}

func dfs(n *Node, m map[string]*Node, cycles *[]NodeLink, ci *ConfigResources) {
	n.Visiting = true

	for i := range n.Imports {
		p, err := core.GetAbsolutePath(filepath.Dir(n.Path), n.Imports[i].Path, "")
		if err != nil {
			importError := ResourceErrors[Import]{Resource: &n.Imports[i], Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			ci.ImportErrors = append(ci.ImportErrors, importError)
			continue
		}

		// Skip visited nodes
		var nc Node
		v, exists := m[p]
		if exists {
			nc = *v
		} else {
			nc = Node{Path: p}
			m[nc.Path] = &nc
		}

		if nc.Visited {
			continue
		}

		// Found cyclic dependency
		if nc.Visiting {
			c := NodeLink{
				A: *n,
				B: nc,
			}

			*cycles = append(*cycles, c)
			break
		}

		// Import Config
		imports, err := parseConfig(nc.Path, ci)
		if err != nil {
			importError := ResourceErrors[Import]{Resource: &n.Imports[i], Errors: []error{err}}
			ci.ImportErrors = append(ci.ImportErrors, importError)
			continue
		}

		nc.Imports = imports

		dfs(&nc, m, cycles, ci)

		// err = dfs(&nc, m, cycles, ci)
		// if err != nil {
		// 	return err
		// }
	}

	n.Visiting = false
	n.Visited = true
}
