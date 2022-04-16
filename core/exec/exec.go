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
	Tasks []dao.Task
	Config dao.Config
}

func (exec *Exec) Run(
	userArgs []string,
	runFlags *core.RunFlags,
	setRunFlags *core.SetRunFlags,
) error {
	projects := exec.Projects
	tasks := exec.Tasks
	config := &exec.Config

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

	exec.ParseTask(userArgs, runFlags, setRunFlags)

	switch tasks[0].SpecData.Output {
	case "table", "html", "markdown" :
		data := exec.Table(runFlags.DryRun)
		options := print.PrintTableOptions { Theme: tasks[0].ThemeData.Name, OmitEmpty: tasks[0].SpecData.OmitEmpty, Output: tasks[0].SpecData.Output,  SuppressEmptyColumns: false }
		print.PrintTable(config, data.Rows, options, data.Headers[0:1], data.Headers[1:])
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

			client := Client { Path: projectPath, Name: project.Name, Env: project.EnvList }
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

func (exec *Exec) ParseTask(userArgs []string, runFlags *core.RunFlags, setRunFlags *core.SetRunFlags) {
	for i := range exec.Tasks {
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
		exec.Tasks[i].EnvList = dao.GetEnvList(exec.Tasks[i].Env, userArgs, []string{}, exec.Config.EnvList)

		// Set environment variables for sub-commands
		for j := range exec.Tasks[i].Commands {
			exec.Tasks[i].Commands[j].EnvList = dao.GetEnvList(exec.Tasks[i].Commands[j].Env, userArgs, exec.Tasks[i].EnvList, exec.Config.EnvList)
		}
	}
}
