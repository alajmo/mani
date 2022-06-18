package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core/dao"
)

func CreateTable(
	options PrintTableOptions,
	defaultHeaders []string,
	taskHeaders []string,
) table.Writer {
	t := table.NewWriter()

	theme := options.Theme

	t.SetOutputMirror(os.Stdout)
	t.SetStyle(FormatTable(theme))
	if options.SuppressEmptyColumns {
		t.SuppressEmptyColumns()
	}

	headerStyles := make(map[string]table.ColumnConfig)
	for _, h := range defaultHeaders {
		switch h {
		case "project":
			headerStyles[h] = table.ColumnConfig{
				Name: "project",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Project.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Project.Fg, theme.Table.Color.Header.Project.Bg, theme.Table.Color.Header.Project.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Project.Align),
				Colors: combineColors(theme.Table.Color.Row.Project.Fg, theme.Table.Color.Row.Project.Bg, theme.Table.Color.Row.Project.Attr),
			}
		case "synced":
			headerStyles[h] = table.ColumnConfig{
				Name: "synced",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Synced.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Synced.Fg, theme.Table.Color.Header.Synced.Bg, theme.Table.Color.Header.Synced.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Synced.Align),
				Colors: combineColors(theme.Table.Color.Row.Synced.Fg, theme.Table.Color.Row.Synced.Bg, theme.Table.Color.Row.Synced.Attr),
			}
		case "tag":
			headerStyles[h] = table.ColumnConfig{
				Name: "tag",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Tag.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Tag.Fg, theme.Table.Color.Header.Tag.Bg, theme.Table.Color.Header.Tag.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Tag.Align),
				Colors: combineColors(theme.Table.Color.Row.Tag.Fg, theme.Table.Color.Row.Tag.Bg, theme.Table.Color.Row.Tag.Attr),
			}
		case "description":
			headerStyles[h] = table.ColumnConfig{
				Name: "description",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Desc.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Desc.Fg, theme.Table.Color.Header.Desc.Bg, theme.Table.Color.Header.Desc.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Desc.Align),
				Colors: combineColors(theme.Table.Color.Row.Desc.Fg, theme.Table.Color.Row.Desc.Bg, theme.Table.Color.Row.Desc.Attr),
			}
		case "relpath":
			headerStyles[h] = table.ColumnConfig{
				Name: "relpath",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.RelPath.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.RelPath.Fg, theme.Table.Color.Header.RelPath.Bg, theme.Table.Color.Header.RelPath.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.RelPath.Align),
				Colors: combineColors(theme.Table.Color.Row.RelPath.Fg, theme.Table.Color.Row.RelPath.Bg, theme.Table.Color.Row.RelPath.Attr),
			}
		case "path":
			headerStyles[h] = table.ColumnConfig{
				Name: "path",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Path.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Path.Fg, theme.Table.Color.Header.Path.Bg, theme.Table.Color.Header.Path.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Path.Align),
				Colors: combineColors(theme.Table.Color.Row.Path.Fg, theme.Table.Color.Row.Path.Bg, theme.Table.Color.Row.Path.Attr),
			}
		case "url":
			headerStyles[h] = table.ColumnConfig{
				Name: "url",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Url.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Url.Fg, theme.Table.Color.Header.Url.Bg, theme.Table.Color.Header.Url.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Url.Align),
				Colors: combineColors(theme.Table.Color.Row.Url.Fg, theme.Table.Color.Row.Url.Bg, theme.Table.Color.Row.Url.Attr),
			}
		case "task":
			headerStyles[h] = table.ColumnConfig{
				Name: "task",

				AlignHeader:  GetAlign(*theme.Table.Color.Header.Task.Align),
				ColorsHeader: combineColors(theme.Table.Color.Header.Task.Fg, theme.Table.Color.Header.Task.Bg, theme.Table.Color.Header.Task.Attr),

				Align:  GetAlign(*theme.Table.Color.Row.Task.Align),
				Colors: combineColors(theme.Table.Color.Row.Task.Fg, theme.Table.Color.Row.Task.Bg, theme.Table.Color.Row.Task.Attr),
			}
		}
	}

	headers := []table.ColumnConfig{}
	for _, h := range defaultHeaders {
		headers = append(headers, headerStyles[h])
	}

	for i := range taskHeaders {
		hh := table.ColumnConfig{
			Number:       len(defaultHeaders) + 1 + i,
			AlignHeader:  GetAlign(*theme.Table.Color.Header.Output.Align),
			ColorsHeader: combineColors(theme.Table.Color.Header.Output.Fg, theme.Table.Color.Header.Output.Bg, theme.Table.Color.Header.Output.Attr),

			Align:  GetAlign(*theme.Table.Color.Row.Output.Align),
			Colors: combineColors(theme.Table.Color.Row.Output.Fg, theme.Table.Color.Row.Output.Bg, theme.Table.Color.Row.Output.Attr),
		}

		headers = append(headers, hh)
	}

	t.SetColumnConfigs(headers)

	return t
}

func FormatTable(theme dao.Theme) table.Style {
	return table.Style{
		Name: theme.Name,
		Box:  theme.Table.Box,

		Format: table.FormatOptions{
			Header: GetFormat(*theme.Table.Format.Header),
			Row:    GetFormat(*theme.Table.Format.Row),
		},

		Options: table.Options{
			DrawBorder:      *theme.Table.Options.DrawBorder,
			SeparateColumns: *theme.Table.Options.SeparateColumns,
			SeparateHeader:  *theme.Table.Options.SeparateHeader,
			SeparateRows:    *theme.Table.Options.SeparateRows,
		},

		// Border colors
		Color: table.ColorOptions{
			Header:       combineColors(theme.Table.Color.Border.Header.Fg, theme.Table.Color.Border.Header.Bg, theme.Table.Color.Border.Header.Attr),
			Row:          combineColors(theme.Table.Color.Border.Row.Fg, theme.Table.Color.Border.Row.Bg, theme.Table.Color.Border.Row.Attr),
			RowAlternate: combineColors(theme.Table.Color.Border.RowAlternate.Fg, theme.Table.Color.Border.RowAlternate.Bg, theme.Table.Color.Border.RowAlternate.Attr),
			Footer:       combineColors(theme.Table.Color.Border.Footer.Fg, theme.Table.Color.Border.Footer.Bg, theme.Table.Color.Border.Footer.Attr),
		},
	}
}

func RenderTable(t table.Writer, output string) {
	fmt.Println()
	switch output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
	fmt.Println()
}
