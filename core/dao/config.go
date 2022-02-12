package dao

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	color "github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

var (
	Version               = "dev"
	DEFAULT_SHELL         = "bash -c"
	ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani", ".mani.yaml", ".mani.yml", "Manifile", "Manifile.yaml", "Manifile.yml"}

	DEFAULT_THEME = Theme {
		Name:  "default",

		Table: "ascii",
		Tree:  "line",
	}

	DEFAULT_TARGET = Target {
		Name:     "default",

		All:      false,
		Projects: []string{},
		Paths:    []string{},
		Tags:     []string{},
		Cwd:      false,
	}

	DEFAULT_SPEC = Spec {
		Name:        "default",

		Output:      "text",
		Parallel:    false,
		IgnoreError: false,
		OmitEmpty:   false,
	}
)

type Config struct {
	// User Defined
	Import      []string `yaml:"import"`
	EnvList     []string
	ThemeList   []Theme
	SpecList    []Spec
	TargetList    []Target
	ProjectList []Project
	TaskList    []Task
	Shell       string `yaml:"shell"`

	// Intermediate
	Env      yaml.Node `yaml:"env"`
	Themes   yaml.Node `yaml:"themes"`
	Specs    yaml.Node `yaml:"specs"`
	Targets  yaml.Node `yaml:"targets"`
	Projects yaml.Node `yaml:"projects"`
	Tasks    yaml.Node `yaml:"tasks"`

	// Internal
	Path string
	Dir  string
	UserConfigFile string
}

// Used for config imports
type ConfigResources struct {
	Themes   []Theme
	Specs    []Spec
	Targets  []Target
	Tasks    []Task
	Projects []Project
	Envs     []string
}

func (c Config) GetEnvList() []string {
	var envs []string
	count := len(c.Env.Content)
	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", c.Env.Content[i].Value, c.Env.Content[i+1].Value)
		envs = append(envs, env)
	}

	return envs
}

func createUserConfigDirIfNotExist(userConfigDir string) string {
	userConfigFile := filepath.Join(userConfigDir, "config.yaml")
	if _, err := os.Stat(userConfigDir); os.IsNotExist(err) {
		err := os.MkdirAll(userConfigDir, os.ModePerm)
		core.CheckIfError(err)

		if _, err := os.Stat(userConfigFile); os.IsNotExist(err) {
			err := ioutil.WriteFile(userConfigFile, []byte(""), 0644)
			core.CheckIfError(err)
		}
	}

	return userConfigFile
}

// Function to read Mani configs.
func ReadConfig(cfgName string, userConfigDir string) (Config, error) {
	var configPath string

	userConfigFile := createUserConfigDirIfNotExist(userConfigDir)

	// Try to find config file in current directory and all parents
	if cfgName != "" {
		filename, err := filepath.Abs(cfgName)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return Config{}, err
		}

		filename, err := core.FindFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
		if err != nil {
			return Config{}, err
		}

		filename, err = filepath.Abs(filename)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	}

	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	// Found config, now try to read it
	var config Config

	config.Path = configPath
	config.Dir = filepath.Dir(configPath)
	config.UserConfigFile = userConfigFile

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		return config, &core.FailedToParseFile{Name: configPath, Msg: err}
	}

	// Set default shell command
	if config.Shell == "" {
		config.Shell = DEFAULT_SHELL
	} else {
		config.Shell = core.FormatShell(config.Shell)
	}

	configResources, err := config.importConfigs()
	if err != nil {
		return config, err
	}

	config.TaskList = configResources.Tasks
	config.ProjectList = configResources.Projects
	config.ThemeList = configResources.Themes
	config.SpecList = configResources.Specs
	config.TargetList = configResources.Targets
	config.EnvList = configResources.Envs

	// Set default config if it's not set already
	_, err = config.GetTheme(DEFAULT_THEME.Name)
	if err != nil {
		config.ThemeList = append(config.ThemeList, DEFAULT_THEME)
	}

	// Set default config if it's not set already
	_, err = config.GetSpec(DEFAULT_SPEC.Name)
	if err != nil {
		config.SpecList = append(config.SpecList, DEFAULT_SPEC)
	}

	// Set default config if it's not set already
	_, err = config.GetTarget(DEFAULT_TARGET.Name)
	if err != nil {
		config.TargetList = append(config.TargetList, DEFAULT_TARGET)
	}

	// Parse all tasks
	for i := range configResources.Tasks {
		configResources.Tasks[i].ParseTask(config)
	}

	return config, nil
}

func (c Config) loadResources(ci *ConfigResources) error {
	tasks, err := c.GetTaskList()
	if err != nil {
		return err
	}

	projects, err := c.GetProjectList()
	if err != nil {
		return err
	}

	themes, err := c.GetThemeList()
	if err != nil {
		return err
	}

	specs, err := c.GetSpecList()
	if err != nil {
		return err
	}

	targets, err := c.GetTargetList()
	if err != nil {
		return err
	}

	envs := c.GetEnvList()

	ci.Tasks = append(ci.Tasks, tasks...)
	ci.Projects = append(ci.Projects, projects...)
	ci.Themes = append(ci.Themes, themes...)
	ci.Specs = append(ci.Specs, specs...)
	ci.Targets = append(ci.Targets, targets...)
	ci.Envs = append(ci.Envs, envs...)

	return nil
}

// Given config imports, use a Depth-first-search algorithm to recursively
// check for resources (tasks, projects, dirs, themes, specs, targets).
// A struct is passed around that is populated with resources from each config.
// In case a cyclic dependency is found (a -> b and b -> a), we return early and
// with an error containing the cyclic dependency found.
func (c Config) importConfigs() (ConfigResources, error) {
	imports := append(c.Import, c.UserConfigFile)
	n := core.Node{
		Path:    c.Path,
		Imports: imports,
	}

	m := make(map[string]*core.Node)
	m[n.Path] = &n
	cycles := []core.NodeLink{}

	ci := ConfigResources{}
	err :=  c.loadResources(&ci)
	if err != nil {
		return ci, err
	}

	err = dfs(&n, m, &cycles, &ci)

	if err != nil {
		return ci, err
	} else if len(cycles) > 0 {
		return ci, &core.FoundCyclicDependency{Cycles: cycles}
	} else {
		return ci, nil
	}
}

func dfs(n *core.Node, m map[string]*core.Node, cycles *[]core.NodeLink, ci *ConfigResources) error {
	n.Visiting = true

	for _, importPath := range n.Imports {
		p, err := core.GetAbsolutePath(filepath.Dir(n.Path), importPath, "")
		if err != nil {
			return err
		}

		// Skip visited nodes
		var nc core.Node
		v, exists := m[p]
		if exists {
			nc = *v
		} else {
			nc = core.Node{Path: p}
			m[nc.Path] = &nc
		}

		if nc.Visited {
			continue
		}

		// Found cyclic dependency
		if nc.Visiting {
			c := core.NodeLink{
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

	err = config.loadResources(ci)
	if err != nil {
		return []string{}, &core.FailedToParseFile{Name: path, Msg: err}
	}

	return config.Import, nil
}

// Open mani config in editor
func (c Config) EditConfig() {
	openEditor(c.Path, -1)
}

// Open mani config in editor and optionally go to line matching the task name
func (c Config) EditTask(name string) {
	configPath := c.Path
	if name != "" {
		task, err := c.GetTask(name)
		core.CheckIfError(err)
		configPath = task.Context
	}

	dat, err := ioutil.ReadFile(configPath)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Tasks yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if name == "" {
		lineNr = configTmp.Tasks.Line - 1
	} else {
		for _, task := range configTmp.Tasks.Content {
			if task.Value == name {
				lineNr = task.Line
				break
			}
		}
	}

	openEditor(configPath, lineNr)
}

// Open mani config in editor and optionally go to line matching the project name
func (c Config) EditProject(name string) {
	configPath := c.Path
	if name != "" {
		project, err := c.GetProject(name)
		core.CheckIfError(err)
		configPath = project.Context
	}

	dat, err := ioutil.ReadFile(configPath)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Projects yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if name == "" {
		lineNr = configTmp.Projects.Line - 1
	} else {
		for _, project := range configTmp.Projects.Content {
			if project.Value == name {
				lineNr = project.Line
				break
			}
		}
	}

	openEditor(configPath, lineNr)
}

func openEditor(path string, lineNr int) {
	editor := os.Getenv("EDITOR")
	var args []string

	if lineNr > 0 {
		switch editor {
		case "vim":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "vi":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "emacs":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "nano":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
		case "code": // visual studio code
			args = []string{"--goto", fmt.Sprintf("%s:%v", path, lineNr)}
		case "idea": // Intellij
			args = []string{"--line", fmt.Sprintf("%v", lineNr), path}
		case "subl": // Sublime
			args = []string{fmt.Sprintf("%s:%v", path, lineNr)}
		case "atom":
			args = []string{fmt.Sprintf("%s:%v", path, lineNr)}
		case "notepad-plus-plus":
			args = []string{"-n", fmt.Sprintf("%v", lineNr), path}
		default:
			args = []string{path}
		}
	} else {
		args = []string{path}
	}

	cmd := exec.Command(editor, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	core.CheckIfError(err)
}

func InitMani(args []string, initFlags core.InitFlags) {
	// Choose to initialize mani in a different directory
	// 1. absolute or
	// 2. relative or
	// 3. working directory
	var configDir string
	if len(args) > 0 && filepath.IsAbs(args[0]) {
		configDir = args[0]
	} else if len(args) > 0 {
		wd, err := os.Getwd()
		core.CheckIfError(err)
		configDir = filepath.Join(wd, args[0])
	} else {
		wd, err := os.Getwd()
		core.CheckIfError(err)
		configDir = wd
	}

	err := os.MkdirAll(configDir, os.ModePerm)
	core.CheckIfError(err)

	configPath := filepath.Join(configDir, "mani.yaml")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("fatal: %q is already a mani directory\n", configDir)
		os.Exit(1)
	}

	url := core.GetWdRemoteUrl(configDir)
	rootName := filepath.Base(configDir)
	rootPath := "."
	rootUrl := url
	rootProject := Project{Name: rootName, Path: rootPath, Url: rootUrl}
	projects := []Project{rootProject}
	if initFlags.AutoDiscovery {
		prs, err := FindVCSystems(configDir)

		if err != nil {
			fmt.Println(err)
		}

		projects = append(projects, prs...)
	}

	funcMap := template.FuncMap{
		"projectItem": func(name string, path string, url string) string {
			var txt = name + ":"

			if name != path {
				txt = txt + "\n    path: " + path
			}

			if url != "" {
				txt = txt + "\n    url: " + url
			}

			return txt
		},
	}

	tmpl, err := template.New("init").Funcs(funcMap).Parse(`projects:
  {{- range .}}
  {{ (projectItem .Name .Path .Url) }}
  {{ end }}
tasks:
  hello:
    desc: Print Hello World
    cmd: echo "Hello World"
`,
	)

	core.CheckIfError(err)

	// Create mani.yaml
	f, err := os.Create(configPath)
	core.CheckIfError(err)

	err = tmpl.Execute(f, projects)
	core.CheckIfError(err)

	f.Close()
	fmt.Println(color.Green("\u2713"), "Initialized mani repository in", configDir)

	hasUrl := false
	for _, project := range projects {
		if project.Url != "" {
			hasUrl = true
			break
		}
	}

	if hasUrl && initFlags.Vcs == "git"  {
		// Add gitignore file
		gitignoreFilepath := filepath.Join(configDir, ".gitignore")
		if _, err := os.Stat(gitignoreFilepath); os.IsNotExist(err) {
			err := ioutil.WriteFile(gitignoreFilepath, []byte(""), 0644)

			core.CheckIfError(err)
		}

		var projectNames []string
		for _, project := range projects {
			if project.Url == "" {
				continue
			}

			if project.Path == "." {
				continue
			}

			projectNames = append(projectNames, project.Path)
		}

		// Add projects to gitignore file
		err = UpdateProjectsToGitignore(projectNames, gitignoreFilepath)
		core.CheckIfError(err)
	}
}

func (c Config) SyncProjects(configDir string, parallelFlag bool) {
	// Get relative project names for gitignore file
	var projectNames []string
	for _, project := range c.ProjectList {
		if project.Url == "" {
			continue
		}

		if project.Path == "." {
			continue
		}

		// Project must be below mani config file to be added to gitignore
		projectPath, _ := core.GetAbsolutePath(c.Path, project.Path, project.Name)
		if !strings.HasPrefix(projectPath, configDir) {
			continue
		}

		if project.Path != "" {
			relPath, _ := filepath.Rel(configDir, projectPath)
			projectNames = append(projectNames, relPath)
		} else {
			projectNames = append(projectNames, project.Name)
		}
	}

	// Only add projects to gitignore if a .gitignore file exists in the mani.yaml directory
	gitignoreFilename := filepath.Join(filepath.Dir(c.Path), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); err == nil {
		core.CheckIfError(err)

		err := UpdateProjectsToGitignore(projectNames, gitignoreFilename)
		core.CheckIfError(err)
	}

	c.CloneRepos(parallelFlag)
}
