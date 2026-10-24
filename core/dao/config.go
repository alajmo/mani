package dao

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alajmo/mani/core"
	"github.com/gookit/color"
	"gopkg.in/yaml.v3"
)

var (
	DEFAULT_SHELL         = "bash -c"
	DEFAULT_SHELL_PROGRAM = "bash"
	ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani.yaml", ".mani.yml"}

	DEFAULT_THEME = Theme{
		Name:   "default",
		Stream: DefaultStream,
		Table:  DefaultTable,
		Tree:   DefaultTree,
		TUI:    DefaultTUI,
		Block:  DefaultBlock,
		Color:  new(true),
	}

	DEFAULT_TARGET = Target{
		Name: "default",

		All: false,
		Cwd: false,

		Projects: []string{},
		Paths:    []string{},
		Tags:     []string{},

		TagsExpr: "",
	}

	DEFAULT_SPEC = Spec{
		Name:   "default",
		Output: "stream",

		Parallel: false,
		Forks:    4,

		IgnoreErrors:      false,
		IgnoreNonExisting: false,

		OmitEmptyRows:    false,
		OmitEmptyColumns: false,

		ClearOutput: true,
	}
)

type Config struct {
	// Internal
	EnvList        []string  `yaml:"-"`
	ImportData     []Import  `yaml:"-"`
	ThemeList      []Theme   `yaml:"-"`
	SpecList       []Spec    `yaml:"-"`
	TargetList     []Target  `yaml:"-"`
	ProjectList    []Project `yaml:"-"`
	TaskList       []Task    `yaml:"-"`
	Path           string    `yaml:"-"`
	Dir            string    `yaml:"-"`
	UserConfigFile *string   `yaml:"-"`
	ConfigPaths    []string  `yaml:"-"` // Absolute paths of the imported configs, including the user config
	Color          bool      `yaml:"-"`

	userConfigs map[string]bool // The user config and the configs only it imports

	Shell                   string `yaml:"shell"`
	SyncRemotes             *bool  `yaml:"sync_remotes"`
	SyncGitignore           *bool  `yaml:"sync_gitignore"`
	RemoveOrphanedWorktrees *bool  `yaml:"remove_orphaned_worktrees"`
	ReloadTUI               *bool  `yaml:"reload_tui_on_change"`

	// Directory that relative project paths in this file resolve against.
	// Unlike the settings above, it's read from every config file, including imports.
	ProjectDir string `yaml:"project_dir"`

	// Intermediate
	Env      yaml.Node `yaml:"env"`
	Import   yaml.Node `yaml:"import"`
	Themes   yaml.Node `yaml:"themes"`
	Specs    yaml.Node `yaml:"specs"`
	Targets  yaml.Node `yaml:"targets"`
	Projects yaml.Node `yaml:"projects"`
	Tasks    yaml.Node `yaml:"tasks"`
}

func (c *Config) GetContext() string {
	return c.Path
}

func (c *Config) GetContextLine() int {
	return -1
}

// Returns the env section of the config
func (c Config) GetEnvList() ([]Env, []ResourceErrors[Env]) {
	envs, err := parseEnv(&c.Env, c.Path)
	if err != nil {
		env := &Env{context: c.Path, contextLine: c.Env.Line}
		return nil, []ResourceErrors[Env]{{Resource: env, Errors: []error{err}}}
	}

	return envs, nil
}

// MergeEnvList combines the env sections of all configs into a string slice in the form
// [key=value, key1=$(echo 123)]. If an env has different values, the value from the config that
// imports the other wins, and if neither config imports the other, it's an error.
func MergeEnvList(envs []Env, imports func(a string, b string) bool) ([]string, []ResourceErrors[Env]) {
	var envErrors []ResourceErrors[Env]
	defs := make([]*Env, len(envs))
	for i := range envs {
		defs[i] = &envs[i]
	}

	merged := mergeByKey(defs, "", "", "env",
		func(e *Env) []*Env { return []*Env{e} },
		func(e *Env) (string, string) { return e.Name, e.Value },
		imports,
		func(e *Env, err error) {
			envErrors = append(envErrors, ResourceErrors[Env]{Resource: e, Errors: []error{err}})
		},
	)

	var envList []string
	for _, env := range merged {
		envList = append(envList, env.String())
	}

	return envList, envErrors
}

// absolutePath returns the absolute path of a config file from a flag or env, where env variables and
// ~ are expanded, and a relative path is relative to the working directory
func absolutePath(path string) (string, error) {
	path, err := core.ResolveTildePath(path)
	if err != nil {
		return "", err
	}

	return filepath.Abs(path)
}

// getUserConfigFile returns the absolute path of the user config, or nil if there's none. A relative
// path from the flag or env is relative to the working directory, like the config. An empty
// MANI_USER_CONFIG counts as not set.
func getUserConfigFile(userConfigPath string) (*string, error) {
	// Flag
	if userConfigPath != "" {
		path, err := absolutePath(userConfigPath)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(path); err == nil {
			return &path, nil
		}
	}

	// Env
	if val := os.Getenv("MANI_USER_CONFIG"); val != "" {
		path, err := absolutePath(val)
		if err != nil {
			return nil, err
		}
		return &path, nil
	}

	// Default
	defaultUserConfigDir, _ := os.UserConfigDir()
	defaultUserConfigPath := filepath.Join(defaultUserConfigDir, "mani", "config.yaml")
	if _, err := os.Stat(defaultUserConfigPath); err == nil {
		return &defaultUserConfigPath, nil
	}

	return nil, nil
}

// Function to read Mani configs.
func ReadConfig(configFilepath string, userConfigPath string, colorFlag bool) (Config, error) {
	color := CheckUserColor(colorFlag)
	var configPath string

	userConfigFile, err := getUserConfigFile(userConfigPath)
	if err != nil {
		return Config{}, err
	}

	// Try to find config file in current directory and all parents
	if configFilepath != "" {
		filename, err := filepath.Abs(configFilepath)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return Config{}, err
		}

		// Check first cwd and all parent directories, then if not found,
		// check if env variable MANI_CONFIG is set, and if not found
		// return no config found
		filename, err := core.FindFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
		if err != nil {
			val, present := os.LookupEnv("MANI_CONFIG")
			if present {
				filename = val
			} else {
				return Config{}, err
			}
		}

		filename, err = absolutePath(filename)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	}

	dat, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	// Found config, now try to read it
	var config Config

	config.Path = configPath
	config.Dir = filepath.Dir(configPath)
	config.UserConfigFile = userConfigFile
	config.Color = color

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		re := ResourceErrors[Config]{Resource: &config, Errors: decodeErrors(err)}
		return config, FormatErrors(re.Resource, re.Errors)
	}

	// Set default shell command
	if config.Shell == "" {
		config.Shell = DEFAULT_SHELL
	} else {
		config.Shell = core.FormatShell(config.Shell)
	}

	// Set Sync Gitignore
	if config.SyncGitignore == nil {
		config.SyncGitignore = new(true)
	}

	// Set Reload TUI
	if config.ReloadTUI == nil {
		config.ReloadTUI = new(false)
	}

	// Set Sync Remote
	if config.SyncRemotes == nil {
		config.SyncRemotes = new(false)
	}

	// Set Remove Orphan Worktrees
	if config.RemoveOrphanedWorktrees == nil {
		config.RemoveOrphanedWorktrees = new(false)
	}

	configResources, err := config.importConfigs()
	if err != nil {
		return config, err
	}

	config.userConfigs = configResources.userConfigs
	config.TaskList = configResources.Tasks
	config.ProjectList = configResources.Projects

	// Relative paths are relative to the root config, since that's where users run mani from,
	// and they're what --paths matches and the tree shows
	for i := range config.ProjectList {
		config.ProjectList[i].RelPath = relativePath(config.Dir, config.ProjectList[i].Path)
	}

	config.ThemeList = configResources.Themes
	config.SpecList = configResources.Specs
	config.TargetList = configResources.Targets
	config.EnvList = configResources.Envs

	config.CheckConfigNoColor()

	config.ConfigPaths = configResources.configFiles

	// Set default theme if it's not set already
	_, err = config.GetTheme(DEFAULT_THEME.Name)
	if err != nil {
		config.ThemeList = append(config.ThemeList, DEFAULT_THEME)
	}

	// Set default spec if it's not set already
	_, err = config.GetSpec(DEFAULT_SPEC.Name)
	if err != nil {
		config.SpecList = append(config.SpecList, DEFAULT_SPEC)
	}

	// Set default target if it's not set already
	_, err = config.GetTarget(DEFAULT_TARGET.Name)
	if err != nil {
		config.TargetList = append(config.TargetList, DEFAULT_TARGET)
	}

	// Parse all tasks
	var taskErrors []ResourceErrors[Task]
	for i := range configResources.Tasks {
		taskErrors = append(taskErrors, configResources.Tasks[i].ParseTask(config)...)
	}

	var configErr = ""
	for _, taskError := range taskErrors {
		if len(taskError.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(taskError.Resource, taskError.Errors))
		}
	}

	if configErr != "" {
		return config, &core.ConfigErr{Msg: configErr}
	}

	// Project env runs shell commands, so it's evaluated last, once the rest of the config is valid
	for _, envError := range EvaluateProjectEnv(config.ProjectList) {
		configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(envError.Resource, envError.Errors))
	}

	if configErr != "" {
		return config, &core.ConfigErr{Msg: configErr}
	}

	return config, nil
}

// Open mani config in editor
func (c Config) EditConfig() error {
	return openEditor(c.Path, -1)
}

func openEditor(path string, lineNr int) error {
	editor := os.Getenv("EDITOR")
	var args []string

	if lineNr > 0 {
		switch editor {
		case "nvim":
			args = []string{fmt.Sprintf("+%v", lineNr), path}
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
	if err != nil {
		return err
	}

	return nil
}

// Open mani config in editor and optionally go to line matching the task name. A task defined in
// multiple configs opens at the definition its cmd comes from.
func (c Config) EditTask(name string) error {
	if name == "" {
		return openEditor(c.Path, c.Tasks.Line-1)
	}

	task, err := c.GetTask(name)
	if err != nil {
		return err
	}

	return openEditor(task.context, task.contextLine)
}

// Open mani config in editor and optionally go to line matching the project name. A project defined
// in multiple configs opens at the definition its path comes from.
func (c Config) EditProject(name string) error {
	if name == "" {
		return openEditor(c.Path, c.Projects.Line-1)
	}

	project, err := c.GetProject(name)
	if err != nil {
		return err
	}

	return openEditor(project.context, project.contextLine)
}

func InitMani(args []string, initFlags core.InitFlags) ([]Project, error) {
	// Choose to initialize mani in a different directory
	// 1. absolute or
	// 2. relative or
	// 3. working directory
	var configDir string
	if len(args) > 0 && filepath.IsAbs(args[0]) {
		// absolute path
		configDir = args[0]
	} else if len(args) > 0 {
		// relative path
		wd, err := os.Getwd()
		if err != nil {
			return []Project{}, err
		}
		configDir = filepath.Join(wd, args[0])
	} else {
		// working directory
		wd, err := os.Getwd()
		if err != nil {
			return []Project{}, err
		}
		configDir = wd
	}

	err := os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		return []Project{}, err
	}

	configPath := filepath.Join(configDir, "mani.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return []Project{}, &core.AlreadyManiDirectory{Dir: configDir}
	}

	// Check if current directory is a git repository
	gitDir := filepath.Join(configDir, ".git")
	isGitRepo := false
	if _, err := os.Stat(gitDir); err == nil {
		isGitRepo = true
	}

	var projects []Project

	// Only add root directory as project if it IS a git repository
	if isGitRepo {
		url, err := core.GetWdRemoteURL(configDir)
		if err != nil {
			return []Project{}, err
		}
		rootName := filepath.Base(configDir)
		rootPath := "."
		rootURL := url
		rootProject := Project{Name: rootName, Path: rootPath, URL: rootURL}

		// Discover worktrees for root project
		if initFlags.AutoDiscovery {
			worktrees, _ := core.GetWorktreeList(configDir)
			for wtPath, branch := range worktrees {
				if branch == "" {
					continue
				}
				wtRelPath, _ := filepath.Rel(configDir, wtPath)
				rootProject.WorktreeList = append(rootProject.WorktreeList, Worktree{
					Path:   wtRelPath,
					Branch: branch,
				})
			}
		}

		projects = []Project{rootProject}
	}

	if initFlags.AutoDiscovery {
		prs, err := FindVCSystems(configDir)
		if err != nil {
			return []Project{}, err
		}
		RenameDuplicates(prs)

		projects = append(projects, prs...)
	}

	funcMap := template.FuncMap{
		"projectItem": func(name string, path string, url string, worktrees []Worktree) string {
			var txt = name + ":"

			if name != path {
				txt = txt + "\n    path: " + path
			}

			if url != "" {
				txt = txt + "\n    url: " + url
			}

			if len(worktrees) > 0 {
				txt = txt + "\n    worktrees:"
				for _, wt := range worktrees {
					txt = txt + "\n      - path: " + wt.Path
					txt = txt + "\n        branch: " + wt.Branch
				}
			}

			return txt
		},
	}

	tmpl, err := template.New("init").Funcs(funcMap).Parse(`projects:
  {{- range .}}
  {{ (projectItem .Name .Path .URL .WorktreeList) }}
  {{ end }}
tasks:
  hello:
    desc: Print Hello World
    cmd: echo "Hello World"
`,
	)
	if err != nil {
		return []Project{}, err
	}

	// Create mani.yaml
	f, err := os.Create(configPath)
	if err != nil {
		return []Project{}, err
	}

	err = tmpl.Execute(f, projects)
	if err != nil {
		return []Project{}, err
	}

	err = f.Close()
	if err != nil {
		return []Project{}, err
	}

	// Update gitignore file only if inside a git repository
	hasURL := false
	for _, project := range projects {
		if project.URL != "" {
			hasURL = true
			break
		}
	}

	if isGitRepo && hasURL && initFlags.SyncGitignore {
		// Add gitignore file
		gitignoreFilepath := filepath.Join(configDir, ".gitignore")
		if _, err := os.Stat(gitignoreFilepath); os.IsNotExist(err) {
			err := os.WriteFile(gitignoreFilepath, []byte(""), 0644)
			if err != nil {
				return []Project{}, err
			}
		}

		var projectNames []string
		for _, project := range projects {
			if project.URL == "" {
				continue
			}

			if project.Path == "." {
				continue
			}

			projectNames = append(projectNames, project.Path)
		}

		// Add projects to gitignore file
		err = UpdateProjectsToGitignore(projectNames, gitignoreFilepath)
		if err != nil {
			return []Project{}, err
		}
	}

	fmt.Println("\nInitialized mani repository in", configDir)
	fmt.Println("- Created mani.yaml")

	if isGitRepo && hasURL && initFlags.SyncGitignore {
		fmt.Println("- Created .gitignore")
	}

	return projects, nil
}

func RenameDuplicates(projects []Project) {
	projectNamesCount := make(map[string]int)
	// Find duplicate names
	for _, p := range projects {
		projectNamesCount[p.Name] += 1
	}

	// Rename duplicate projects
	for i, p := range projects {
		if projectNamesCount[p.Name] > 1 {
			projects[i].Name = p.Path
		}
	}
}

func CheckUserColor(colorFlag bool) bool {
	_, present := os.LookupEnv("NO_COLOR")
	if present || !colorFlag {
		color.Disable()
		return false
	}

	return true
}

func (c *Config) CheckConfigNoColor() {
	for _, env := range c.EnvList {
		name, _, _ := strings.Cut(env, "=")
		if name == "NO_COLOR" {
			color.Disable()
		}
	}
}
