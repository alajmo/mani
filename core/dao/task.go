package dao

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/theckman/yacspin"
	"gopkg.in/yaml.v3"

	core "github.com/alajmo/mani/core"
)

var (
	build_mode = "dev"
)

type Command struct {
	Name    string    `yaml:"name"`
	Desc    string    `yaml:"desc"`
	Env     yaml.Node `yaml:"env"`
	EnvList []string
	Shell   string `yaml:"shell"`
	Command string `yaml:"command"`
	Task    string `yaml:"task"`
}

type Target struct {
	Projects     []string `yaml:"projects"`
	Dirs     []string
	Paths []string `yaml:"Paths"`
	Tags []string
	Cwd bool
}

type Task struct {
	Context string
	Theme   yaml.Node `yaml:"theme"`
	Output  string

	Target    Target
	Parallel  bool
	Abort     bool
	ThemeData Theme

	Name     string    `yaml:"name"`
	Desc     string    `yaml:"desc"`
	Env      yaml.Node `yaml:"env"`
	EnvList  []string
	Shell    string `yaml:"shell"`
	Command  string `yaml:"command"`
	Commands []Command
}

func (t *Task) ParseTheme(config Config) {
	if len(t.Theme.Content) > 0 {
		// Theme value
		theme := &Theme{}
		t.Theme.Decode(theme)

		t.ThemeData = *theme
	} else if t.Theme.Value != "" {
		// Theme reference
		theme, err := config.GetTheme(t.Theme.Value)
		core.CheckIfError(err)

		t.ThemeData = *theme
	} else {
		// Default theme
		theme, err := config.GetTheme(DEFAULT_THEME.Name)
		core.CheckIfError(err)

		t.ThemeData = *theme
	}
}

func (t *Task) ParseTask(config Config) {
	if t.Shell == "" {
		t.Shell = DEFAULT_SHELL
	}

	for j, cmd := range t.Commands {
		if cmd.Task != "" {
			cmdRef, err := config.GetCommand(cmd.Task)
			core.CheckIfError(err)

			t.Commands[j] = *cmdRef
		}

		if t.Commands[j].Shell == "" {
			t.Commands[j].Shell = DEFAULT_SHELL
		}
	}

	if len(t.Theme.Content) > 0 {
		// Theme value
		theme := &Theme{}
		t.Theme.Decode(theme)

		t.ThemeData = *theme
	} else if t.Theme.Value != "" {
		// Theme reference
		theme, err := config.GetTheme(t.Theme.Value)
		core.CheckIfError(err)

		t.ThemeData = *theme
	} else {
		// Default theme
		theme, err := config.GetTheme(DEFAULT_THEME.Name)
		core.CheckIfError(err)

		t.ThemeData = *theme
	}

	if t.Output == "" {
		t.Output = "table"
	}
}

func (t *Task) ParseOutput(config Config) {
}

func formatShellString(shell string, command string) (string, []string) {
	shellProgram := strings.SplitN(shell, " ", 2)
	return shellProgram[0], append(shellProgram[1:], command)
}

func TaskSpinner() (yacspin.Spinner, error) {
	var cfg yacspin.Config

	// NOTE: Don't print the spinner in tests since it causes
	// golden files to produce different results.
	if build_mode == "TEST" {
		cfg = yacspin.Config{
			Frequency:       100 * time.Millisecond,
			CharSet:         yacspin.CharSets[9],
			SuffixAutoColon: false,
			Writer:          io.Discard,
		}
	} else {
		cfg = yacspin.Config{
			Frequency:       100 * time.Millisecond,
			CharSet:         yacspin.CharSets[9],
			SuffixAutoColon: false,
		}
	}

	spinner, err := yacspin.New(cfg)

	return *spinner, err
}

func (c Command) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return c.Name
	case "Desc", "desc":
		return c.Desc
	case "Command", "command":
		return c.Command
	}

	return ""
}

func (t Task) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return t.Name
	case "Desc", "desc", "Description", "description":
		return t.Desc
	case "Command", "command":
		return t.Command
	}

	return ""
}

func (c *Config) GetTaskList() []Task {
	var tasks []Task
	count := len(c.Tasks.Content)

	for i := 0; i < count; i += 2 {
		task := &Task{}

		if c.Tasks.Content[i+1].Kind == 8 {
			task.Command = c.Tasks.Content[i+1].Value
		} else {
			c.Tasks.Content[i+1].Decode(task)
		}

		// Add context to each task
		task.Name = c.Tasks.Content[i].Value
		task.Context = c.Path

		tasks = append(tasks, *task)
	}

	return tasks
}

func GetEnvList(env yaml.Node, userEnv []string, parentEnv []string, configEnv []string) []string {
	pEnv, err := core.EvaluateEnv(parentEnv)
	core.CheckIfError(err)

	cmdEnv, err := core.EvaluateEnv(GetEnv(env))
	core.CheckIfError(err)

	globalEnv, err := core.EvaluateEnv(configEnv)
	core.CheckIfError(err)

	envList := core.MergeEnv(userEnv, cmdEnv, pEnv, globalEnv)

	return envList
}

func (c Config) GetTaskEntities(task *Task, runFlags core.RunFlags) ([]Entity, []Entity) {
	// TAGS
	var tags = runFlags.Tags
	if len(tags) == 0 {
		tags = task.Target.Tags
	}

	// CWD
	cwd := runFlags.Cwd
	if task.Target.Cwd == true && cwd == false {
		cwd = true
	} else if task.Target.Cwd == true && cwd == true {
		cwd = true
	} else if task.Target.Cwd == false && cwd == true {
		cwd = true
	} else if task.Target.Cwd == false && cwd == false {
		cwd = false
	}

	var projects []Project
	// If any runtime target flags are used, disregard task targets
	if len(runFlags.Projects) > 0 || len(runFlags.Paths) > 0 || len(runFlags.Tags) > 0 || runFlags.Cwd == true || runFlags.AllProjects == true {
		projects = c.FilterProjects(runFlags.Cwd, runFlags.AllProjects, runFlags.Paths, runFlags.Projects, runFlags.Tags)
	} else {
		// PROJECTS
		var projectNames = runFlags.Projects
		if len(projectNames) == 0 {
			projectNames = task.Target.Projects
		}

		var paths = runFlags.Paths
		if len(runFlags.Paths) == 0 {
			paths = task.Target.Paths
		}

		projects = c.FilterProjects(cwd, runFlags.AllProjects, paths, projectNames, tags)
	}

	var projectEntities []Entity
	for i := range projects {
		var entity Entity
		entity.Name = projects[i].Name
		entity.Path = projects[i].Path
		entity.Type = "project"

		projectEntities = append(projectEntities, entity)
	}

	var dirs []Dir
	// If any runtime target flags are used, disregard task targets
	if len(runFlags.Dirs) > 0 || len(runFlags.Paths) > 0 || len(runFlags.Tags) > 0 || runFlags.Cwd == true || runFlags.AllDirs == true {
		dirs = c.FilterDirs(runFlags.Cwd, runFlags.AllDirs, runFlags.Paths, runFlags.Dirs, runFlags.Tags)
	} else {
		// DIRS
		var dirNames = runFlags.Dirs
		if len(dirNames) == 0 {
			dirNames = task.Target.Dirs
		}

		var paths = runFlags.Paths
		if len(paths) == 0 {
			paths = task.Target.Paths
		}

		dirs = c.FilterDirs(cwd, runFlags.AllDirs, paths, dirNames, tags)
	}

	var dirEntities []Entity
	for i := range dirs {
		var entity Entity
		entity.Name = dirs[i].Name
		entity.Path = dirs[i].Path
		entity.Type = "directory"

		dirEntities = append(dirEntities, entity)
	}

	return projectEntities, dirEntities
}

func (c Config) GetEntities(runFlags core.RunFlags) ([]Entity, []Entity) {
	projects := c.FilterProjects(runFlags.Cwd, runFlags.AllProjects, runFlags.Paths, runFlags.Projects, runFlags.Tags)
	var projectEntities []Entity
	for i := range projects {
		var entity Entity
		entity.Name = projects[i].Name
		entity.Path = projects[i].Path
		entity.Type = "project"

		projectEntities = append(projectEntities, entity)
	}

	dirs := c.FilterDirs(runFlags.Cwd, runFlags.AllDirs, runFlags.Paths, runFlags.Dirs, runFlags.Tags)
	var dirEntities []Entity
	for i := range dirs {
		var entity Entity
		entity.Name = dirs[i].Name
		entity.Path = dirs[i].Path
		entity.Type = "directory"

		dirEntities = append(dirEntities, entity)
	}

	return projectEntities, dirEntities
}

func getDefaultArguments(configPath string, configDir string, entity Entity) []string {
	// Default arguments
	maniConfigPath := fmt.Sprintf("MANI_CONFIG_PATH=%s", configPath)
	maniConfigDir := fmt.Sprintf("MANI_CONFIG_DIR=%s", configDir)
	projectNameEnv := fmt.Sprintf("MANI_PROJECT_NAME=%s", entity.Name)
	projectPathEnv := fmt.Sprintf("MANI_PROJECT_PATH=%s", entity.Path)

	defaultArguments := []string{maniConfigPath, maniConfigDir, projectNameEnv, projectPathEnv}

	return defaultArguments
}

// TODO: Not used, remove
func GetEnv(node yaml.Node) []string {
	var envs []string
	count := len(node.Content)

	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", node.Content[i].Value, node.Content[i+1].Value)
		envs = append(envs, env)
	}

	return envs
}

func (c Config) GetTasksByNames(names []string) []Task {
	if len(names) == 0 {
		return c.TaskList
	}

	var filteredTasks []Task
	var foundTasks []string
	for _, name := range names {
		if core.StringInSlice(name, foundTasks) {
			continue
		}

		for _, task := range c.TaskList {
			if name == task.Name {
				filteredTasks = append(filteredTasks, task)
				foundTasks = append(foundTasks, name)
			}
		}
	}

	return filteredTasks
}

func (c Config) GetTaskNames() []string {
	taskNames := []string{}
	for _, task := range c.TaskList {
		taskNames = append(taskNames, task.Name)
	}

	return taskNames
}

func (c Config) GetTask(task string) (*Task, error) {
	for _, cmd := range c.TaskList {
		if task == cmd.Name {
			return &cmd, nil
		}
	}

	return nil, &core.TaskNotFound{Name: task}
}

func (c Config) GetCommand(task string) (*Command, error) {
	for _, cmd := range c.TaskList {
		if task == cmd.Name {
			cmdRef := &Command{
				Name:    cmd.Name,
				Desc:    cmd.Desc,
				EnvList: cmd.EnvList,
				Shell:   cmd.Shell,
				Command: cmd.Command,
			}

			return cmdRef, nil
		}
	}

	return nil, &core.TaskNotFound{Name: task}
}
