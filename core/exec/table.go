package exec

import (
    "fmt"
    "sync"
    "io"
    "io/ioutil"
    "strings"
    "os"

    "github.com/theckman/yacspin"

    core "github.com/alajmo/mani/core"
    dao "github.com/alajmo/mani/core/dao"
)


func (exec *Exec) Table() dao.TableOutput {
    task := exec.Task
    clients := exec.Clients
    projects := exec.Projects

    spinner := initSpinner()

    var data dao.TableOutput
    var dataMutex = sync.RWMutex{}


    /**
    ** Headers
    **/
    data.Headers = append(data.Headers, "project")

    // Append Command names if set
    for _, subTask := range task.Commands {
	if subTask.Name != "" {
	    data.Headers = append(data.Headers, subTask.Name)
	} else {
	    data.Headers = append(data.Headers, "output")
	}
    }

    // Append Command name if set
    if task.Cmd != "" {
	if task.Name != "" {
	    data.Headers = append(data.Headers, task.Name)
	} else {
	    data.Headers = append(data.Headers, "output")
	}
    }

    // Populate the rows (project name is first cell, then commands and cmd output is set to empty string)
    for i, p := range projects {
	data.Rows = append(data.Rows, dao.Row { Columns: []string{p.Name} })

	for range task.Commands {
	    data.Rows[i].Columns = append(data.Rows[i].Columns, "")
	}

	if task.Cmd != "" {
	    data.Rows[i].Columns = append(data.Rows[i].Columns, "")
	}
    }

    var wg sync.WaitGroup
    /**
    ** Values
    **/
    for i, c := range clients {
	wg.Add(1)
	if task.SpecData.Parallel {
	    go func(i int, c Client) {
		defer wg.Done()
		exec.TableWork(i, data, &dataMutex)
	    }(i, c)
	} else {
	    func(i int, c Client) {
		defer wg.Done()
		exec.TableWork(i, data, &dataMutex)
	    }(i, c)
	}
    }
    wg.Wait()

    err := spinner.Stop()
    core.CheckIfError(err)

    return data
}

func (exec *Exec) TableWork(rIndex int, data dao.TableOutput, dataMutex *sync.RWMutex) {
    client := exec.Clients[rIndex]
    task := exec.Task
    var wg sync.WaitGroup

    for j, cmd := range task.Commands {
	err := RunTableCmd(rIndex, j, client, cmd.Shell, cmd.EnvList, cmd.Cmd, data, dataMutex, &wg)
	if err != nil && !task.SpecData.IgnoreError {
	    fmt.Println(err)
	}

	if err != nil {
	    fmt.Println(err)
	}
    }

    if task.Cmd != "" {
	err := RunTableCmd(rIndex, len(task.Commands), client, task.Shell, task.EnvList, task.Cmd, data, dataMutex, &wg)
	if err != nil {
	    fmt.Println(err)
	}
    }

    wg.Wait()
}

func RunTableCmd(
    rIndex int,
    cIndex int,
    c Client,
    shell string,
    env []string,
    cmd string,
    data dao.TableOutput,
    dataMutex *sync.RWMutex,
    wg *sync.WaitGroup,
) error {
    err := c.Run(shell, env, cmd)
    if err != nil {
	return err
    }

    // Copy over commands STDOUT.
    var stdoutHandler = func(c Client) {
	defer wg.Done()
	dataMutex.Lock()
	out, err := ioutil.ReadAll(c.Stdout())
	data.Rows[rIndex].Columns[cIndex+1] = fmt.Sprintf("%s%s", data.Rows[rIndex].Columns[cIndex+1],  strings.TrimSuffix(string(out), "\n"))
	dataMutex.Unlock()
	if err != nil && err != io.EOF {
	    // TODO: io.Copy() should not return io.EOF at all.
	    // Upstream bug? Or prefixer.WriteTo() bug?
	    fmt.Fprintf(os.Stderr, "%v", err)
	}
    }
    wg.Add(1)
    go stdoutHandler(c)

    // Copy over tasks's STDERR.
    var stderrHandler = func(c Client) {
	defer wg.Done()
	dataMutex.Lock()
	out, err := ioutil.ReadAll(c.Stderr())
	data.Rows[rIndex].Columns[cIndex+1] = fmt.Sprintf("%s%s", data.Rows[rIndex].Columns[cIndex+1],  strings.TrimSuffix(string(out), "\n"))
	dataMutex.Unlock()
	if err != nil && err != io.EOF {
	    fmt.Fprintf(os.Stderr, "%v", err)
	}
    }
    wg.Add(1)
    go stderrHandler(c)

    wg.Wait()

    if err := c.Wait(); err != nil {
	os.Exit(1)
    }

    return nil
}

func initSpinner() yacspin.Spinner {
    spinner, err := dao.TaskSpinner()
    core.CheckIfError(err)

    err = spinner.Start()
    core.CheckIfError(err)

    spinner.Message(" Running")

    return spinner
}
