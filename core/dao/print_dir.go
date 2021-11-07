package dao

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"

	"github.com/alajmo/mani/core"
)

func PrintDirs(
	config *Config,
	dirs []Dir,
	listFlags core.ListFlags,
	dirFlags core.DirFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	// Table Style
	switch theme.Table {
	case "ascii":
		core.ManiList.Box = core.StyleBoxASCII
	default:
		core.ManiList.Box = core.StyleBoxLight
	}

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

func PrintDirBlocks(dirs []Dir) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	for _, dir := range dirs {
		t.AppendRows([]table.Row{
			{"Name: ", dir.Name},
			{"Path: ", dir.RelPath},
			{"Desc: ", dir.Desc},
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
