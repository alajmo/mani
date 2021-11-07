package dao

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
)

func PrintTags(
	config *Config,
	keys []string,
	tags map[string]TagAssocations,
	listFlags core.ListFlags,
	tagFlags core.TagFlags,
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
	for _, h := range tagFlags.Headers {
		headers = append(headers, h)
	}

	if !listFlags.NoHeaders {
		t.AppendHeader(headers)
	}

	for _, key := range keys {
		var row []interface{}
		for _, h := range headers {
			value := tags[key].GetValue(fmt.Sprintf("%v", h))
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
