package exec

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"

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

	clientCh := make(chan Client, len(projects))
	errCh := make(chan error, len(projects))
	err := exec.SetClients(clientCh, errCh)
	if err != nil {
		return err
	}

	// Describe task
	if runFlags.Describe {
		print.PrintTaskBlock([]dao.Task{tasks[0]})
	}

	err = exec.ParseTask(userArgs, runFlags, setRunFlags)
	if err != nil {
		return err
	}
	exec.CheckTaskNoColor()

	switch tasks[0].SpecData.Output {
	case "table", "html", "markdown":
		data := exec.Table(runFlags)
		options := print.PrintTableOptions{Theme: tasks[0].ThemeData, OmitEmpty: tasks[0].SpecData.OmitEmpty, Output: tasks[0].SpecData.Output, SuppressEmptyColumns: false}
		print.PrintTable(data.Rows, options, data.Headers[0:1], data.Headers[1:])
	default:
		exec.Text(runFlags.DryRun)
	}

	return nil
}

func (exec *Exec) SetClients(
	clientCh chan Client,
	errCh chan error,
) error {
	config := exec.Config
	projects := exec.Projects

	var clients []Client
	for i, project := range projects {
		func(i int, project dao.Project) {
			projectPath, err := core.GetAbsolutePath(config.Path, project.Path, project.Name)
			if err != nil {
				errCh <- &core.FailedToParsePath{Name: projectPath}
				return
			}
			if _, err := os.Stat(projectPath); os.IsNotExist(err) {
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

		// Update output property if user flag is provided
		if runFlags.Output != "" {
			exec.Tasks[i].SpecData.Output = runFlags.Output
		}

		// Omit projects which provide empty output
		if setRunFlags.OmitEmpty {
			exec.Tasks[i].SpecData.OmitEmpty = runFlags.OmitEmpty
		}

		// If parallel flag is set to true, then update task specs
		if setRunFlags.Parallel {
			exec.Tasks[i].SpecData.Parallel = runFlags.Parallel
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

func (e *Exec) CheckTaskNoColor() {
	task := e.Tasks[0]

	for _, env := range task.EnvList {
		name := strings.Split(env, "=")[0]
		if name == "NO_COLOR" {
			text.DisableColors()
		}
	}
}
