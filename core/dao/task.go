package dao

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"sync"

	"github.com/alajmo/goph"
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
	GetEnv() []string
	SetEnvList() []string
	GetValue(string) string
}

type CommandBase struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Env         yaml.Node `yaml:"env"`
	EnvList     []string
	User		string `yaml:"user"`
	Command     string `yaml:"command"`
	Task         string `yaml:"task"`
}

type Command struct {
	CommandBase `yaml:",inline"`
}

type Target struct {
	Projects     []string `yaml:"projects"`
	ProjectPaths []string `yaml:"projectPaths"`

	Dirs     []string
	DirPaths []string

	Networks []string
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
	} else {
		// Theme Reference
		theme, err := config.GetTheme(t.Theme.Value)
		core.CheckIfError(err)

		t.ThemeData = *theme
	}
}

func (c Config) ParseTask(task *Task, runFlags core.RunFlags) ([]Entity, []Entity, []Entity) {
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


	// NETWORKS
	var networkNames = runFlags.Networks
	if len(networkNames) == 0 {
		networkNames = task.Target.Networks
	}

	var hosts = runFlags.Hosts
	if len(hosts) == 0 {
		hosts = task.Target.Hosts
	}

	networks := c.FilterNetworks(runFlags.AllNetworks, networkNames, hosts, tags)
	var networkEntities []Entity
	for i := range networks {
		for j := range networks[i].Hosts {
			var entity Entity
			entity.Type = "host"
			entity.User = networks[i].User
			entity.Name = networks[i].Name
			entity.Host = networks[i].Hosts[j]

			networkEntities = append(networkEntities, entity)
		}
	}

	return projectEntities, dirEntities, networkEntities
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
	if entityList.Type == "Host" {
		data.Headers = append(data.Headers, "Network", entityList.Type)
	} else {
		data.Headers = append(data.Headers, entityList.Type)
	}

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
		if entity.Type == "host" {
			data.Rows = append(data.Rows, table.Row{entity.Name, entity.Host})
		} else {
			data.Rows = append(data.Rows, table.Row{entity.Name})
		}
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
		if entity.Type == "host" {
			output, err = t.RunRemoteCmd(*config, entity, dryRunFlag)
		} else {
			output, err = t.RunCmd(*config, entity, dryRunFlag)
		}

		if err != nil {
			data.Rows[i] = append(data.Rows[i], err)
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}

	for _, cmd := range t.Commands {
		var output string
		var err error
		if entity.Type == "host" {
			output, err = cmd.RunRemoteCmd(*config, entity, dryRunFlag)
		} else {
			output, err = cmd.RunCmd(*config, entity, dryRunFlag)
		}

		if err != nil {
			data.Rows[i] = append(data.Rows[i], output)
			return
		} else {
			data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
		}
	}
}

func formatCmd(cmdString string) (string, string) {
	parts := strings.SplitN(cmdString, " ", 2)
	return parts[0], strings.Join(parts[1:], "")
}

func getDefaultArguments(configPath string, entity Entity) []string {
	// Default arguments
	maniConfigPath := fmt.Sprintf("MANI_CONFIG_PATH=%s", configPath)
	maniConfigDir := fmt.Sprintf("MANI_CONFIG_DIR=%s", filepath.Dir(configPath))
	projectNameEnv := fmt.Sprintf("MANI_PROJECT_NAME=%s", entity.Name)
	projectPathEnv := fmt.Sprintf("MANI_PROJECT_PATH=%s", entity.Path)

	defaultArguments := []string{maniConfigPath, maniConfigDir, projectNameEnv, projectPathEnv}

	return defaultArguments
}

func (c CommandBase) RunCmd(
	config Config,
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

	defaultArguments := getDefaultArguments(config.Path, entity)

	var shellProgram string
	var commandStr string

	if c.Task != "" {
		refTask, err := config.GetTask(c.Task)
		if err != nil {
			return "", err
		}

		shellProgram, commandStr = formatCmd(refTask.Command)
	} else {
		shellProgram, commandStr = formatCmd(c.Command)
	}

	// Execute Command
	cmd := exec.Command(shellProgram, commandStr)
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
	// defaultArguments := getDefaultArguments(configPath, project)

	// Execute Command
	shellProgram, commandStr := formatCmd(cmdString)
	cmd := exec.Command(shellProgram, commandStr)
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

func (c CommandBase) RunRemoteCmd(
	config Config,
	entity Entity,
	dryRun bool,
) (string, error) {
	// SSH Init
	auth, err := goph.UseAgent()
	core.CheckIfError(err)

	client, err := goph.New(entity.User, entity.Host, auth)
	core.CheckIfError(err)

	defer client.Close()

	defaultArguments := getDefaultArguments(config.Path, entity)

	var shellProgram string
	var commandStr string

	if c.Task != "" {
		refTask, err := config.GetTask(c.Task)
		if err != nil {
			return "", err
		}

		shellProgram, commandStr = formatCmd(refTask.Command)
	} else {
		shellProgram, commandStr = formatCmd(c.Command)
	}

	// Execute Command
	cmd, err := client.Command(shellProgram, commandStr)
	core.CheckIfError(err)

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
		// cmd.Env = append(os.Environ(), defaultArguments...)
		// cmd.Env = append(cmd.Env, c.EnvList...)

		out, err := cmd.Output()
		if err != nil {
			output = err.Error()
		} else {
			output = string(out)
		}

		return output, err
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
