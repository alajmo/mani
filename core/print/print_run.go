package print

import (
	"github.com/alajmo/mani/core"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

func PrintRun(
	commands []core.Command,
	listFlags core.ListFlags,
	commandFlags core.ListCommandFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

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
		t.Style().Box = core.StyleNoBorders
		t.Style().Options.SeparateHeader = false
		t.Style().Options.DrawBorder = false
	}

	switch listFlags.Format {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}
