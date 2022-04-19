package dao

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"errors"

	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

var (
	Version               = "dev"
	DEFAULT_SHELL         = "bash -c"
	ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani", ".mani.yaml", ".mani.yml", "Manifile", "Manifile.yaml", "Manifile.yml"}

	DEFAULT_THEME = Theme {
		Name:  "default",
		Table: DefaultTable,
		Text: DefaultText,
		Tree:  DefaultTree,
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
	EnvList     []string
	ImportData  []Import
	ThemeList   []Theme
	SpecList    []Spec
	TargetList  []Target
	ProjectList []Project
	TaskList    []Task
	Shell       string `yaml:"shell"`

	// Intermediate
	Env      yaml.Node `yaml:"env"`
	Import   yaml.Node `yaml:"import"`
	Themes   yaml.Node `yaml:"themes"`
	Specs    yaml.Node `yaml:"specs"`
	Targets  yaml.Node `yaml:"targets"`
	Projects yaml.Node `yaml:"projects"`
	Tasks    yaml.Node `yaml:"tasks"`

	// Internal
	Path string
	Dir  string
	UserConfigFile *string
}

func (c *Config) GetContext() string {
	return c.Path
}

func (c *Config) GetContextLine() int {
	return -1
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

func getUserConfigFile(userConfigDir string) *string {
	userConfigFile := filepath.Join(userConfigDir, "config.yaml")

	if _, err := os.Stat(userConfigFile); err == nil {
		return &userConfigFile
	}

	userConfigFile = filepath.Join(userConfigDir, "config.yml")
	if _, err := os.Stat(userConfigFile); err == nil {
		return &userConfigFile
	}

	return nil
}

// Function to read Mani configs.
func ReadConfig(configFilepath string, userConfigDir string, noColor bool) (Config, error) {
	var configPath string

	userConfigFile := getUserConfigFile(userConfigDir)

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
		re := ResourceErrors[Config]{ Resource: &config, Errors: []error{err} }
		return config, FormatErrors(re.Resource, re.Errors)
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

	MaybeDisableColor(noColor)

	// Parse all tasks
	taskErrors := make([]ResourceErrors[Task], len(configResources.Tasks))
	for i := range configResources.Tasks {
		taskErrors[i].Resource = &configResources.Tasks[i]
		configResources.Tasks[i].ParseTask(config, &taskErrors[i])
	}

	var configErr = ""
	for _, taskError := range taskErrors {
		if len(taskError.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(taskError.Resource, taskError.Errors))
		}
	}

	if configErr != "" {
		return config, errors.New(configErr)
	}

	return config, nil
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
		configPath = task.context
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
		configPath = project.context
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

func InitMani(args []string, initFlags core.InitFlags) (string, []Project) {
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
		RenameDuplicates(prs)

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

	// Update gitignore file if vcs set to git
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

	fmt.Println("\nInitialized mani repository in", configDir)
	fmt.Println("- Created mani.yaml")

	if hasUrl && initFlags.Vcs == "git" {
		fmt.Println("- Created .gitignore")
	}

	return configDir, projects
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

func MaybeDisableColor(noColorFlag bool) {
	_, present := os.LookupEnv("NO_COLOR")
	if noColorFlag || present  {
		text.DisableColors()
	}
}
