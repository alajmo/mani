package dao

import (
	"io/ioutil"
	"path/filepath"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"
	"github.com/alajmo/mani/core"
)

// Used for config imports
type ConfigResources struct {
	Themes   []Theme
	Specs    []Spec
	Targets  []Target
	Tasks    []Task
	Projects []Project
	Envs     []string

	ThemeErrors []ResourceErrors[Theme]
	SpecErrors []ResourceErrors[Spec]
	TargetErrors []ResourceErrors[Target]
	TaskErrors []ResourceErrors[Task]
	ProjectErrors []ResourceErrors[Project]
}

type Node struct {
	Path     string
	Imports  []string
	Children []*Node
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
	var msg = ""

	var errPrefix = text.FgRed.Sprintf("error")
	var ptrPrefix = text.FgBlue.Sprintf("-->")
	msg = fmt.Sprintf("%s: %s\n", errPrefix, "Found direct or indirect circular dependency")
	for i := range c.Cycles {
		msg += fmt.Sprintf("  %s %s\n  %s %s\n", ptrPrefix, c.Cycles[i].A.Path, ptrPrefix, c.Cycles[i].B.Path)
	}

	return msg
}

// Given config imports, use a Depth-first-search algorithm to recursively
// check for resources (tasks, projects, dirs, themes, specs, targets).
// A struct is passed around that is populated with resources from each config.
// In case a cyclic dependency is found (a -> b and b -> a), we return early and
// with an error containing the cyclic dependency found.
func (c Config) importConfigs() (ConfigResources, error) {
	var imports = c.Import
	if c.UserConfigFile != nil  {
		imports = append(imports, *c.UserConfigFile)
	}

	n := Node{
		Path:    c.Path,
		Imports: imports,
	}

	m := make(map[string]*Node)
	m[n.Path] = &n
	cycles := []NodeLink{}

	ci := ConfigResources{}
	c.loadResources(&ci)

	err := dfs(&n, m, &cycles, &ci)
	// TODO: Combine all errors into a string here and return it

	configErr := combineErrors(ci)
	fmt.Printf(configErr)

	if err != nil {
		return ci, err
	} else if len(cycles) > 0 {
		return ci, &FoundCyclicDependency{Cycles: cycles}
	} else {
		return ci, nil
	}
}

func combineErrors(ci ConfigResources) string {
	var configErr = ""

	for _, theme := range ci.ThemeErrors {
		if len(theme.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, CombineErrors(theme.Resource, theme.Errors))
		}
	}

	for _, spec := range ci.SpecErrors {
		if len(spec.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, CombineErrors(spec.Resource, spec.Errors))
		}
	}

	for _, target := range ci.TargetErrors {
		if len(target.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, CombineErrors(target.Resource, target.Errors))
		}
	}

	for _, task := range ci.TaskErrors {
		if len(task.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, CombineErrors(task.Resource, task.Errors))
		}
	}

	for _, project := range ci.ProjectErrors {
		if len(project.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, CombineErrors(project.Resource, project.Errors))
		}
	}

	return configErr
}

func importConfig(path string, ci *ConfigResources) ([]string, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return []string{}, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return []string{}, err
	}

	// Found config, now try to read it
	var config Config
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return []string{}, &core.FailedToParseFile{Name: path, Msg: err}
	}

	config.Path = absPath
	config.Dir = filepath.Dir(absPath)

	config.loadResources(ci)

	return config.Import, nil
}

func (c Config) loadResources(ci *ConfigResources) {
	tasks, err := c.GetTaskList()
	if err != nil {
		// return err
	}

	projects, projectErrors := c.GetProjectList()
	if projectErrors != nil {
		for i := range projectErrors {
			ci.ProjectErrors = append(ci.ProjectErrors, projectErrors[i])
		}
	}

	themes, themeErrors := c.GetThemeList()
	if themeErrors != nil {
		for i := range themeErrors {
			ci.ThemeErrors = append(ci.ThemeErrors, themeErrors[i])
		}
	}

	specs, specErrors := c.GetSpecList()
	if specErrors != nil {
		for i := range specErrors {
			ci.SpecErrors = append(ci.SpecErrors, specErrors[i])
		}
	}

	targets, targetErrors := c.GetTargetList()
	if targetErrors != nil {
		for i := range targetErrors {
			ci.TargetErrors = append(ci.TargetErrors, targetErrors[i])
		}
	}

	envs := c.GetEnvList()

	ci.Tasks = append(ci.Tasks, tasks...)
	ci.Projects = append(ci.Projects, projects...)
	ci.Themes = append(ci.Themes, themes...)
	ci.Specs = append(ci.Specs, specs...)
	ci.Targets = append(ci.Targets, targets...)
	ci.Envs = append(ci.Envs, envs...)
}

func dfs(n *Node, m map[string]*Node, cycles *[]NodeLink, ci *ConfigResources) error {
	n.Visiting = true

	for _, importPath := range n.Imports {
		p, err := core.GetAbsolutePath(filepath.Dir(n.Path), importPath, "")
		if err != nil {
			return err
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

		// Import Data
		imports, err := importConfig(nc.Path, ci)
		if err != nil {
			return err
		}

		nc.Imports = imports

		err = dfs(&nc, m, cycles, ci)
		if err != nil {
			return err
		}
	}

	n.Visiting = false
	n.Visited = true

	return nil
}
