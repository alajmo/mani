package print

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"strings"

	"github.com/alajmo/mani/core/dao"
)

type ListCommandFlags struct {
	Headers []string
}

func PrintCommands(
	commands []dao.Command,
	listFlags ListFlags,
	commandFlags ListCommandFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(ManiList)

	var headers[]interface{}
	for _, h := range commandFlags.Headers {
		headers = append(headers, h)
	}

	if (!listFlags.NoHeaders) {
		t.AppendHeader(headers)
	}

	for _, command := range commands {
		var row[]interface{}
		for _, h := range headers {
			value := command.GetValue(fmt.Sprintf("%v", h))
			row = append(row, value)
		}

		t.AppendRow(row)
	}

	if (listFlags.NoBorders) {
		t.Style().Box = StyleNoBorders
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

func PrintCommandBlocks(commands []dao.Command) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(ManiList)

	for _, command := range commands {
		t.AppendRows([] table.Row {
			{ "Name: ", command.Name },
			{ "Description: ", command.Description },
			{ "Shell: ", command.Shell },
			{ "Env: ", printEnv(command.EnvList) },
		})

		if (command.Command != "") {
			t.AppendRow(table.Row { "Command: ", command.Command })
		}

		if len(command.Commands) > 0 {
			t.AppendRow(table.Row{ "Commands:" })
			for _, subCommand := range command.Commands {
				t.AppendRows([] table.Row {
					{ " - Name: ", subCommand.Name },
					{ "   Description: ", subCommand.Description },
					{ "   Shell: ", subCommand.Shell },
					{ "   Env: ", printEnv(subCommand.EnvList) },
					{ "   Command: ", subCommand.Command },
				})
				t.AppendRow(table.Row{})
				t.AppendSeparator()
			}
		}

		t.AppendSeparator()
		t.AppendRow(table.Row{})
		t.AppendSeparator()
	}

	t.Style().Box = StyleNoBorders
	t.Style().Options.SeparateHeader = false
	t.Style().Options.DrawBorder = false

	t.Render()
}

func printEnv(env []string) string {
	var str string = ""
	var i int = 0
	for _, env := range env {
		str = fmt.Sprintf("%s%s", str, strings.TrimSuffix(env, "\n"))

		if (i  < len(env) - 1) {
			str = str + "\n"
		}

		i += 1
	}

	return strings.TrimSuffix(str, "\n")
}
