package dao

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

var (
	Version               = "dev"
	DEFAULT_SHELL         = "sh -c"
	ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani", ".mani.yaml", ".mani.yml", "Manifile", "Manifile.yaml", "Manifile.yml"}
)

type Config struct {
	Path string

	Import      []string `yaml:"import"`
	EnvList     []string
	NetworkList []Network
	ThemeList   []Theme
	Projects    []Project `yaml:"projects"`
	Dirs        []Dir     `yaml:"dirs"`
	Tasks       []Task    `yaml:"tasks"`

	Env	    yaml.Node `yaml:"env"`
	Networks    yaml.Node `yaml:"networks"`
	Themes	    yaml.Node `yaml:"themes"`
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
	config.Path = configPath

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		parseError := &core.FailedToParseFile{Name: configPath, Msg: err}
		return config, parseError
	}

	// Update the config
	// TODO: Handle default theme (add default theme which all tasks inherit, unless specified)

	// if config.Theme.Table == "" {
	// 	config.Theme.Table = "box"
	// }
	// if config.Theme.Tree == "" {
	// 	config.Theme.Tree = "line"
	// }

	// Append absolute and relative path for each project
	for i := range config.Projects {
		config.Projects[i].Path, err = core.GetAbsolutePath(configPath, config.Projects[i].Path, config.Projects[i].Name)
		core.CheckIfError(err)

		config.Projects[i].RelPath, err = GetProjectRelPath(configPath, config.Projects[i].Path)
		core.CheckIfError(err)
	}

	// Append absolute and relative path for each dir
	for i := range config.Dirs {
		var abs, err = core.GetAbsolutePath(configPath, config.Dirs[i].Path, "")
		core.CheckIfError(err)

		config.Dirs[i].Name = path.Base(abs)
		config.Dirs[i].Path = abs

		config.Dirs[i].RelPath, err = GetProjectRelPath(configPath, config.Dirs[i].Path)
		core.CheckIfError(err)
	}

	// Import Tasks/Projects/Networks
	tasks := config.Tasks
	projects := config.Projects
	networks := config.SetNetworkList()
	themes := config.SetThemeList()
	for _, importPath := range config.Import {
		ts, thms, ps, ns, err := readExternalConfig(importPath)
		core.CheckIfError(err)

		tasks = append(tasks, ts...)
		projects = append(projects, ps...)
		networks = append(networks, ns...)
		themes = append(themes, thms...)
	}

	// Parse and update tasks
	for i := range tasks {
	    tasks[i].ParseTheme(config)
	}

	config.Projects = projects
	config.NetworkList = networks
	config.ThemeList = themes
	config.Tasks = tasks

	return config, nil
}

func readExternalConfig(importPath string) ([]Task, []Theme, []Project, []Network, error) {
	dat, err := ioutil.ReadFile(importPath)
	core.CheckIfError(err)

	// Found config, now try to read it
	var config Config
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		parseError := &core.FailedToParseFile{Name: importPath, Msg: err}
		core.CheckIfError(parseError)
	}

	// Append absolute and relative path for each project
	for i := range config.Projects {
		config.Projects[i].Path, err = core.GetAbsolutePath(importPath, config.Projects[i].Path, config.Projects[i].Name)
		core.CheckIfError(err)

		config.Projects[i].RelPath, err = GetProjectRelPath(importPath, config.Projects[i].Path)
		core.CheckIfError(err)
	}

	// Unpack Network to NetworkList
	networks := config.SetNetworkList()

	// Unpack Theme to ThemeList
	themes := config.SetThemeList()

	return config.Tasks, themes, config.Projects, networks, nil
}

// Open mani config in editor
func (c Config) EditConfig() {
	openEditor(c.Path, -1)
}

// Open mani config in editor and optionally go to line matching the task name
func (c Config) EditTask(taskName string) {
	dat, err := ioutil.ReadFile(c.Path)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Tasks yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if taskName == "" {
		lineNr = configTmp.Tasks.Line - 1
	} else {
	out:
		for _, task := range configTmp.Tasks.Content {
			for _, node := range task.Content {
				if node.Value == taskName {
					lineNr = node.Line
					break out
				}
			}
		}
	}

	openEditor(c.Path, lineNr)
}

// Open mani config in editor and optionally go to line matching the project name
func (c Config) EditProject(projectName string) {
	dat, err := ioutil.ReadFile(c.Path)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Projects yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if projectName == "" {
		lineNr = configTmp.Projects.Line - 1
	} else {
	out:
		for _, project := range configTmp.Projects.Content {
			for _, node := range project.Content {
				if node.Value == projectName {
					lineNr = node.Line
					break out
				}
			}
		}
	}

	openEditor(c.Path, lineNr)
}

// Open mani config in editor and optionally go to line matching the dir name
func (c Config) EditDir(name string) {
	dat, err := ioutil.ReadFile(c.Path)
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

	openEditor(c.Path, lineNr)
}

// Open mani config in editor and optionally go to line matching the network name
func (c Config) EditNetworks(networkName string) {
	dat, err := ioutil.ReadFile(c.Path)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Networks yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if networkName == "" {
		lineNr = configTmp.Networks.Line - 1
	} else {
	out:
		for _, network := range configTmp.Networks.Content {
			if network.Value == networkName {
				lineNr = network.Line
				break out
			}
		}
	}

	openEditor(c.Path, lineNr)
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
