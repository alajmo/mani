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

func (t *Task) RunTask(
	projects []Project,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	if runFlags.Describe {
		PrintTaskBlock([]Task{*t})
	}

	if runFlags.Output != "" {
		t.SpecData.Output = runFlags.Output
	}

	switch t.SpecData.Output {
	case "table", "html", "markdown":
		t.tableTask(projects, userArgs, config, runFlags)
	default: // text
		t.textTask(projects, userArgs, config, runFlags)
	}
}

func (t *Task) tableTask(
	projects []Project,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	t.EnvList = GetEnvList(t.Env, userArgs, []string{}, config.EnvList)

	if runFlags.OmitEmpty {
		t.SpecData.OmitEmpty = true
	}

	if runFlags.Parallel {
		t.SpecData.Parallel = true
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
	data.Headers = append(data.Headers, "Project")

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

	// Append Command name if set
	if t.Cmd != "" {
		data.Headers = append(data.Headers, t.Name)
	}

	for _, project := range projects {
		data.Rows = append(data.Rows, table.Row{project.Name})
	}

	/**
	** Values
	**/
	var wg sync.WaitGroup

	for i, project := range projects {
		wg.Add(1)

		if t.SpecData.Parallel {
			spinner.Message(" Running")
			go t.tableWork(config, &data, project, runFlags.DryRun, i, &wg)
		} else {
			spinner.Message(fmt.Sprintf(" %v", project.Name))
			t.tableWork(config, &data, project, runFlags.DryRun, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	printTable(t.ThemeData.Table, t.SpecData.OmitEmpty, t.SpecData.Output, data)
}

func (t Task) tableWork(
	config *Config,
	data *core.TableOutput,
	project Project,
	dryRunFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for _, cmd := range t.Commands {
		var output string
		var err error
		output, err = RunTable(*config, cmd.Cmd, cmd.EnvList, cmd.Shell, project, dryRunFlag)
		// TODO: Thread safety? Perhaps re-write this
		// TODO: Also, if project path does not exist, no error is shown, which can be confusing
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))

		if err != nil && !t.SpecData.IgnoreError {
			return
		}
	}

	if t.Cmd != "" {
		var output string
		output, _ = RunTable(*config, t.Cmd, t.EnvList, t.Shell, project, dryRunFlag)
		data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
	}
}

func (t *Task) textTask(
	projects []Project,
	userArgs []string,
	config *Config,
	runFlags *core.RunFlags,
) {
	t.EnvList = GetEnvList(t.Env, userArgs, []string{}, config.EnvList)

	if runFlags.Parallel {
		t.SpecData.Parallel = true
	}

	// Set env for sub-commands
	for i := range t.Commands {
		t.Commands[i].EnvList = GetEnvList(t.Commands[i].Env, userArgs, t.EnvList, config.EnvList)
	}

	var wg sync.WaitGroup

	for i, project := range projects {
		wg.Add(1)

		colorIndex := core.COLOR_INDEX[i % len(core.COLOR_INDEX)]
		if t.SpecData.Parallel {
			go t.textWork(uint8(colorIndex), config, project, runFlags.DryRun, &wg)
		} else {
			t.textWork(uint8(colorIndex), config, project, runFlags.DryRun, &wg)
		}
	}

	wg.Wait()
}

// TODO: Update design
func (t Task) textWork(
	colorIndex uint8,
	config *Config,
	project Project,
	dryRunFlag bool,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var header string
	if t.Desc != "" {
		header = fmt.Sprintf("[%s] %s [%s: %s]", color.Index(colorIndex, project.Name), "TASK", color.Bold(t.Name), t.Desc)
	} else {
		header = fmt.Sprintf("[%s] %s [%s]", color.Index(colorIndex, project.Name), "TASK", t.Name)
	}

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)
	headerLength := len(core.Strip(header))
	// separators := strings.Repeat("=", headerLength)
	header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat("*", width - headerLength - 1))
	fmt.Println(header)

	for i, cmd := range t.Commands {
		var header string
		if cmd.Desc != "" {
			header = fmt.Sprintf("%s %d/%d [%s: %s]", "TASK", i+1, len(t.Commands), color.Bold(cmd.Name), cmd.Desc)
		} else {
			header = fmt.Sprintf("%s %d/%d [%s]", "TASK", i+1, len(t.Commands), color.Bold(cmd.Name))
		}

		// separators := strings.Repeat("*", len(core.Strip(header)))
		headerLength := len(core.Strip(header))
		header = fmt.Sprintf("%s %s", header, strings.Repeat("*", width - headerLength -1 ))
		fmt.Println(header)

		err := RunText(cmd.Cmd, cmd.EnvList, *config, cmd.Shell, project, dryRunFlag)

		if err != nil && !t.SpecData.IgnoreError {
			return
		}
		fmt.Println()
	}

	if t.Cmd != "" {
		err = RunText(t.Cmd, t.EnvList, *config, t.Shell, project, dryRunFlag)
		core.CheckIfError(err)
	}
}
