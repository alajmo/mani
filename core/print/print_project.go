package print

import (
	"github.com/alajmo/mani/core"
	"fmt"
	tabby "github.com/cheynewallace/tabby"
	"github.com/jedib0t/go-pretty/v6/table"
	"path/filepath"
	"strings"
	"os"
)

func PrintProjects(
	configPath string,
	projects []core.Project,
	listFlags core.ListFlags,
	projectFlags core.ListProjectFlags,
) {

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	var headers[]interface{}
	for _, h := range projectFlags.Headers {
		headers = append(headers, strings.Title(h))
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

func PrintProjectBlocks(configPath string, projects []core.Project) {
	baseDir := filepath.Dir(configPath)
	t := tabby.New()
	for _, project := range projects {
		relPath, err := filepath.Rel(baseDir, project.Path)
		core.CheckIfError(err)

		t.AddLine("Name:", project.Name)
		t.AddLine("Path:", relPath)
		t.AddLine("Description:", project.Description)
		t.AddLine("Url:", project.Url)
		t.AddLine("Tags:", strings.Join(project.Tags, ", "))
		t.AddLine("")
		t.AddLine("")
	}

	t.Print()
}
