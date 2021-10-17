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
    DEFAULT_THEME         = Theme{
	Name:   "default",
	Table:  "ascii",
	Tree:   "line",
    }
)

type Config struct {
    Path string
    Dir  string

    Import    []string `yaml:"import"`
    EnvList   []string
    ThemeList []Theme
    Shell     string    `yaml:"shell"`
    Projects  []Project `yaml:"projects"`
    Dirs      []Dir     `yaml:"dirs"`
    Tasks     []Task    `yaml:"tasks"`

    Env    yaml.Node `yaml:"env"`
    Themes yaml.Node `yaml:"themes"`
}

func (c Config) GetEnv() []string {
    var envs []string
    count := len(c.Env.Content)
    for i := 0; i < count; i += 2 {
	env := fmt.Sprintf("%v=%v", c.Env.Content[i].Value, c.Env.Content[i+1].Value)
	envs = append(envs, env)
    }

    return envs
}

func ReadConfig(cfgName string) (Config, error) {
    var configPath string

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

    err = yaml.Unmarshal(dat, &config)
    if err != nil {
	parseError := &core.FailedToParseFile{Name: configPath, Msg: err}
	return config, parseError
    }

    config.Path = configPath
    config.Dir = filepath.Dir(configPath)
    config.ParseConfig()

    // Set default shell command
    if config.Shell == "" {
	config.Shell = DEFAULT_SHELL
    }

    // Import Tasks/Projects
    tasks := config.Tasks
    projects := config.Projects
    dirs := config.Dirs
    themes := config.SetThemeList()
    for _, importPath := range config.Import {
	ts, ps, ds, thms, err := importConfig(importPath)
	core.CheckIfError(err)

	tasks = append(tasks, ts...)
	projects = append(projects, ps...)
	dirs = append(dirs, ds...)
	themes = append(themes, thms...)
    }

    // Parse all tasks
    for i := range tasks {
	tasks[i].ParseTasks(config)
	tasks[i].ParseTheme(config)
	tasks[i].ParseOutput(config)
    }

    config.Tasks = tasks
    config.Projects = projects
    config.Dirs = dirs
    config.ThemeList = themes

    return config, nil
}

func importConfig(importPath string) ([]Task, []Project, []Dir, []Theme, error) {
    dat, err := ioutil.ReadFile(importPath)
    core.CheckIfError(err)

    absPath, err := filepath.Abs(importPath)
    core.CheckIfError(err)

    // Found config, now try to read it
    var config Config
    err = yaml.Unmarshal(dat, &config)
    if err != nil {
	parseError := &core.FailedToParseFile{Name: importPath, Msg: err}
	core.CheckIfError(parseError)
    }

    config.Dir = filepath.Dir(absPath)
    config.ParseConfig()

    // Unpack Theme to ThemeList
    themes := config.SetThemeList()

    return config.Tasks, config.Projects, config.Dirs, themes, nil
}

// Open mani config in editor
func (c Config) EditConfig() {
    openEditor(c.Path, -1)
}

func (c Config) ParseConfig() {
    // Add absolute and relative path for each project
    var err error
    for i := range c.Projects {
	c.Projects[i].Path, err = core.GetAbsolutePath(c.Dir, c.Projects[i].Path, c.Projects[i].Name)
	core.CheckIfError(err)

	c.Projects[i].RelPath, err = core.GetRelativePath(c.Dir, c.Projects[i].Path)
	core.CheckIfError(err)

	c.Projects[i].Context = c.Path
    }

    // Add absolute and relative path for each dir
    for i := range c.Dirs {
	c.Dirs[i].Path, err = core.GetAbsolutePath(c.Dir, c.Dirs[i].Path, c.Dirs[i].Name)
	core.CheckIfError(err)

	c.Dirs[i].RelPath, err = core.GetRelativePath(c.Dir, c.Dirs[i].Path)
	core.CheckIfError(err)

	c.Dirs[i].Context = c.Path
    }

    // Add context to ech task
    for i := range c.Tasks {
	c.Tasks[i].Context = c.Path
    }
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
	out:
	for _, task := range configTmp.Tasks.Content {
	    for _, node := range task.Content {
		if node.Value == name {
		    lineNr = node.Line
		    break out
		}
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
	out:
	for _, project := range configTmp.Projects.Content {
	    for _, node := range project.Content {
		if node.Value == name {
		    lineNr = node.Line
		    break out
		}
	    }
	}
    }

    openEditor(configPath, lineNr)
}

// Open mani config in editor and optionally go to line matching the dir name
func (c Config) EditDir(name string) {
    configPath := c.Path
    if name != "" {
	dir, err := c.GetDir(name)
	core.CheckIfError(err)
	configPath = dir.Context
    }

    dat, err := ioutil.ReadFile(configPath)
    core.CheckIfError(err)

    type ConfigTmp struct {
	Dirs yaml.Node
    }

    var configTmp ConfigTmp
    err = yaml.Unmarshal([]byte(dat), &configTmp)
    core.CheckIfError(err)

    lineNr := 0
    if name == "" {
	lineNr = configTmp.Dirs.Line - 1
    } else {
	out:
	for _, dir := range configTmp.Dirs.Content {
	    for _, node := range dir.Content {
		if node.Value == name {
		    lineNr = node.Line
		    break out
		}
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
	    var txt = "- name: " + name

	    if name != path {
		txt = txt + "\n    path: " + path
	    }

	    if url != "" {
		txt = txt + "\n    url: " + url
	    }

	    return txt
	},
    }

    // - name: {{ .Name }}
    // {{ if ne .Name .Path }}path: {{ .Path }}{{ end }}
    // {{ if .Url }}url: {{ .Url }} {{ end }}

    // Path, Name, Url
    tmpl, err := template.New("init").Funcs(funcMap).Parse(`projects:
    {{- range .}}
    {{ (projectItem .Name .Path .Url) }}
    {{ end }}
    tasks:
    - name: hello-world
    description: Print Hello World
    command: echo "Hello World"
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

if hasUrl {
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

func (c Config) SyncDirs(configDir string, parallellFlag bool) {
    for _, dir := range c.Dirs {
	if _, err := os.Stat(dir.Path); os.IsNotExist(err) {
	    os.MkdirAll(dir.Path, os.ModePerm)
	}
    }
}

func (c Config) SyncProjects(configDir string, parallellFlag bool) {
    // Get relative project names for gitignore file
    var projectNames []string
    for _, project := range c.Projects {
	if project.Url == "" {
	    continue
	}

	if project.Path == "." {
	    continue
	}

	// Project must be below mani config file
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

    if len(projectNames) > 0 {
	gitignoreFilename := filepath.Join(filepath.Dir(c.Path), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); os.IsNotExist(err) {
	    err := ioutil.WriteFile(gitignoreFilename, []byte(""), 0644)
	    core.CheckIfError(err)
	}

	err := UpdateProjectsToGitignore(projectNames, gitignoreFilename)
	if err != nil {
	    fmt.Println(err)
	    return
	}

	c.CloneRepos(parallellFlag)
    }
}
