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


func (exec *Exec) Table(dryRun bool) dao.TableOutput {
	task := exec.Tasks[0]
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
				exec.TableWork(i, dryRun, data, &dataMutex)
			}(i, c)
		} else {
			func(i int, c Client) {
				defer wg.Done()
				exec.TableWork(i, dryRun, data, &dataMutex)
			}(i, c)
		}
	}
	wg.Wait()

	err := spinner.Stop()
	core.CheckIfError(err)

	return data
}

func (exec *Exec) TableWork(rIndex int, dryRun bool, data dao.TableOutput, dataMutex *sync.RWMutex) {
	client := exec.Clients[rIndex]
	task := exec.Tasks[rIndex]
	var wg sync.WaitGroup

	for j, cmd := range task.Commands {
		err := RunTableCmd(rIndex, j + 1, client, dryRun, cmd.ShellProgram, cmd.EnvList, cmd.Cmd, cmd.CmdArg, data, dataMutex, &wg)
		if err != nil && !task.SpecData.IgnoreError {
			return
		}
	}

	if task.Cmd != "" {
		_ = RunTableCmd(rIndex, len(task.Commands) + 1, client, dryRun, task.ShellProgram, task.EnvList, task.Cmd, task.CmdArg, data, dataMutex, &wg)
	}

	wg.Wait()
}

func RunTableCmd(
	rIndex int,
	cIndex int,
	c Client,
	dryRun bool,
	shell string,
	env []string,
	cmd string,
	cmdArr []string,
	data dao.TableOutput,
	dataMutex *sync.RWMutex,
	wg *sync.WaitGroup,
) error {
	combinedEnvs := core.MergeEnvs(c.Env, env)

	if dryRun {
		data.Rows[rIndex].Columns[cIndex] = cmd
		return nil
	}

	err := c.Run(shell, combinedEnvs, cmdArr)
	if err != nil {
		return err
	}

	// Copy over commands STDOUT.
	var stdoutHandler = func(c Client) {
		defer wg.Done()
		dataMutex.Lock()
		out, err := ioutil.ReadAll(c.Stdout())
		data.Rows[rIndex].Columns[cIndex] = fmt.Sprintf("%s%s", data.Rows[rIndex].Columns[cIndex],  strings.TrimSuffix(string(out), "\n"))
		dataMutex.Unlock()
		if err != nil && err != io.EOF {
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
		data.Rows[rIndex].Columns[cIndex] = fmt.Sprintf("%s%s", data.Rows[rIndex].Columns[cIndex],  strings.TrimSuffix(string(out), "\n"))
		dataMutex.Unlock()
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	wg.Add(1)
	go stderrHandler(c)

	wg.Wait()

	if err := c.Wait(); err != nil {
		data.Rows[rIndex].Columns[cIndex] = fmt.Sprintf("%s\n%s", data.Rows[rIndex].Columns[cIndex], err.Error())
		return err
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
