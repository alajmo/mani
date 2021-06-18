package print

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"

	"github.com/alajmo/mani/core/dao"
)

type ListProjectFlags struct {
	Tags []string
	Dirs []string
	Headers []string
}

func PrintProjects(
	projects []dao.Project,
	listFlags ListFlags,
	projectFlags ListProjectFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(ManiList)

	var headers[]interface{}
	for _, h := range projectFlags.Headers {
		headers = append(headers, h)
	}

	if (!listFlags.NoHeaders) {
		t.AppendHeader(headers)
	}

	for _, project := range projects {
		var row[]interface{}
		for _, h := range headers {
			value := project.GetValue(fmt.Sprintf("%v", h))
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

func PrintProjectBlocks(projects []dao.Project) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(ManiList)

	for _, project := range projects {
		t.AppendRows([] table.Row {
			{ "Name: ", project.Name },
			{ "Path: ", project.RelPath },
			{ "Description: ", project.Description },
			{ "Url: ", project.Url },
			{ "Tags: ", project.GetValue("Tags") },
		})

		t.AppendSeparator()
		t.AppendRow(table.Row{})
		t.AppendSeparator()
	}

	t.Style().Box = StyleNoBorders
	t.Style().Options.SeparateHeader = false
	t.Style().Options.DrawBorder = false

	t.Render()
}
