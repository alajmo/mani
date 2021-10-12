package dao

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
	"sync"

	"github.com/theckman/yacspin"
	"gopkg.in/yaml.v3"
	"github.com/jedib0t/go-pretty/v6/table"

	core "github.com/alajmo/mani/core"
	render "github.com/alajmo/mani/core/render"
)

var (
	build_mode = "dev"
)

type CommandInterface interface {
	RunRemoteCmd() (string, error)
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
	User		string `yaml:"user"`
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

	Hosts    []string

	Tags []string
}

type Task struct {
	Theme		yaml.Node `yaml:"theme"`

	Target		Target
	ThemeData	Theme

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
	} else if (t.Theme.Value != "") {
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

func (c Config) ParseTask(task *Task, runFlags core.RunFlags) ([]Entity, []Entity) {
	// OUTPUT
	// var output = runFlags.Output
	// if task.Output != "" && runFlags.Output == "" {
	// 	runFlags.Output = task.Output
	// }

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

func (t *Task) RunTask(
	entityList EntityList,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	t.SetEnvList(userArgs, []string{}, config.GetEnv())

	// Set env for sub-commands
	for i := range t.Commands {
		t.Commands[i].SetEnvList(userArgs, t.EnvList, config.GetEnv())
	}

	spinner, err := TaskSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	core.CheckIfError(err)

	var data core.TableOutput

	/**
	** Column Headers
	**/

	// Headers
	data.Headers = append(data.Headers, entityList.Type)

	// Append Command name if set
	if t.Command != "" {
		data.Headers = append(data.Headers, t.Name)
	}

	// Append Command names if set
	for _, cmd := range t.Commands {
		if cmd.Task != "" {
			task, err := config.GetTask(cmd.Task)
			core.CheckIfError(err)

			if cmd.Name != "" {
				data.Headers = append(data.Headers, cmd.Name)
			} else {
				data.Headers = append(data.Headers, task.Name)
			}
		} else {
			data.Headers = append(data.Headers, cmd.Name)
		}
	}

	for _, entity := range  entityList.Entities {
		data.Rows = append(data.Rows, table.Row{entity.Name})
	}

	/**
	** Table Rows
	**/

	var wg sync.WaitGroup

	for i, entity := range entityList.Entities {
		wg.Add(1)

		if runFlags.Serial {
			spinner.Message(fmt.Sprintf(" %v", entity.Name))
			t.work(config, &data, entity, runFlags.DryRun, i, &wg)
		} else {
			spinner.Message(" Running")
			go t.work(config, &data, entity, runFlags.DryRun, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	/**
	** Print output
	**/
	render.Render(runFlags.Output, data)
}

func (t Task) work(
	config *Config,
	data *core.TableOutput,
	entity Entity,
	dryRunFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if t.Command != "" {
		var output string
		var err error
		output, err = t.RunCmd(*config, t.Shell, entity, dryRunFlag)

		if err != nil {
			data.Rows[i] = append(data.Rows[i], err)
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}

	for _, cmd := range t.Commands {
		var output string
		var err error
		output, err = cmd.RunCmd(*config, cmd.Shell, entity, dryRunFlag)

		if err != nil {
			data.Rows[i] = append(data.Rows[i], output)
			return
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}
}

func formatShellString(shell string, command string) (string, []string) {
	shellProgram := strings.SplitN(shell, " ", 2)
	return shellProgram[0], append(shellProgram[1:], command)
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

func (c CommandBase) RunCmd(
	config Config,
	shell string,
	entity Entity,
	dryRun bool,
) (string, error) {
	entityPath, err := core.GetAbsolutePath(config.Path, entity.Path, entity.Name)
	if err != nil {
		return "", &core.FailedToParsePath{Name: entityPath}
	}
	if _, err := os.Stat(entityPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{Path: entityPath}
	}

	defaultArguments := getDefaultArguments(config.Path, config.Dir, entity)

	var shellProgram string
	var commandStr []string

	if c.Task != "" {
		refTask, err := config.GetTask(c.Task)
		if err != nil {
			return "", err
		}

		shellProgram, commandStr = formatShellString(refTask.Shell, refTask.Command)
	} else {
		shellProgram, commandStr = formatShellString(shell, c.Command)
	}

	// Execute Command
	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = entityPath

	var output string
	if dryRun {
		for _, arg := range defaultArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		for _, arg := range c.EnvList {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		output = os.ExpandEnv(c.Command)
	} else {
		cmd.Env = append(os.Environ(), defaultArguments...)
		cmd.Env = append(cmd.Env, c.EnvList...)

		var outb bytes.Buffer
		var errb bytes.Buffer

		cmd.Stdout = &outb
		cmd.Stderr = &errb

		err := cmd.Run()
		if err != nil {
			output = errb.String()
		} else {
			output = outb.String()
		}

		return output, err
	}

	return output, nil
}

func ExecCmd(
	configPath string,
	shell string,
	project Project,
	cmdString string,
	dryRun bool,
) (string, error) {
	projectPath, err := core.GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		return "", &core.FailedToParsePath{Name: projectPath}
	}
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return "", &core.PathDoesNotExist{Path: projectPath}
	}
	// TODO: FIX THIS
	// defaultArguments := getDefaultArguments(config.Path, config.Dir, entity)

	// Execute Command
	shellProgram, commandStr := formatShellString(shell, cmdString)

	cmd := exec.Command(shellProgram, commandStr...)
	cmd.Dir = projectPath

	var output string
	if dryRun {
		// for _, arg := range defaultArguments {
		// 	env := strings.SplitN(arg, "=", 2)
		// 	os.Setenv(env[0], env[1])
		// }

		output = os.ExpandEnv(cmdString)
	} else {
		// cmd.Env = append(os.Environ(), defaultArguments...)
		out, _ := cmd.CombinedOutput()
		output = string(out)
	}

	return output, nil
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
