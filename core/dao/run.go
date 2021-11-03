package dao

import (
	"fmt"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"
	"golang.org/x/term"

	core "github.com/alajmo/mani/core"
)

func (t *Task) RunTask(
	entityList EntityList,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	if runFlags.Describe {
		PrintTaskBlock([]Task{*t})
	}

	if runFlags.Output != "" {
		t.Output = runFlags.Output
	}

	switch t.Output {
	case "table", "markdown", "html":
		t.tableTask(entityList, userArgs, config, runFlags)
	default:
		t.lineTask(entityList, userArgs, config, runFlags)
	}
}

func (t *Task) tableTask(
	entityList EntityList,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	t.EnvList = GetEnvList(t.Env, userArgs, []string{}, config.EnvList)

	if runFlags.Parallel {
		t.Parallel = true
	}

	// Set env for sub-commands
	for i := range t.Commands {
		t.Commands[i].EnvList = GetEnvList(t.Commands[i].Env, userArgs, t.EnvList, config.EnvList)
	}

	spinner, err := TaskSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	core.CheckIfError(err)

	var data core.TableOutput

	/**
	** Headers
	**/
	data.Headers = append(data.Headers, entityList.Type)

	// Append Command name if set
	if t.Command != "" {
		data.Headers = append(data.Headers, t.Name)
	}

	// Append Command names if set
	for _, cmd := range t.Commands {
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

	for _, entity := range entityList.Entities {
		data.Rows = append(data.Rows, table.Row{entity.Name})
	}

	/**
	** Values
	**/
	var wg sync.WaitGroup

	for i, entity := range entityList.Entities {
		wg.Add(1)

		if t.Parallel {
			spinner.Message(" Running")
			go t.tableWork(config, &data, entity, runFlags.DryRun, i, &wg)
		} else {
			spinner.Message(fmt.Sprintf(" %v", entity.Name))
			t.tableWork(config, &data, entity, runFlags.DryRun, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	printTable(t.ThemeData.Table, t.Output, data)
}

func (t Task) tableWork(
	config *Config,
	data *core.TableOutput,
	entity Entity,
	dryRunFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for _, cmd := range t.Commands {
		var output string
		var err error
		output, err = RunTable(*config, cmd.Command, cmd.EnvList, cmd.Shell, entity, dryRunFlag)
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))

		if err != nil && t.Abort {
			return
		}
	}

	if t.Command != "" {
		var output string
		output, _ = RunTable(*config, t.Command, t.EnvList, t.Shell, entity, dryRunFlag)
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
	}
}

func (t *Task) lineTask(
	entityList EntityList,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	t.EnvList = GetEnvList(t.Env, userArgs, []string{}, config.EnvList)

	if runFlags.Parallel {
		t.Parallel = true
	}

	// Set env for sub-commands
	for i := range t.Commands {
		t.Commands[i].EnvList = GetEnvList(t.Commands[i].Env, userArgs, t.EnvList, config.EnvList)
	}

	var wg sync.WaitGroup

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)
	var header string
	if t.Desc != "" {
		header = fmt.Sprintf("%s [%s: %s]", color.Bold("TASK"), t.Name, t.Desc)
	} else {
		header = fmt.Sprintf("%s [%s]", color.Bold("TASK"), t.Name)
	}

	fmt.Printf("\n%s %s\n", header, strings.Repeat("*", width-len(header)-1))

	maxNameLength := entityList.GetLongestNameLength()

	for _, entity := range entityList.Entities {
		wg.Add(1)
		if t.Parallel {
			go t.lineWork(config, entity, runFlags.DryRun, maxNameLength, &wg)
		} else {
			t.lineWork(config, entity, runFlags.DryRun, maxNameLength, &wg)
		}
	}

	wg.Wait()
}

func (t Task) lineWork(
	config *Config,
	entity Entity,
	dryRunFlag bool,
	maxNameLength int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for i, cmd := range t.Commands {
		var header string
		if cmd.Desc != "" {
			header = fmt.Sprintf("TASK %d/%d [%s: %s]", i+1, len(t.Commands), cmd.Name, cmd.Desc)
		} else {
			header = fmt.Sprintf("TASK %d/%d [%s]", i+1, len(t.Commands), cmd.Name)
		}

		fmt.Println(header)
		err := RunList(cmd.Command, cmd.EnvList, *config, cmd.Shell, entity, dryRunFlag, maxNameLength)

		if err != nil && t.Abort {
			return
		}
		fmt.Println()
	}

	if t.Command != "" {
		RunList(t.Command, t.EnvList, *config, t.Shell, entity, dryRunFlag, maxNameLength)
	}
}
