package exec

import (
	"os"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

type Exec struct {
	Clients []Client
	Projects []dao.Project
	Task dao.Task
	Config dao.Config
}

func (exec *Exec) Run(
	userArgs []string,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) error {
	projects := exec.Projects
	task := &exec.Task
	config := &exec.Config

	clientCh := make(chan Client, len(projects))
	errCh := make(chan error, len(projects))
	err := exec.setClients(clientCh, errCh)
	if err != nil {
		return err
	}

	// Describe task
	if runFlags.Describe {
		print.PrintTaskBlock([]dao.Task{*task})
	}

	// Update output property if user flag is provided
	if runFlags.Output != "" {
		task.SpecData.Output = runFlags.Output
	}

	// Omit projects which provide empty output
	if setRunFlags.OmitEmpty {
		task.SpecData.OmitEmpty = runFlags.OmitEmpty
	}

	// If parallel flag is set to true, then update task specs
	if setRunFlags.Parallel {
		task.SpecData.Parallel = runFlags.Parallel
	}

	// Parse env here instead of config since we're only interested in tasks run, and not all tasks.
	// Also, userArgs is not present in the config.
	task.EnvList = dao.GetEnvList(task.Env, userArgs, []string{}, config.EnvList)

	// Set environment variables for sub-commands
	for i := range task.Commands {
		task.Commands[i].EnvList = dao.GetEnvList(task.Commands[i].Env, userArgs, task.EnvList, config.EnvList)
	}

	switch task.SpecData.Output {
	case "table", "html", "markdown" :
		data := exec.Table(runFlags.DryRun)
		options := print.PrintTableOptions { Theme: task.ThemeData.Name, OmitEmpty: task.SpecData.OmitEmpty, Output: task.SpecData.Output,  SuppressEmptyColumns: false }
		print.PrintTable(config, data.Rows, options, data.Headers[0:1], data.Headers[1:])
	default:
		exec.Text(runFlags.DryRun)
	}

	return nil
}

func (exec *Exec) setClients(
	clientCh chan Client,
	errCh chan error,
) error {
	config := exec.Config
	task := exec.Task
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

			client := Client { User: task.User, Path: projectPath, Name: project.Name, Env: project.EnvList }
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
