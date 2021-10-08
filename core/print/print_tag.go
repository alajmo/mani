package print

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
)

type ListTagFlags struct {
	Headers []string
}

func PrintTags(
	tags []string,
	listFlags ListFlags,
	tagFlags ListTagFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	var headers []interface{}
	for _, h := range tagFlags.Headers {
		headers = append(headers, h)
	}

	if !listFlags.NoHeaders {
		t.AppendHeader(headers)
	}

	for _, tag := range tags {
		var row []interface{}
		row = append(row, tag)

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
