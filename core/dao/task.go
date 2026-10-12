package dao

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/theckman/yacspin"
	"gopkg.in/yaml.v3"

	core "github.com/alajmo/mani/core"
)

var (
	buildMode = "dev"
)

type Command struct {
	Name    string    `yaml:"name"`
	Desc    string    `yaml:"desc"`
	Shell   string    `yaml:"shell"` // should be in the format: <program> <command flag>, for instance "sh -c", "node -e"
	Cmd     string    `yaml:"cmd"`   // "echo hello world", it should not include the program flag (-c,-e, .etc)
	Task    string    `yaml:"task"`
	TaskRef string    `yaml:"-"` // Keep a reference to the task
	TTY     bool      `yaml:"tty"`
	Env     yaml.Node `yaml:"env"`
	EnvList []string  `yaml:"-"`

	// Internal
	ShellProgram string   `yaml:"-"` // should be in the format: <program>, example: "sh", "node"
	CmdArg       []string `yaml:"-"` // is in the format ["-c echo hello world"] or ["-c", "echo hello world"], it includes the shell flag
}

type Task struct {
	SpecData   Spec
	TargetData Target
	ThemeData  Theme

	Name     string    `yaml:"name"`
	Desc     string    `yaml:"desc"`
	Shell    string    `yaml:"shell"`
	Cmd      string    `yaml:"cmd"`
	Commands []Command `yaml:"commands"`
	EnvList  []string  `yaml:"-"`
	TTY      bool      `yaml:"tty"`

	Env    yaml.Node `yaml:"env"`
	Spec   yaml.Node `yaml:"spec"`
	Target yaml.Node `yaml:"target"`
	Theme  yaml.Node `yaml:"theme"`

	// Internal
	ShellProgram string   `yaml:"-"` // should be in the format: <program>, example: "sh", "node"
	CmdArg       []string `yaml:"-"` // is in the format ["-c echo hello world"] or ["-c", "echo hello world"], it includes the shell flag
	context      string
	contextLine  int
	node         *yaml.Node        // definition in the config, used to merge tasks defined in multiple configs
	envs         []Env             // env with the config each comes from, used to merge tasks
	fieldSources map[string]source // config each field comes from, for tasks defined in multiple configs
}

func (t *Task) GetContext() string {
	return t.context
}

func (t *Task) GetContextLine() int {
	return t.contextLine
}

// ParseTask parses tasks and builds the correct "AST". Depending on if the data is specified inline,
// or if it is a reference to resource, it will handle them differently. Errors in a field of a task
// defined in multiple configs are reported at the config that sets the field.
func (t *Task) ParseTask(config Config) []ResourceErrors[Task] {
	var taskErrors []ResourceErrors[Task]
	addError := func(field string, err error) {
		resource := t
		if s, found := t.fieldSources[field]; found {
			resource = &Task{Name: t.Name, context: s.context, contextLine: s.contextLine}
		}
		taskErrors = append(taskErrors, ResourceErrors[Task]{Resource: resource, Errors: []error{err}})
	}

	if t.Shell == "" {
		t.Shell = config.Shell
	} else {
		t.Shell = core.FormatShell(t.Shell)
	}

	program, cmdArgs := core.FormatShellString(t.Shell, t.Cmd)
	t.ShellProgram = program
	t.CmdArg = cmdArgs

	for j, cmd := range t.Commands {
		// Task reference
		if cmd.Task != "" {
			cmdRef, err := config.GetCommand(cmd.Task)
			if err != nil {
				addError("cmd", err)
				continue
			}

			t.Commands[j] = *cmdRef
			t.Commands[j].TaskRef = cmd.Task
		}

		if t.Commands[j].Shell == "" {
			t.Commands[j].Shell = DEFAULT_SHELL
		}

		program, cmdArgs := core.FormatShellString(t.Commands[j].Shell, t.Commands[j].Cmd)
		t.Commands[j].ShellProgram = program
		t.Commands[j].CmdArg = cmdArgs
	}

	// Aliases are resolved first, since the value of an alias is the name of its anchor
	themeNode := resolveAlias(&t.Theme)
	if len(themeNode.Content) > 0 {
		// Theme value
		theme := &Theme{}
		err := themeNode.Decode(theme)
		if err != nil {
			addError("theme", err)
		} else {
			t.ThemeData = *theme
		}
	} else if themeNode.Value != "" {
		// Theme reference
		theme, err := config.GetTheme(themeNode.Value)
		if err != nil {
			addError("theme", err)
		} else {
			t.ThemeData = *theme
		}
	} else {
		// Default theme
		theme, err := config.GetTheme(DEFAULT_THEME.Name)
		if err != nil {
			addError("theme", err)
		} else {
			t.ThemeData = *theme
		}
	}

	specNode := resolveAlias(&t.Spec)
	if len(specNode.Content) > 0 {
		// Spec value
		spec := &Spec{}
		err := specNode.Decode(spec)

		if err != nil {
			addError("spec", err)
		} else {
			t.SpecData = *spec
		}
	} else if specNode.Value != "" {
		// Spec reference
		spec, err := config.GetSpec(specNode.Value)
		if err != nil {
			addError("spec", err)
		} else {
			t.SpecData = *spec
		}
	} else {
		// Default spec
		spec, err := config.GetSpec(DEFAULT_SPEC.Name)
		if err != nil {
			addError("spec", err)
		} else {
			t.SpecData = *spec
		}
	}

	targetNode := resolveAlias(&t.Target)
	if len(targetNode.Content) > 0 {
		// Target value, whose paths are relative to the config that sets it
		target := &Target{}
		err := targetNode.Decode(target)
		if err != nil {
			addError("target", err)
		} else {
			context := t.context
			if s, found := t.fieldSources["target"]; found {
				context = s.context
			}
			target.dir = config.pathsDir(context)
			t.TargetData = *target
		}
	} else if targetNode.Value != "" {
		// Target reference
		target, err := config.GetTarget(targetNode.Value)
		if err != nil {
			addError("target", err)
		} else {
			t.TargetData = *target
		}
	} else {
		// Default target
		target, err := config.GetTarget(DEFAULT_TARGET.Name)
		if err != nil {
			addError("target", err)
		} else {
			t.TargetData = *target
		}
	}

	return taskErrors
}

// MergeTasks combines tasks that are defined in multiple config files. Each field comes from the
// config that imports the other configs setting it, and env is merged by name the same way. If
// neither config imports the other, different values are an error. cmd and commands count as one
// field, so overriding cmd also replaces commands.
//
// A task may only be defined once per config file. imports returns true if config file a imports
// config file b.
func MergeTasks(tasks []Task, imports func(a string, b string) bool) ([]Task, []ResourceErrors[Task]) {
	merged, taskErrors := mergeByName(tasks, "task",
		func(t *Task) string { return t.Name },
		func(defs []*Task) (Task, int, []ResourceErrors[Task]) { return mergeTask(defs, imports) },
	)

	// The definitions are only needed for merging
	for i := range merged {
		merged[i].node = nil
		merged[i].envs = nil
	}

	return merged, taskErrors
}

// mergeTask merges the definitions of a task, and returns the index of the definition
// that owns it, meaning the one its cmd comes from.
func mergeTask(defs []*Task, imports func(a string, b string) bool) (Task, int, []ResourceErrors[Task]) {
	var taskErrors []ResourceErrors[Task]
	addError := func(def *Task, err error) {
		taskErrors = append(taskErrors, ResourceErrors[Task]{Resource: def, Errors: []error{err}})
	}

	name := defs[0].Name

	type field struct {
		key   string
		value string
		def   *Task
	}
	fields := mergeByKey(defs, "task", name, "",
		func(t *Task) []field {
			fieldNodes := taskFields(t.node)
			var fields []field
			for _, key := range []string{"desc", "shell", "tty", "spec", "target", "theme"} {
				if node, found := fieldNodes[key]; found {
					fields = append(fields, field{key: key, value: nodeString(node), def: t})
				}
			}

			var command []string
			if node, found := fieldNodes["cmd"]; found {
				command = append(command, nodeString(node))
			}
			if node, found := fieldNodes["commands"]; found {
				command = append(command, "commands: "+nodeString(node))
			}
			if len(command) > 0 {
				fields = append(fields, field{key: "cmd", value: strings.Join(command, ", "), def: t})
			}

			return fields
		},
		func(f field) (string, string) { return f.key, f.value },
		imports,
		addError,
	)

	task := Task{Name: name, fieldSources: make(map[string]source)}
	owner := defs[0]
	for _, f := range fields {
		task.fieldSources[f.key] = sourceOf(f.def)
		switch f.key {
		case "desc":
			task.Desc = f.def.Desc
		case "shell":
			task.Shell = f.def.Shell
		case "tty":
			task.TTY = f.def.TTY
		case "spec":
			task.Spec = f.def.Spec
		case "target":
			task.Target = f.def.Target
		case "theme":
			task.Theme = f.def.Theme
		case "cmd":
			task.Cmd = f.def.Cmd
			task.Commands = f.def.Commands
			owner = f.def
		}
	}
	task.context = owner.context
	task.contextLine = owner.contextLine

	// Env is kept as a YAML node, since it's evaluated when the task runs
	envs := mergeByKey(defs, "task", name, "env",
		func(t *Task) []Env { return t.envs },
		func(e Env) (string, string) { return e.Name, e.Value },
		imports,
		addError,
	)
	if len(envs) > 0 {
		task.Env = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, env := range envs {
			task.Env.Content = append(task.Env.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: env.Name},
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: env.Value},
			)
		}
	}

	return task, slices.Index(defs, owner), taskErrors
}

// taskFields returns the fields a task definition sets. Like project fields, a field that's null or
// an empty string isn't set, and neither is a task without a value, so that they don't override the
// value from another config.
func taskFields(node *yaml.Node) map[string]*yaml.Node {
	node = resolveAlias(node)
	if node.Kind == yaml.ScalarNode {
		if isEmpty(node) {
			return nil
		}
		return map[string]*yaml.Node{"cmd": node}
	}

	fields := make(map[string]*yaml.Node)
	pairs, _ := mappingPairs(node)
	for _, p := range pairs {
		if !isEmpty(p.value) {
			fields[p.key.Value] = p.value
		}
	}

	return fields
}

func TaskSpinner() (yacspin.Spinner, error) {
	var cfg yacspin.Config

	// NOTE: Don't print the spinner in tests since it causes
	// golden files to produce different results.
	if buildMode == "TEST" {
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
			ShowCursor:      true,
		}
	}

	spinner, err := yacspin.New(cfg)

	return *spinner, err
}

func (t Task) GetValue(key string, _ int) string {
	switch strings.ToLower(key) {
	case "name", "task":
		return t.Name
	case "desc", "description":
		return t.Desc
	case "command":
		return t.Cmd
	case "spec":
		return t.SpecData.Name
	case "target":
		return t.TargetData.Name
	default:
		return ""
	}
}

func (c *Config) GetTaskList() ([]Task, []ResourceErrors[Task]) {
	var tasks []Task
	count := len(c.Tasks.Content)

	taskErrors := []ResourceErrors[Task]{}
	foundErrors := false
	for i := 0; i < count; i += 2 {
		task := &Task{
			Name:        c.Tasks.Content[i].Value,
			context:     c.Path,
			contextLine: c.Tasks.Content[i].Line,
		}

		// Shorthand definition: example_task: echo 123
		node := resolveAlias(c.Tasks.Content[i+1])
		task.node = node
		if node.Kind == yaml.ScalarNode {
			if node.Tag != "!!null" {
				task.Cmd = node.Value
			}
		} else { // Full definition
			err := node.Decode(task)
			if err != nil {
				foundErrors = true
				taskError := ResourceErrors[Task]{Resource: task, Errors: decodeErrors(err)}
				taskErrors = append(taskErrors, taskError)
				continue
			}

			var envErrors []error
			task.envs, err = parseEnv(&task.Env, c.Path)
			if err != nil {
				envErrors = append(envErrors, err)
			}
			for _, cmd := range task.Commands {
				if _, err := parseEnv(&cmd.Env, c.Path); err != nil {
					envErrors = append(envErrors, err)
				}
			}

			if len(envErrors) > 0 {
				foundErrors = true
				taskErrors = append(taskErrors, ResourceErrors[Task]{Resource: task, Errors: envErrors})
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

func ParseTaskEnv(
	env yaml.Node,
	userEnv []string,
	parentEnv []string,
	configEnv []string,
) ([]string, error) {
	cmdEnv, err := EvaluateEnv(ParseNodeEnv(env))
	if err != nil {
		return []string{}, err
	}

	pEnv, err := EvaluateEnv(parentEnv)
	if err != nil {
		return []string{}, err
	}

	envList := MergeEnvs(userEnv, cmdEnv, pEnv, configEnv)

	return envList, nil
}

func ParseTasksEnv(tasks []Task) {
	for i := range tasks {
		envs, err := ParseTaskEnv(tasks[i].Env, []string{}, []string{}, []string{})
		core.CheckIfError(err)

		tasks[i].EnvList = envs

		for j := range tasks[i].Commands {
			envs, err = ParseTaskEnv(tasks[i].Commands[j].Env, []string{}, []string{}, []string{})
			core.CheckIfError(err)

			tasks[i].Commands[j].EnvList = envs
		}
	}
}

// GetTaskProjects retrieves a filtered list of projects for a given task, applying
// runtime flag overrides and target configurations.
//
// Behavior depends on the provided runtime flags (flags, setFlags) and task target:
//   - If runtime flags are set (Projects, Paths, Tags, etc.), they take precedence
//     and reset the task's target configuration.
//   - If a target is explicitly specified (flags.Target), it loads and applies that
//     target's configuration before applying runtime flag overrides.
//   - If no runtime flags or target are provided, the task's default target data is used.
//
// Filtering priority (highest to lowest):
//  1. Runtime flags (e.g., --projects, --tags, --cwd)
//  2. Explicit target configuration (--target)
//  3. Task's default target data (if no overrides exist)
//
// Returns:
//   - Filtered []Project based on the resolved configuration.
//   - Non-nil error if target resolution or project filtering fails.
func (c Config) GetTaskProjects(
	task *Task,
	flags *core.RunFlags,
	setFlags *core.SetRunFlags,
) ([]Project, error) {
	var err error
	var projects []Project

	// Reset target if any runtime flags are used
	if len(flags.Projects) > 0 ||
		len(flags.Paths) > 0 ||
		len(flags.Tags) > 0 ||
		flags.TagsExpr != "" ||
		flags.Target != "" ||
		setFlags.Cwd ||
		setFlags.All {
		task.TargetData = Target{}
	}

	if flags.Target != "" {
		target, err := c.GetTarget(flags.Target)
		if err != nil {
			return []Project{}, err
		}
		task.TargetData = *target
	}

	if len(flags.Projects) > 0 {
		task.TargetData.Projects = flags.Projects
	}

	if len(flags.Paths) > 0 {
		task.TargetData.Paths = flags.Paths
		task.TargetData.dir = ""
	}

	if len(flags.Tags) > 0 {
		task.TargetData.Tags = flags.Tags
	}

	if flags.TagsExpr != "" {
		task.TargetData.TagsExpr = flags.TagsExpr
	}

	if setFlags.Cwd {
		task.TargetData.Cwd = flags.Cwd
	}

	if setFlags.All {
		task.TargetData.All = flags.All
	}

	projects, err = c.filterProjects(task.TargetData)
	if err != nil {
		return []Project{}, err
	}

	return projects, nil
}

func (c Config) GetTasksByNames(names []string) ([]Task, error) {
	if len(names) == 0 {
		return c.TaskList, nil
	}

	foundTasks := make(map[string]bool)
	for _, t := range names {
		foundTasks[t] = false
	}

	var filteredTasks []Task
	for _, name := range names {
		if foundTasks[name] {
			continue
		}

		for _, task := range c.TaskList {
			if name == task.Name {
				foundTasks[task.Name] = true
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	nonExistingTasks := []string{}
	for k, v := range foundTasks {
		if !v {
			nonExistingTasks = append(nonExistingTasks, k)
		}
	}

	if len(nonExistingTasks) > 0 {
		return []Task{}, &core.TaskNotFound{Name: nonExistingTasks}
	}

	return filteredTasks, nil
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

func (c Config) GetTask(name string) (*Task, error) {
	for _, cmd := range c.TaskList {
		if name == cmd.Name {
			return &cmd, nil
		}
	}

	return nil, &core.TaskNotFound{Name: []string{name}}
}

func (c Config) GetCommand(taskName string) (*Command, error) {
	for _, cmd := range c.TaskList {
		if taskName == cmd.Name {
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

	return nil, &core.TaskNotFound{Name: []string{taskName}}
}

func (t Task) ConvertTaskToCommand() Command {
	cmd := Command{
		Name:         t.Name,
		Desc:         t.Desc,
		EnvList:      t.EnvList,
		Shell:        t.Shell,
		Cmd:          t.Cmd,
		CmdArg:       t.CmdArg,
		ShellProgram: t.ShellProgram,
	}

	return cmd
}

func ParseCmd(
	cmd string,
	runFlags *core.RunFlags,
	setFlags *core.SetRunFlags,
	config *Config,
) ([]Task, []Project, error) {
	task := Task{Name: "output", Cmd: cmd, TTY: runFlags.TTY}
	taskErrors := task.ParseTask(*config)

	var configErr = ""
	for _, taskError := range taskErrors {
		if len(taskError.Errors) > 0 {
			configErr = fmt.Sprintf("%s%s", configErr, FormatErrors(taskError.Resource, taskError.Errors))
		}
	}
	if configErr != "" {
		core.CheckIfError(errors.New(configErr))
	}

	projects, err := config.GetTaskProjects(&task, runFlags, setFlags)
	if err != nil {
		return nil, nil, err
	}
	core.CheckIfError(err)

	var tasks []Task
	for range projects {
		t := Task{}
		err := copier.Copy(&t, &task)
		core.CheckIfError(err)
		tasks = append(tasks, t)
	}

	if len(projects) == 0 {
		return nil, nil, &core.NoTargets{}
	}

	return tasks, projects, err
}

func ParseSingleTask(
	taskName string,
	runFlags *core.RunFlags,
	setFlags *core.SetRunFlags,
	config *Config,
) ([]Task, []Project, error) {
	task, err := config.GetTask(taskName)
	core.CheckIfError(err)

	projects, err := config.GetTaskProjects(task, runFlags, setFlags)
	core.CheckIfError(err)

	var tasks []Task
	for range projects {
		t := Task{}
		err := copier.Copy(&t, &task)
		core.CheckIfError(err)
		tasks = append(tasks, t)
	}

	if len(projects) == 0 {
		return nil, nil, &core.NoTargets{}
	}

	return tasks, projects, err
}

func ParseManyTasks(
	taskNames []string,
	runFlags *core.RunFlags,
	setFlags *core.SetRunFlags,
	config *Config,
) ([]Task, []Project, error) {
	parentTask := Task{Name: "Tasks", Cmd: "", Commands: []Command{}}
	parentTask.ParseTask(*config)

	for _, taskName := range taskNames {
		task, err := config.GetTask(taskName)
		core.CheckIfError(err)

		if task.Cmd != "" {
			cmd := task.ConvertTaskToCommand()
			parentTask.Commands = append(parentTask.Commands, cmd)
		} else if len(task.Commands) > 0 {
			parentTask.Commands = append(parentTask.Commands, task.Commands...)
		}
	}

	projects, err := config.GetTaskProjects(&parentTask, runFlags, setFlags)
	var tasks []Task
	for range projects {
		t := Task{}
		err := copier.Copy(&t, &parentTask)
		core.CheckIfError(err)
		tasks = append(tasks, t)
	}

	if len(projects) == 0 {
		return nil, nil, &core.NoTargets{}
	}

	return tasks, projects, err
}
