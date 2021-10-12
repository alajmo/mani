package print

import (
	"os"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func PrintTags(
	tags map[string]dao.TagAssocations,
	listFlags core.ListFlags,
	tagFlags core.TagFlags,
) {
	// Table Style
	// switch config.Theme.Table {
	// case "ascii":
	// 	core.ManiList.Box = core.StyleBoxASCII
	// default:
	// 	core.ManiList.Box = core.StyleBoxDefault
	// }

	core.ManiList.Box = core.StyleBoxASCII

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

	for _, data := range tags {
		var row []interface{}
		for _, h := range headers {
			value := data.GetValue(fmt.Sprintf("%v", h))
			row = append(row, value)
		}

		row = append(row)

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
