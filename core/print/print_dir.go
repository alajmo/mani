package print

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

type ListDirFlags struct {
	Tags     []string
	DirPaths []string
	Headers  []string
}

func PrintDirs(
	dirs []dao.Dir,
	listFlags ListFlags,
	dirFlags ListDirFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	var headers []interface{}
	for _, h := range dirFlags.Headers {
		headers = append(headers, h)
	}

	if !listFlags.NoHeaders {
		t.AppendHeader(headers)
	}

	for _, dir := range dirs {
		var row []interface{}
		for _, h := range headers {
			value := dir.GetValue(fmt.Sprintf("%v", h))
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

func PrintDirBlocks(dirs []dao.Dir) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	for _, dir := range dirs {
		t.AppendRows([]table.Row{
			{"Name: ", dir.Name},
			{"Path: ", dir.RelPath},
			{"Description: ", dir.Description},
			{"Tags: ", dir.GetValue("Tags")},
		})

		t.AppendSeparator()
		t.AppendRow(table.Row{})
		t.AppendSeparator()
	}

	t.Style().Box = core.StyleNoBorders
	t.Style().Options.SeparateHeader = false
	t.Style().Options.DrawBorder = false

	t.Render()
}
