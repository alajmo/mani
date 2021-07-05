package print

import (
	"github.com/alajmo/mani/core"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

func PrintTags(
	tags []string, 
	listFlags core.ListFlags, 
	tagFlags core.ListTagFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	var headers[]interface{}
	for _, h := range tagFlags.Headers {
		headers = append(headers, h)
	}

	if (!listFlags.NoHeaders) {
		t.AppendHeader(headers)
	}

	for _, tag := range tags {
		var row[]interface{}
		row = append(row, tag)

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
