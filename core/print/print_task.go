package print

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

type ListTaskFlags struct {
	Headers []string
}

func PrintTasks(
	tasks []dao.Task,
	listFlags ListFlags,
	taskFlags ListTaskFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	var headers []interface{}
	for _, h := range taskFlags.Headers {
		headers = append(headers, h)
	}

	if !listFlags.NoHeaders {
		t.AppendHeader(headers)
	}

	for _, task := range tasks {
		var row []interface{}
		for _, h := range headers {
			value := task.GetValue(fmt.Sprintf("%v", h))
			row = append(row, value)
		}

		t.AppendRow(row)
	}

	if listFlags.NoBorders {
		t.Style().Box = core.StyleNoBorders
		t.Style().Options.SeparateHeader = false
		t.Style().Options.DrawBorder = false
	}

	switch listFlags.Output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}

func PrintTaskBlock(tasks []dao.Task) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	for _, task := range tasks {
		t.AppendRows([]table.Row{
			{"Name: ", task.Name},
			{"Description: ", task.Description},
			{"Env: ", printEnv(task.EnvList)},
		})

		if task.Command != "" {
			t.AppendRow(table.Row{"Command: ", task.Command})
		}

		if len(task.Commands) > 0 {
			t.AppendRow(table.Row{"Commands:"})
			for _, subCommand := range task.Commands {
				t.AppendRows([]table.Row{
					{" - Name: ", subCommand.Name},
					{"   Description: ", subCommand.Description},
					{"   Env: ", printEnv(subCommand.EnvList)},
					{"   Command: ", subCommand.Command},
				})
				t.AppendRow(table.Row{})
				t.AppendSeparator()
			}
		}

		t.AppendSeparator()
		t.AppendRow(table.Row{})
		t.AppendSeparator()
	}

	t.Style().Box = core.StyleNoBorders
	t.Style().Options.SeparateHeader = false
	t.Style().Options.DrawBorder = false

	t.Render()
}

func printEnv(env []string) string {
	var str string = ""
	var i int = 0
	for _, env := range env {
		str = fmt.Sprintf("%s%s", str, strings.TrimSuffix(env, "\n"))

		if i < len(env)-1 {
			str = str + "\n"
		}

		i += 1
	}

	return strings.TrimSuffix(str, "\n")
}
