package dao

import (
	"fmt"
	"io"
	"time"

	"github.com/theckman/yacspin"
	"gopkg.in/yaml.v3"

	core "github.com/alajmo/mani/core"
)

var (
	build_mode = "dev"
)

type Command struct {
	Name    string `yaml:"name"`
	Desc    string `yaml:"desc"`
	EnvList []string
	Shell   string `yaml:"shell"`
	Cmd     string `yaml:"cmd"`
	Task    string `yaml:"task"`

	Env yaml.Node `yaml:"env"`
}

type Task struct {
	SpecData  Spec
	TargetData  Target
	ThemeData Theme

	Name     string    `yaml:"name"`
	Desc     string    `yaml:"desc"`
	EnvList  []string
	Shell    string `yaml:"shell"`
	Cmd      string `yaml:"cmd"`
	Commands []Command

	Env      yaml.Node `yaml:"env"`
	Spec     yaml.Node `yaml:"spec"`
	Target   yaml.Node `yaml:"target"`
	Theme    yaml.Node `yaml:"theme"`

	context string
	contextLine int
}

func (t *Task) GetContext() string {
	return t.context
}

func (t *Task) GetContextLine() int {
	return t.contextLine
}

func (t *Task) ParseTask(config Config, taskErrors *ResourceErrors[Task]) {
	if t.Shell == "" {
		t.Shell = config.Shell
	} else {
		t.Shell = core.FormatShell(t.Shell)
	}

	for j, cmd := range t.Commands {
		if cmd.Task != "" {
			cmdRef, err := config.GetCommand(cmd.Task)
			if err != nil {
				taskErrors.Errors = append(taskErrors.Errors, err)
				continue
			}

			t.Commands[j] = *cmdRef
		}

		if t.Commands[j].Shell == "" {
			t.Commands[j].Shell = DEFAULT_SHELL
		}
	}

	if len(t.Theme.Content) > 0 {
		// Theme value
		theme := &Theme{}
		err := t.Theme.Decode(theme)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.ThemeData = *theme
		}
	} else if t.Theme.Value != "" {
		// Theme reference
		theme, err := config.GetTheme(t.Theme.Value)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.ThemeData = *theme
		}
	} else {
		// Default theme
		theme, err := config.GetTheme(DEFAULT_THEME.Name)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.ThemeData = *theme
		}
	}

	if len(t.Spec.Content) > 0 {
		// Spec value
		spec := &Spec{}
		err := t.Spec.Decode(spec)

		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.SpecData = *spec
		}
	} else if t.Spec.Value != "" {
		// Spec reference
		spec, err := config.GetSpec(t.Spec.Value)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.SpecData = *spec
		}
	} else {
		// Default spec
		spec, err := config.GetSpec(DEFAULT_SPEC.Name)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.SpecData = *spec
		}
	}

	if len(t.Target.Content) > 0 {
		// Target value
		target := &Target{}
		err := t.Target.Decode(target)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.TargetData = *target
		}
	} else if t.Target.Value != "" {
		// Target reference
		target, err := config.GetTarget(t.Target.Value)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.TargetData = *target
		}
	} else {
		// Default target
		target, err := config.GetTarget(DEFAULT_TARGET.Name)
		if err != nil {
			taskErrors.Errors = append(taskErrors.Errors, err)
		} else {
			t.TargetData = *target
		}
	}
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

func (t Task) GetValue(key string, _ int) string {
	switch key {
	case "Name", "name", "Task", "task":
		return t.Name
	case "Desc", "desc", "Description", "description":
		return t.Desc
	case "Command", "command":
		return t.Cmd
	}

	return ""
}

func (c *Config) GetTaskList() ([]Task, []ResourceErrors[Task]) {
	var tasks []Task
	count := len(c.Tasks.Content)

	taskErrors := []ResourceErrors[Task]{}
	foundErrors := false
	for i := 0; i < count; i += 2 {
		task := &Task{
			Name: c.Tasks.Content[i].Value,
			context: c.Path,
			contextLine: c.Tasks.Content[i].Line,
		}

		// Shorthand definition: example_task: echo 123
		if c.Tasks.Content[i+1].Kind == 8 {
			task.Cmd = c.Tasks.Content[i+1].Value
		} else { // Full definition
			err := c.Tasks.Content[i+1].Decode(task)
			if err != nil {
				foundErrors = true
				taskError := ResourceErrors[Task]{ Resource: task, Errors: StringsToErrors(err.(*yaml.TypeError).Errors) }
				taskErrors = append(taskErrors, taskError)
				continue
			}
		}

		tasks = append(tasks, *task)
	}

	if foundErrors {
		return tasks, taskErrors
	}

	return tasks, nil
}

func GetEnvList(env yaml.Node, userEnv []string, parentEnv []string, configEnv []string) []string {
	pEnv, err := core.EvaluateEnv(parentEnv)
	core.CheckIfError(err)

	cmdEnv, err := core.EvaluateEnv(core.GetEnv(env))
	core.CheckIfError(err)

	globalEnv, err := core.EvaluateEnv(configEnv)
	core.CheckIfError(err)

	envList := core.MergeEnvs(userEnv, cmdEnv, pEnv, globalEnv)

	return envList
}

func (c Config) GetTaskProjects(task *Task, runFlags *core.RunFlags) ([]Project) {
	var projects []Project
	// If any runtime target flags are used, disregard task targets
	if len(runFlags.Projects) > 0 || len(runFlags.Paths) > 0 || len(runFlags.Tags) > 0 || runFlags.Cwd || runFlags.All {
		projects = c.FilterProjects(runFlags.Cwd, runFlags.All, runFlags.Paths, runFlags.Projects, runFlags.Tags)
	} else {
		projects = c.FilterProjects(task.TargetData.Cwd, task.TargetData.All, task.TargetData.Paths, task.TargetData.Projects, task.TargetData.Tags)
	}

	return projects
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

func (c Config) GetTaskNameAndDesc() []string {
	taskNames := []string{}
	for _, task := range c.TaskList {
		taskNames = append(taskNames, fmt.Sprintf("%s\t%s", task.Name, task.Desc))
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
				Cmd:     cmd.Cmd,
			}

			return cmdRef, nil
		}
	}

	return nil, &core.TaskNotFound{Name: task}
}
