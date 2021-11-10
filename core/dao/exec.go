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
	case "table", "html", "markdown" :
		tableExec(cmd, entityList, config, runFlags)
	default: // text
		textExec(cmd, entityList, config, runFlags)
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

func textExec(
	cmd string,
	entityList EntityList,
	config *Config,
	runFlags *core.RunFlags,
) {
	var wg sync.WaitGroup

	for i, entity := range entityList.Entities {
		colorIndex := core.COLOR_INDEX[i % len(core.COLOR_INDEX)]

		wg.Add(1)
		if runFlags.Parallel {
			go textWork(uint8(colorIndex), config, cmd, entity, runFlags.DryRun, &wg)
		} else {
			textWork(uint8(colorIndex), config, cmd, entity, runFlags.DryRun, &wg)
		}
	}

	wg.Wait()
}

func textWork(
	colorIndex uint8,
	config *Config,
	cmd string,
	entity Entity,
	dryRunFlag bool,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	header := fmt.Sprintf("[%s]", color.Index(colorIndex, entity.Name))

	width, _, err := term.GetSize(0)
	core.CheckIfError(err)
	headerLength := len(core.Strip(header))
	header = fmt.Sprintf("\n%s %s\n", header, strings.Repeat("*", width - headerLength - 1))
	fmt.Println(header)

	err = RunText(cmd, []string{}, *config, config.Shell, entity, dryRunFlag)
	core.CheckIfError(err)
}
