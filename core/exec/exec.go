package exec

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gookit/color"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

type Exec struct {
	Clients  []Client
	Projects []dao.Project
	Tasks    []dao.Task
	Config   dao.Config
}

type TableCmd struct {
	rIndex int
	cIndex int
	client Client
	dryRun bool

	desc     string
	name     string
	shell    string
	env      []string
	cmd      string
	cmdArr   []string
	numTasks int
}

func (exec *Exec) Run(
	userArgs []string,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) error {
	projects := exec.Projects
	tasks := exec.Tasks

	err := exec.ParseTask(userArgs, runFlags, setRunFlags)
	if err != nil {
		return err
	}

	clientCh := make(chan Client, len(projects))
	errCh := make(chan error, len(projects))
	err = exec.SetClients(clientCh, errCh)
	if err != nil {
		return err
	}

	// Describe task
	if runFlags.Describe {
		out := print.PrintTaskBlock([]dao.Task{tasks[0]}, false, tasks[0].ThemeData.Block, print.GookitFormatter{})
		fmt.Print(out)
	}

	exec.CheckTaskNoColor()

	switch tasks[0].SpecData.Output {
	case "table", "html", "markdown":
		fmt.Println("")
		data := exec.Table(runFlags)
		options := print.PrintTableOptions{
			Theme:            tasks[0].ThemeData,
			Output:           tasks[0].SpecData.Output,
			Color:            *tasks[0].ThemeData.Color,
			AutoWrap:         true,
			OmitEmptyRows:    tasks[0].SpecData.OmitEmptyRows,
			OmitEmptyColumns: tasks[0].SpecData.OmitEmptyColumns,
		}
		print.PrintTable(data.Rows, options, data.Headers[0:1], data.Headers[1:], os.Stdout)
		fmt.Println("")
	default:
		exec.Text(runFlags.DryRun, os.Stdout, os.Stderr)
	}

	return nil
}

func (exec *Exec) RunTUI(
	userArgs []string,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
	output string,
	outWriter io.Writer,
	errWriter io.Writer,
) error {
	projects := exec.Projects
	err := exec.ParseTask(userArgs, runFlags, setRunFlags)
	if err != nil {
		return err
	}

	tasks := exec.Tasks

	clientCh := make(chan Client, len(projects))
	errCh := make(chan error, len(projects))
	err = exec.SetClients(clientCh, errCh)
	if err != nil {
		return err
	}

	data := dao.TableOutput{}
	switch output {
	case "table":
		data = exec.Table(runFlags)
		options := print.PrintTableOptions{
			Theme:            tasks[0].ThemeData,
			Output:           tasks[0].SpecData.Output,
			Color:            *tasks[0].ThemeData.Color,
			AutoWrap:         false,
			OmitEmptyRows:    tasks[0].SpecData.OmitEmptyRows,
			OmitEmptyColumns: tasks[0].SpecData.OmitEmptyColumns,
		}
		print.PrintTable(data.Rows, options, data.Headers[0:1], data.Headers[1:], outWriter)
		return nil
	default:
		exec.Text(runFlags.DryRun, outWriter, errWriter)
	}

	return err
}

func (exec *Exec) SetClients(
	clientCh chan Client,
	errCh chan error,
) error {
	config := exec.Config
	ignoreNonExisting := exec.Tasks[0].SpecData.IgnoreNonExisting
	projects := exec.Projects

	var clients []Client
	for i, project := range projects {
		func(i int, project dao.Project) {
			projectPath, err := core.GetAbsolutePath(config.Path, project.Path, project.Name)
			if err != nil {
				errCh <- &core.FailedToParsePath{Name: projectPath}
				return
			}
			if _, err := os.Stat(projectPath); os.IsNotExist(err) && !ignoreNonExisting {
				errCh <- &core.PathDoesNotExist{Path: projectPath}
				return
			}

			client := Client{Path: projectPath, Name: project.Name, Env: project.EnvList}
			clientCh <- client

			clients = append(clients, client)
		}(i, project)
	}

	close(clientCh)
	close(errCh)

	// Return if there's any errors
	for err := range errCh {
		return err
	}

	exec.Clients = clients

	return nil
}

// ParseTask processes and updates task configurations based on runtime flags and user arguments.
// It handles theme, specification, environment variables, and execution settings for each task.
//
// The function performs these operations for each task:
// 1. Evaluates configuration environment variables
// 2. Updates theme if specified
// 3. Updates spec settings if provided
// 4. Applies runtime execution flags
// 5. Processes environment variables for the task and its commands
//
// Environment variable processing order:
// 1. Configuration level variables
// 2. Task level variables
// 3. Command level variables
// 4. User provided arguments
func (exec *Exec) ParseTask(userArgs []string, runFlags *core.RunFlags, setRunFlags *core.SetRunFlags) error {
	configEnv, err := dao.EvaluateEnv(exec.Config.EnvList)
	if err != nil {
		return err
	}

	for i := range exec.Tasks {
		// Update theme property if user flag is provided
		if runFlags.Theme != "" {
			theme, err := exec.Config.GetTheme(runFlags.Theme)
			if err != nil {
				return err
			}

			exec.Tasks[i].ThemeData = *theme
		}

		if runFlags.Spec != "" {
			spec, err := exec.Config.GetSpec(runFlags.Spec)
			if err != nil {
				return err
			}
			exec.Tasks[i].SpecData = *spec
		}

		// Update output property if user flag is provided
		if runFlags.Output != "" {
			exec.Tasks[i].SpecData.Output = runFlags.Output
		}

		// TTY
		if setRunFlags.TTY {
			exec.Tasks[i].TTY = runFlags.TTY
		}

		// Omit rows which provide empty output
		if setRunFlags.OmitEmptyRows {
			exec.Tasks[i].SpecData.OmitEmptyRows = runFlags.OmitEmptyRows
		}

		// Omit columns which provide empty output
		if setRunFlags.OmitEmptyColumns {
			exec.Tasks[i].SpecData.OmitEmptyColumns = runFlags.OmitEmptyColumns
		}

		if setRunFlags.IgnoreErrors {
			exec.Tasks[i].SpecData.IgnoreErrors = runFlags.IgnoreErrors
		}

		if setRunFlags.IgnoreNonExisting {
			exec.Tasks[i].SpecData.IgnoreNonExisting = runFlags.IgnoreNonExisting
		}

		// If parallel flag is set to true, then update task specs
		if setRunFlags.Parallel {
			exec.Tasks[i].SpecData.Parallel = runFlags.Parallel
		}

		if setRunFlags.Forks {
			exec.Tasks[i].SpecData.Forks = runFlags.Forks
		}

		// Parse env here instead of config since we're only interested in tasks run, and not all tasks.
		// Also, userArgs is not present in the config.
		envs, err := dao.ParseTaskEnv(exec.Tasks[i].Env, userArgs, []string{}, configEnv)
		if err != nil {
			return err
		}
		exec.Tasks[i].EnvList = envs

		// Set environment variables for sub-commands
		for j := range exec.Tasks[i].Commands {
			envs, err := dao.ParseTaskEnv(exec.Tasks[i].Commands[j].Env, userArgs, exec.Tasks[i].EnvList, configEnv)
			if err != nil {
				return err
			}
			exec.Tasks[i].Commands[j].EnvList = envs
		}
	}

	return nil
}

func (exec *Exec) CheckTaskNoColor() {
	task := exec.Tasks[0]

	for _, env := range task.EnvList {
		name := strings.Split(env, "=")[0]
		if name == "NO_COLOR" {
			color.Disable()
			break
		}
	}
}
