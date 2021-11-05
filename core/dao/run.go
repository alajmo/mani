package dao

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/term"
	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"

	core "github.com/alajmo/mani/core"
)

var COLOR_INDEX = []int {2, 32, 179, 63, 148, 205}

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
	case "table", "html", "markdown":
		t.tableTask(entityList, userArgs, config, runFlags)
	default: // text
		t.textTask(entityList, userArgs, config, runFlags)
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
	if t.Cmd != "" {
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
		output, err = RunTable(*config, cmd.Cmd, cmd.EnvList, cmd.Shell, entity, dryRunFlag)
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))

		if err != nil && !t.IgnoreError {
			return
		}
	}

	if t.Cmd != "" {
		var output string
		output, _ = RunTable(*config, t.Cmd, t.EnvList, t.Shell, entity, dryRunFlag)
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
	}
}

func (t *Task) textTask(
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

	// for i := uint8(16); i <= 231; i++ {
	// 	fmt.Println(i, color.Index(i, "pew-pew"))
	// }

	for i, entity := range entityList.Entities {
		wg.Add(1)

		colorIndex := COLOR_INDEX[i % len(COLOR_INDEX)]
		if t.Parallel {
			go t.textWork(uint8(colorIndex), config, entity, runFlags.DryRun, &wg)
		} else {
			t.textWork(uint8(colorIndex), config, entity, runFlags.DryRun, &wg)
		}
	}

	wg.Wait()
}

// TODO: Update design
func (t Task) textWork(
	colorIndex uint8,
	config *Config,
	entity Entity,
	dryRunFlag bool,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var header string
	if t.Desc != "" {
		header = fmt.Sprintf("[%s] %s [%s: %s]", color.Index(colorIndex, entity.Name), "TASK", color.Bold(t.Name), t.Desc)
	} else {
		header = fmt.Sprintf("[%s] %s [%s]", color.Index(colorIndex, entity.Name), "TASK", t.Name)
	}

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)

	headerLength := len(core.Strip(header))

	// separators := strings.Repeat("=", headerLength)
	header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat("-", width - headerLength - 1))
	fmt.Println(header)

	for i, cmd := range t.Commands {
		var header string
		if cmd.Desc != "" {
			header = fmt.Sprintf("[%s] %s %d/%d [%s: %s]", color.Index(colorIndex, entity.Name), "TASK", i+1, len(t.Commands), color.Bold(cmd.Name), cmd.Desc)
		} else {
			header = fmt.Sprintf("[%s] %s %d/%d [%s]", color.Index(colorIndex, entity.Name), "TASK", i+1, len(t.Commands), color.Bold(cmd.Name))
		}

		// separators := strings.Repeat("-", len(core.Strip(header)))
		headerLength := len(core.Strip(header))
		header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat("-", width - headerLength -1 ))
		fmt.Println(header)

		err := RunText(cmd.Cmd, cmd.EnvList, *config, cmd.Shell, entity, dryRunFlag)

		if err != nil && !t.IgnoreError {
			return
		}
		fmt.Println()
	}

	if t.Cmd != "" {
		RunText(t.Cmd, t.EnvList, *config, t.Shell, entity, dryRunFlag)
	}
}
