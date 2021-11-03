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
	entityList EntityList,
	config *Config,
	runFlags *core.RunFlags,
) {
	switch runFlags.Output {
	case "table", "markdown", "html":
		tableExec(cmd, entityList, config, runFlags)
	default:
		lineExec(cmd, entityList, config, runFlags)
	}
}

func tableExec(
	cmd string,
	entityList EntityList,
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
	data.Headers = append(data.Headers, entityList.Type)

	// Append Command name if set
	data.Headers = append(data.Headers, "Output")

	for _, entity := range entityList.Entities {
		data.Rows = append(data.Rows, table.Row{entity.Name})
	}

	/**
	** Values
	**/
	var wg sync.WaitGroup

	for i, entity := range entityList.Entities {
		wg.Add(1)

		if runFlags.Parallel {
			spinner.Message(" Running")
			go tableWork(config, &data, cmd, entity, runFlags.DryRun, i, &wg)
		} else {
			spinner.Message(fmt.Sprintf(" %v", entity.Name))
			tableWork(config, &data, cmd, entity, runFlags.DryRun, i, &wg)
		}
	}

	wg.Wait()

	err = spinner.Stop()
	core.CheckIfError(err)

	theme, err := config.GetTheme("default")
	core.CheckIfError(err)

	printTable(theme.Table, runFlags.Output, data)
}

func tableWork(
	config *Config,
	data *core.TableOutput,
	cmd string,
	entity Entity,
	dryRunFlag bool,
	i int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var output string
	output, _ = RunTable(*config, cmd, []string{}, config.Shell, entity, dryRunFlag)
	data.Rows[i] = append(data.Rows[i], strings.TrimSuffix(output, "\n"))
}

func lineExec(
	cmd string,
	entityList EntityList,
	config *Config,
	runFlags *core.RunFlags,
) {
	var wg sync.WaitGroup

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)
	var header = fmt.Sprintf("%s [%s]", color.Bold("TASK"), "Output")

	fmt.Printf("\n%s %s\n", header, strings.Repeat("*", width-len(header)-1))
	maxNameLength := entityList.GetLongestNameLength()

	for _, entity := range entityList.Entities {
		wg.Add(1)
		if runFlags.Parallel {
			go lineWork(config, cmd, entity, runFlags.DryRun, maxNameLength, &wg)
		} else {
			lineWork(config, cmd, entity, runFlags.DryRun, maxNameLength, &wg)
		}
	}

	wg.Wait()
}

func lineWork(
	config *Config,
	cmd string,
	entity Entity,
	dryRunFlag bool,
	maxNameLength int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	RunList(cmd, []string{}, *config, config.Shell, entity, dryRunFlag, maxNameLength)
}
