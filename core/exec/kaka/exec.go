// +build exclude
package exec

import (
    "fmt"
    "os/signal"
    "sync"
    "errors"
    "os"

    "github.com/jedib0t/go-pretty/v6/table"
    "github.com/theckman/yacspin"

    "github.com/alajmo/mani/core"
    "github.com/alajmo/mani/core/dao"
    "github.com/alajmo/mani/core/client"
)

func (task *Task) Run(
    userArgs []string,
	projects []Project,
    config *dao.Config,
    runFlags *core.RunFlags,
) {
    servers := exec.Servers

    // Parse env here instead of config since we're only interested in tasks run, and not all tasks.
    // Also, userArgs is not present in the config.
    task.EnvList = dao.GetEnvList(task.Env, userArgs, []string{}, config.EnvList)

    // Describe task
    if runFlags.Describe {
	dao.PrintTaskBlock([]dao.Task{*task})
    }

    // Update output property if user flag is provided
    if runFlags.Output != "" {
	task.SpecData.Output = runFlags.Output
    }

    // Omit servers which provide empty output
    if runFlags.OmitEmpty {
	task.SpecData.OmitEmpty = true
    }

    // If parallel flag is set to true, then update task specs
    if runFlags.Parallel {
	task.SpecData.Parallel = true
    }

    // Set environment variables for sub-commands
    for i := range task.Commands {
	task.Commands[i].EnvList = dao.GetEnvList(task.Commands[i].Env, userArgs, task.EnvList, config.EnvList)
    }

    clientCh := make(chan client.Client, len(servers))
    errCh := make(chan error, len(servers))
    clients, prefixMaxLen, err := exec.createClients(clientCh, errCh)
    if err != nil {
	core.CheckIfError(err)
    }

    var wg sync.WaitGroup
    switch task.SpecData.Output {
    case "table", "html", "markdown" :
	spinner := initSpinner()

	var data core.TableOutput
	var dataMutex = sync.RWMutex{}

	/**
	** Headers
	**/
	data.Headers = append(data.Headers, "Server")
	// Append Command names if set
	for _, cmd := range task.Commands {
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

	// Append Command name if set
	if task.Cmd != "" {
	    data.Headers = append(data.Headers, task.Name)
	}

	// First row values are server names
	for i, c := range clients {
	    _, host, _, _ := c.GetConnectionDetails()
	    data.Rows = append(data.Rows, table.Row{ host })

	    for range task.Commands {
		data.Rows[i] = append(data.Rows[i], "")
	    }

	    if task.Cmd != "" {
		data.Rows[i] = append(data.Rows[i], "")
	    }
	}

	for i, c := range clients {
	    wg.Add(1)
	    if task.SpecData.Parallel {
		go func(i int, c client.Client) {
		    defer wg.Done()
		    exec.tableWork(i, c, data, &dataMutex, config)
		}(i, c)
	    } else {
		func(i int, c client.Client) {
		    defer wg.Done()
		    exec.tableWork(i, c, data, &dataMutex, config)
		}(i, c)
	    }
	}

	cleanupClients(clients, &wg)
	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)


	printTable(task.ThemeData.Table, task.SpecData.OmitEmpty, task.SpecData.Output, data)
	default: // text
	for i, c := range clients {
	    wg.Add(1)
	    colorIndex := core.COLOR_INDEX[i % len(core.COLOR_INDEX)]
	    if task.SpecData.Parallel {
		go func(c client.Client) {
		    defer wg.Done()
		    exec.textWork(c, uint8(colorIndex), prefixMaxLen, config)
		}(c)
	    } else {
		func(c client.Client) {
		    defer wg.Done()
		    exec.textWork(c, uint8(colorIndex), prefixMaxLen, config)
		}(c)
	    }
	}

	cleanupClients(clients, &wg)
	wg.Wait()
    }
}

func (exec Exec) createClients(
    clientCh chan client.Client,
    errCh chan error,
) ([]client.Client, int, error) {
    servers := exec.Servers
    // Establish connection to server
    var wg sync.WaitGroup
    for i, server := range servers {
	wg.Add(1)

	go func(i int, server dao.Server) {
	    defer wg.Done()

	    if server.Local {
		local := &client.LocalhostClient{
		    User: server.User,
		    Host: server.Host,
		    Env: server.EnvList,
		}

		if err := local.Connect(); err != nil {
		    errCh <- errors.New("connecting to localhost failed")
		    return
		}

		clientCh <- local
	    } else {
		remote := &client.SSHClient{
		    User:  server.User,
		    Host: server.Host,
		    Env: server.EnvList,
		    Port: server.Port,
		}

		if err := remote.Connect(); err != nil {
		    errCh <- errors.New("connecting to remote host failed")
		}

		clientCh <- remote
	    }
	}(i, server)
    }
    wg.Wait()

    close(clientCh)
    close(errCh)

    // Configure clients connection
    var prefixMaxLen int = 0
    var clients []client.Client
    for c := range clientCh {
	// Setup max length of prefix
	_, prefixLen := c.Prefix()
	if prefixLen > prefixMaxLen {
	    prefixMaxLen = prefixLen
	}

	clients = append(clients, c)
    }

    // Return on error
    for err := range errCh {
	return []client.Client{}, 0, err
    }

    return clients, prefixMaxLen, nil
}

func cleanupClients(clients []client.Client, wg *sync.WaitGroup) {
    trap := make(chan os.Signal, 1)
    signal.Notify(trap, os.Interrupt)
    go func() {
	for {
	    select {
	    case sig, ok := <-trap:
		if !ok {
		    return
		}
		for _, c := range clients {
		    err := c.Signal(sig)
		    if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		    }
		}
	    }
	}
    }()
    wg.Wait()

    signal.Stop(trap)
    close(trap)

    // Close remote connections
    for _, c := range clients {
	if remote, ok := c.(*client.SSHClient); ok {
	    remote.Close()
	}
    }
}

func initSpinner() yacspin.Spinner {
    spinner, err := dao.TaskSpinner()
    core.CheckIfError(err)

    err = spinner.Start()
    core.CheckIfError(err)

    spinner.Message(" Running")

    return spinner
}

