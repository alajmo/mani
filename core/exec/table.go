package exec

import (
	"fmt"
	"sync"
	"io"
	"io/ioutil"
	"strings"
	"os"
	"os/signal"

	"github.com/theckman/yacspin"

	dao "github.com/alajmo/mani/core/dao"
)


func (exec *Exec) Table(dryRun bool) (dao.TableOutput) {
	task := exec.Tasks[0]
	clients := exec.Clients
	projects := exec.Projects

	spinner, spinnerErr := initSpinner()

	// In-case user interrupts, make sure spinner is stopped
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		if spinnerErr == nil {
			_ = spinner.Stop()
		}
		os.Exit(0)
	}()

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

	if spinnerErr == nil {
		_ = spinner.Stop()
	}

	return data
}

func (exec *Exec) TableWork(rIndex int, dryRun bool, data dao.TableOutput, dataMutex *sync.RWMutex) {
	client := exec.Clients[rIndex]
	task := exec.Tasks[rIndex]
	var wg sync.WaitGroup

	for j, cmd := range task.Commands {
		args := TableCmd {
			rIndex: rIndex,
			cIndex: j + 1,
			client: client,
			dryRun: dryRun,
			shell: cmd.ShellProgram,
			env: cmd.EnvList,
			cmd: cmd.Cmd,
			cmdArr: cmd.CmdArg,
		}

		err := RunTableCmd(args, data, dataMutex, &wg)
		if err != nil && !task.SpecData.IgnoreError {
			return
		}
	}

	if task.Cmd != "" {
		args := TableCmd {
			rIndex: rIndex,
			cIndex: len(task.Commands) + 1,
			client: client,
			dryRun: dryRun,
			shell: task.ShellProgram,
			env: task.EnvList,
			cmd: task.Cmd,
			cmdArr: task.CmdArg,
		}

		err := RunTableCmd(args, data, dataMutex, &wg)
		if err != nil && !task.SpecData.IgnoreError {
			return
		}
	}

	wg.Wait()
}

func RunTableCmd(t TableCmd, data dao.TableOutput, dataMutex *sync.RWMutex, wg *sync.WaitGroup) error {
	combinedEnvs := dao.MergeEnvs(t.client.Env, t.env)

	if t.dryRun {
		data.Rows[t.rIndex].Columns[t.cIndex] = t.cmd
		return nil
	}

	err := t.client.Run(t.shell, combinedEnvs, t.cmdArr)
	if err != nil {
		return err
	}

	// Copy over commands STDOUT.
	var stdoutHandler = func(client Client) {
		defer wg.Done()
		dataMutex.Lock()
		out, err := ioutil.ReadAll(client.Stdout())
		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s%s", data.Rows[t.rIndex].Columns[t.cIndex],  strings.TrimSuffix(string(out), "\n"))
		dataMutex.Unlock()

		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	wg.Add(1)
	go stdoutHandler(t.client)

	// Copy over tasks's STDERR.
	var stderrHandler = func(client Client) {
		defer wg.Done()
		dataMutex.Lock()
		out, err := ioutil.ReadAll(client.Stderr())
		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s%s", data.Rows[t.rIndex].Columns[t.cIndex],  strings.TrimSuffix(string(out), "\n"))
		dataMutex.Unlock()
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "%v", err)
		}
	}
	wg.Add(1)
	go stderrHandler(t.client)

	wg.Wait()

	if err := t.client.Wait(); err != nil {
		data.Rows[t.rIndex].Columns[t.cIndex] = fmt.Sprintf("%s\n%s", data.Rows[t.rIndex].Columns[t.cIndex], err.Error())
		return err
	}

	return nil
}

func initSpinner() (yacspin.Spinner, error) {
	spinner, err := dao.TaskSpinner()
	if err != nil {
		return spinner, err
	}

	err = spinner.Start()
	if err != nil {
		return spinner, err
	}

	spinner.Message(" Running")

	return spinner, nil
}
