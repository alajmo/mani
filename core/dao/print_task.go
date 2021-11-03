package dao

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
)

func PrintTasks(
	config *Config,
	tasks []Task,
	listFlags core.ListFlags,
	taskFlags core.TaskFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	// Table Style
	switch theme.Table {
	case "ascii":
		core.ManiList.Box = core.StyleBoxASCII
	default:
		core.ManiList.Box = core.StyleBoxDefault
	}

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

func PrintTaskBlock(tasks []Task) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	for _, task := range tasks {
		t.AppendRows([]table.Row{
			{"Name: ", task.Name},
			{"Desc: ", task.Desc},
			{"Target: ", printTarget(task.Target)},
			{"Env: ", printEnv(task.EnvList)},
			{"Parallel: ", task.Parallel},
			{"Abort: ", task.Abort},
		})

		if task.Command != "" {
			t.AppendRow(table.Row{"Command: ", task.Command})
		}

		if len(task.Commands) > 0 {
			t.AppendRow(table.Row{"Commands:"})
			for _, subCommand := range task.Commands {
				t.AppendRows([]table.Row{
					{" - Name: ", subCommand.Name},
					{"   Desc: ", subCommand.Desc},
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

func printTarget(target Target) string {
	var str string = ""

	if len(target.Projects) > 0 {
		str = fmt.Sprintf("%sProjects: %s\n", str, strings.Join(target.Projects, ", "))
	}

	if len(target.Dirs) > 0 {
		str = fmt.Sprintf("%sDirs: %s\n", str, strings.Join(target.Dirs, ", "))
	}

	if len(target.Paths) > 0 {
		str = fmt.Sprintf("%sPaths: %s\n", str, strings.Join(target.Paths, ", "))
	}

	if len(target.Tags) > 0 {
		str = fmt.Sprintf("%sTags: %s", str, strings.Join(target.Tags, ", "))
	}

	if len(str) > 0 {
		str = fmt.Sprintf("\n%s", str)
	}

	return str
}
