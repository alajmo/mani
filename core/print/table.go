package print

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func CreateTable(theme *dao.Theme, options PrintTableOptions, defaultHeaders []string, taskHeaders []string) table.Writer {
	t := table.NewWriter()

	t.SetOutputMirror(os.Stdout)
	t.SetStyle(FormatTable(*theme))
	t.SuppressEmptyColumns()

	headerStyles := map[string]table.ColumnConfig {
		"project": {
			Name: "project",
			AlignHeader: getAlign(*theme.Table.Color.Header.Project.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.Project.Fg, theme.Table.Color.Header.Project.Bg, theme.Table.Color.Header.Project.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.Project.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.Project.Fg, theme.Table.Color.Row.Project.Bg, theme.Table.Color.Row.Project.Attr)
				return colors.Sprint(val)
			}),
		},

		"tag": {
			Name: "tag",
			AlignHeader: getAlign(*theme.Table.Color.Header.Tag.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.Tag.Fg, theme.Table.Color.Header.Tag.Bg, theme.Table.Color.Header.Tag.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.Tag.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.Tag.Fg, theme.Table.Color.Row.Tag.Bg, theme.Table.Color.Row.Tag.Attr)
				return colors.Sprint(val)
			}),
		},

		"description": {
			Name: "description",
			AlignHeader: getAlign(*theme.Table.Color.Header.Desc.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.Desc.Fg, theme.Table.Color.Header.Desc.Bg, theme.Table.Color.Header.Desc.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.Desc.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.Desc.Fg, theme.Table.Color.Row.Desc.Bg, theme.Table.Color.Row.Desc.Attr)
				return colors.Sprint(val)
			}),
		},

		"relpath": {
			Name: "relpath",
			AlignHeader: getAlign(*theme.Table.Color.Header.RelPath.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.RelPath.Fg, theme.Table.Color.Header.RelPath.Bg, theme.Table.Color.Header.RelPath.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.RelPath.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.RelPath.Fg, theme.Table.Color.Row.RelPath.Bg, theme.Table.Color.Row.RelPath.Attr)
				return colors.Sprint(val)
			}),
		},

		"path": {
			Name: "path",
			AlignHeader: getAlign(*theme.Table.Color.Header.Path.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.Path.Fg, theme.Table.Color.Header.Path.Bg, theme.Table.Color.Header.Path.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.Path.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.Path.Fg, theme.Table.Color.Row.Path.Bg, theme.Table.Color.Row.Path.Attr)
				return colors.Sprint(val)
			}),
		},

		"url": {
			Name: "url",
			AlignHeader: getAlign(*theme.Table.Color.Header.Url.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.Url.Fg, theme.Table.Color.Header.Url.Bg, theme.Table.Color.Header.Url.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.Url.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.Url.Fg, theme.Table.Color.Row.Url.Bg, theme.Table.Color.Row.Url.Attr)
				return colors.Sprint(val)
			}),
		},

		"task": {
			Name: "task",
			AlignHeader: getAlign(*theme.Table.Color.Header.Task.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.Task.Fg, theme.Table.Color.Header.Task.Bg, theme.Table.Color.Header.Task.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.Task.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.Task.Fg, theme.Table.Color.Row.Task.Bg, theme.Table.Color.Row.Task.Attr)
				return colors.Sprint(val)
			}),
		},
	}

	headers := []table.ColumnConfig{}
	for _, h := range defaultHeaders {
		headers = append(headers, headerStyles[h])
	}

	for i := range taskHeaders {
		hh := table.ColumnConfig {
			Number: len(defaultHeaders) + 1 + i,
			AlignHeader: getAlign(*theme.Table.Color.Header.Output.Align),
			TransformerHeader: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Header.Output.Fg, theme.Table.Color.Header.Output.Bg, theme.Table.Color.Header.Output.Attr)
				return colors.Sprint(val)
			}),

			Align: getAlign(*theme.Table.Color.Row.Output.Align),
			Transformer: text.Transformer(func(val interface{}) string {
				colors := combineColors(theme.Table.Color.Row.Output.Fg, theme.Table.Color.Row.Output.Bg, theme.Table.Color.Row.Output.Attr)
				return colors.Sprint(val)
			}),
		}

		headers = append(headers, hh)
	}

	t.SetColumnConfigs(headers)

	return t
}

func FormatTable(theme dao.Theme) table.Style {
	return table.Style {
		Name: theme.Name,
		Box: theme.Table.Box,

		Format: table.FormatOptions {
			Header: getFormat(*theme.Table.Format.Header),
			Row: getFormat(*theme.Table.Format.Row),
		},

		Options: table.Options {
			DrawBorder:      *theme.Table.Options.DrawBorder,
			SeparateColumns: *theme.Table.Options.SeparateColumns,
			SeparateHeader:  *theme.Table.Options.SeparateHeader,
			SeparateRows:    *theme.Table.Options.SeparateRows,
		},

		// Border colors
		Color: table.ColorOptions {
			Header: combineColors(theme.Table.Color.Border.Header.Fg, theme.Table.Color.Border.Header.Bg, core.Ptr("")),
			Row: combineColors(theme.Table.Color.Border.Row.Fg, theme.Table.Color.Border.Row.Bg, core.Ptr("")),
			RowAlternate: combineColors(theme.Table.Color.Border.RowAlternate.Fg, theme.Table.Color.Border.RowAlternate.Bg, core.Ptr("")),
			Footer: combineColors(theme.Table.Color.Border.Footer.Fg, theme.Table.Color.Border.Footer.Bg, core.Ptr("")),
		},
	}
}

func RenderTable(t table.Writer, output string) {
	switch output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}

// Format map against go-pretty/table
func getFormat(s string) text.Format {
	switch s {
	case "default":
		return text.FormatDefault
	case "lower":
		return text.FormatLower
	case "title":
		return text.FormatTitle
	case "upper":
		return text.FormatUpper
	default:
		return text.FormatDefault
	}
}

// Align map against go-pretty/table
func getAlign(s string) text.Align {
	switch s {
	case "left":
		return text.AlignLeft
	case "center":
		return text.AlignCenter
	case "justify":
		return text.AlignJustify
	case "right":
		return text.AlignRight
	default:
		return text.AlignLeft
	}
}

// Foreground color map against go-pretty/table
func getFg(s string) *text.Color {
	switch s {
		// Normal colors
	case "black":
		return core.Ptr(text.FgBlack)
	case "red":
		return core.Ptr(text.FgRed)
	case "green":
		return core.Ptr(text.FgGreen)
	case "yellow":
		return core.Ptr(text.FgYellow)
	case "blue":
		return core.Ptr(text.FgBlue)
	case "magenta":
		return core.Ptr(text.FgMagenta)
	case "cyan":
		return core.Ptr(text.FgCyan)
	case "white":
		return core.Ptr(text.FgWhite)

		// High-intensity colors
	case "hi_black":
		return core.Ptr(text.FgHiBlack)
	case "hi_red":
		return core.Ptr(text.FgHiRed)
	case "hi_green":
		return core.Ptr(text.FgHiGreen)
	case "hi_yellow":
		return core.Ptr(text.FgHiYellow)
	case "hi_blue":
		return core.Ptr(text.FgHiBlue)
	case "hi_magenta":
		return core.Ptr(text.FgHiMagenta)
	case "hi_cyan":
		return core.Ptr(text.FgHiCyan)
	case "hi_white":
		return core.Ptr(text.FgHiWhite)

	default:
		return nil
	}
}

// Background color map against go-pretty/table
func getBg(s string) *text.Color {
	switch s {
		// Normal colors
	case "black":
		return core.Ptr(text.BgBlack)
	case "red":
		return core.Ptr(text.BgRed)
	case "green":
		return core.Ptr(text.BgGreen)
	case "yellow":
		return core.Ptr(text.BgYellow)
	case "blue":
		return core.Ptr(text.BgBlue)
	case "magenta":
		return core.Ptr(text.BgMagenta)
	case "cyan":
		return core.Ptr(text.BgCyan)
	case "white":
		return core.Ptr(text.BgWhite)

		// High-intensity colors
	case "hi_black":
		return core.Ptr(text.BgHiBlack)
	case "hi_red":
		return core.Ptr(text.BgHiRed)
	case "hi_green":
		return core.Ptr(text.BgHiGreen)
	case "hi_yellow":
		return core.Ptr(text.BgHiYellow)
	case "hi_blue":
		return core.Ptr(text.BgHiBlue)
	case "hi_magenta":
		return core.Ptr(text.BgHiMagenta)
	case "hi_cyan":
		return core.Ptr(text.BgHiCyan)
	case "hi_white":
		return core.Ptr(text.BgHiWhite)

	default:
		return nil
	}
}

// Attr (color) map against go-pretty/table (attributes belong to the same types as fg/bg)
func getAttr(s string) *text.Color {
	switch s {
	case "normal":
		return nil
	case "bold":
		return core.Ptr(text.Bold)
	case "faint":
		return core.Ptr(text.Faint)
	case "italic":
		return core.Ptr(text.Italic)
	case "underline":
		return core.Ptr(text.Underline)
	case "crossed_out":
		return core.Ptr(text.CrossedOut)

	default:
		return nil
	}
}

// Combine colors and attributes in one slice. We check if the values are valid, otherwise
// we get a nil pointer, in which case the values are not appended to the colors slice.
func combineColors(fg *string, bg *string, attr *string) text.Colors {
	colors := text.Colors{}

	fgVal := getFg(*fg)
	if *fg != "" && fgVal != nil {
		colors = append(colors, *fgVal)
	}

	bgVal := getBg(*bg)
	if *bg != "" && bgVal != nil {
		colors = append(colors, *bgVal)
	}

	attrVal := getAttr(*attr)
	if *attr != "" && attrVal != nil {
		colors = append(colors, *attrVal)
	}

	return colors
}
