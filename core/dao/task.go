package dao

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/theckman/yacspin"
	"gopkg.in/yaml.v3"

	core "github.com/alajmo/mani/core"
)

var (
	build_mode = "dev"
)

type CommandInterface interface {
	RunCmd() (string, error)
	ExecCmd() (string, error)
	GetEnv() []string
	SetEnvList() []string
	GetValue(string) string
}

type CommandBase struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Env         yaml.Node `yaml:"env"`
	EnvList     []string
	Shell       string `yaml:"shell"`
	User        string `yaml:"user"`
	Command     string `yaml:"command"`
	Task        string `yaml:"task"`
}

type Command struct {
	CommandBase `yaml:",inline"`
}

type Target struct {
	Projects     []string `yaml:"projects"`
	ProjectPaths []string `yaml:"projectPaths"`

	Dirs     []string
	DirPaths []string

	Hosts []string

	Tags []string
}

type Task struct {
	Theme		yaml.Node `yaml:"theme"`
	Output		string

	Target		Target
	ThemeData	Theme

	Serial		bool
	Abort       bool
	Commands    []Command
	CommandBase `yaml:",inline"`
}

func (t *Task) ParseTheme(config Config) {
	if len(t.Theme.Content) > 0 {
		// Theme Value
		theme := &Theme{}
		t.Theme.Decode(theme)

		t.ThemeData = *theme
	} else if t.Theme.Value != "" {
		// Theme Reference
		theme, err := config.GetTheme(t.Theme.Value)
		core.CheckIfError(err)

		t.ThemeData = *theme
	} else {
		theme, err := config.GetTheme(DEFAULT_THEME.Name)
		core.CheckIfError(err)

		t.ThemeData = *theme
	}
}

func (t *Task) ParseShell(config Config) {
	if t.Shell == "" {
		t.Shell = DEFAULT_SHELL
	}

	for j := range t.Commands {
		if t.Commands[j].Shell == "" {
			t.Commands[j].Shell = DEFAULT_SHELL
		}
	}
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

func (c CommandBase) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return c.Name
	case "Description", "description":
		return c.Description
	case "Command", "command":
		return c.Command
	}

	return ""
}

func (c *CommandBase) SetEnvList(userEnv []string, parentEnv []string, configEnv []string) {
	pEnv, err := core.EvaluateEnv(parentEnv)
	core.CheckIfError(err)

	cmdEnv, err := core.EvaluateEnv(c.GetEnv())
	core.CheckIfError(err)

	globalEnv, err := core.EvaluateEnv(configEnv)
	core.CheckIfError(err)

	envList := core.MergeEnv(userEnv, cmdEnv, pEnv, globalEnv)

	c.EnvList = envList
}

func (c Config) GetEntities(task *Task, runFlags core.RunFlags) ([]Entity, []Entity) {
	// TAGS
	var tags = runFlags.Tags
	if len(tags) == 0 {
		tags = task.Target.Tags
	}

	// PROJECTS
	var projectNames = runFlags.Projects
	if len(projectNames) == 0 {
		projectNames = task.Target.Projects
	}

	var projectPaths = runFlags.ProjectPaths
	if len(runFlags.ProjectPaths) == 0 {
		projectPaths = task.Target.ProjectPaths
	}

	projects := c.FilterProjects(runFlags.Cwd, runFlags.AllProjects, projectPaths, projectNames, tags)
	var projectEntities []Entity
	for i := range projects {
		var entity Entity
		entity.Name = projects[i].Name
		entity.Path = projects[i].Path
		entity.Type = "project"

		projectEntities = append(projectEntities, entity)
	}

	// DIRS
	var dirNames = runFlags.Dirs
	if len(dirNames) == 0 {
		dirNames = task.Target.Dirs
	}

	var dirPaths = runFlags.DirPaths
	if len(dirPaths) == 0 {
		dirPaths = task.Target.DirPaths
	}

	dirs := c.FilterDirs(runFlags.Cwd, runFlags.AllDirs, dirPaths, dirNames, tags)
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

func (c CommandBase) GetEnv() []string {
	var envs []string
	count := len(c.Env.Content)

	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", c.Env.Content[i].Value, c.Env.Content[i+1].Value)
		envs = append(envs, env)
	}

	return envs
}

func (c Config) GetTasksByNames(names []string) []Task {
	if len(names) == 0 {
		return c.Tasks
	}

	var filteredTasks []Task
	var foundTasks []string
	for _, name := range names {
		if core.StringInSlice(name, foundTasks) {
			continue
		}

		for _, task := range c.Tasks {
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
	for _, task := range c.Tasks {
		taskNames = append(taskNames, task.Name)
	}

	return taskNames
}

func (c Config) GetTask(task string) (*Task, error) {
	for _, cmd := range c.Tasks {
		if task == cmd.Name {
			return &cmd, nil
		}
	}

	return nil, &core.TaskNotFound{Name: task}
}
