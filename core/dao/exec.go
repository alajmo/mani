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

func RunExec(
	cmd string,
	projects []Project,
	config *Config,
	runFlags *core.RunFlags,
) {
	switch runFlags.Output {
	case "table", "html", "markdown" :
		tableExec(cmd, projects, config, runFlags)
	default: // text
		textExec(cmd, projects, config, runFlags)
	}
}

func tableExec(
	cmd string,
	projects []Project,
	config *Config,
	runFlags *core.RunFlags,
) {
	spinner, err := TaskSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	core.CheckIfError(err)

	var data core.TableOutput

	/**
	** Headers
	**/
	data.Headers = append(data.Headers, "Project")

	// Append Command name if set
	data.Headers = append(data.Headers, "Output")

	for _, project := range projects {
		data.Rows = append(data.Rows, table.Row{project.Name})
	}

	/**
	** Values
	**/
	var wg sync.WaitGroup

	for i, project := range projects {
		wg.Add(1)

		if runFlags.Parallel {
			spinner.Message(" Running")
			go tableWork(config, &data, cmd, project, runFlags.DryRun, i, &wg)
		} else {
			spinner.Message(fmt.Sprintf(" %v", project.Name))
			tableWork(config, &data, cmd, project, runFlags.DryRun, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	theme, err := config.GetTheme("default")
	core.CheckIfError(err)

	printTable(theme.Table, runFlags.OmitEmpty, runFlags.Output, data)
}

func tableWork(
	config *Config,
	data *core.TableOutput,
	cmd string,
	project Project,
	dryRunFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var output string
	output, _ = RunTable(*config, cmd, []string{}, config.Shell, project, dryRunFlag)
	data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
}

func textExec(
	cmd string,
	projects []Project,
	config *Config,
	runFlags *core.RunFlags,
) {
	var wg sync.WaitGroup

	for i, project := range projects {
		colorIndex := core.COLOR_INDEX[i % len(core.COLOR_INDEX)]

		wg.Add(1)
		if runFlags.Parallel {
			go textWork(uint8(colorIndex), config, cmd, project, runFlags.DryRun, &wg)
		} else {
			textWork(uint8(colorIndex), config, cmd, project, runFlags.DryRun, &wg)
		}
	}

	wg.Wait()
}

func textWork(
	colorIndex uint8,
	config *Config,
	cmd string,
	project Project,
	dryRunFlag bool,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	header := fmt.Sprintf("[%s]", color.Index(colorIndex, project.Name))

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)
	headerLength := len(core.Strip(header))
	header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat("*", width - headerLength - 1))
	fmt.Println(header)

	err = RunText(cmd, []string{}, *config, config.Shell, project, dryRunFlag)
	core.CheckIfError(err)
}
