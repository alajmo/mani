package dao

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"

	"github.com/alajmo/mani/core"
)

func PrintProjects(
	config *Config,
	projects []Project,
	listFlags core.ListFlags,
	projectFlags core.ProjectFlags,
) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	// Table Style
	switch theme.Table {
	case "ascii":
		core.ManiList.Box = core.StyleBoxASCII
	default: // light
		core.ManiList.Box = core.StyleBoxLight
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	var headers []interface{}
	for _, h := range projectFlags.Headers {
		headers = append(headers, h)
	}

	if !listFlags.NoHeaders {
		t.AppendHeader(headers)
	}

	for _, project := range projects {
		var row []interface{}
		for _, h := range headers {
			value := project.GetValue(fmt.Sprintf("%v", h))
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
	default: // text
		t.Render()
	}
}

func PrintProjectBlocks(projects []Project) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	for _, project := range projects {
		t.AppendRows([]table.Row{
			{"Name: ", project.Name},
			{"Path: ", project.RelPath},
			{"Description: ", project.Desc},
			{"Url: ", project.Url},
			{"Tags: ", project.GetValue("Tags")},
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
