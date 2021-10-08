package print

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

type ListProjectFlags struct {
	Tags         []string
	ProjectPaths []string
	Headers      []string
}

func PrintProjects(
	projects []dao.Project,
	listFlags ListFlags,
	projectFlags ListProjectFlags,
) {
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
	default:
		t.Render()
	}
}

func PrintProjectBlocks(projects []dao.Project) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	for _, project := range projects {
		t.AppendRows([]table.Row{
			{"Name: ", project.Name},
			{"Path: ", project.RelPath},
			{"Description: ", project.Description},
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
