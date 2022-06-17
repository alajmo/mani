package dao

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type TableOptions struct {
	DrawBorder      *bool `yaml:"draw_border"`
	SeparateColumns *bool `yaml:"separate_columns"`
	SeparateHeader  *bool `yaml:"separate_header"`
	SeparateRows    *bool `yaml:"separate_rows"`
	SeparateFooter  *bool `yaml:"separate_footer"`
}

type TableFormat struct {
	Header *string `yaml:"header"`
	Row    *string `yaml:"row"`
}

type ColorOptions struct {
	Fg    *string `yaml:"fg"`
	Bg    *string `yaml:"bg"`
	Align *string `yaml:"align"`
	Attr  *string `yaml:"attr"`
}

type BorderColors struct {
	Header       *ColorOptions `yaml:"header"`
	Row          *ColorOptions `yaml:"row"`
	RowAlternate *ColorOptions `yaml:"row_alt"`
	Footer       *ColorOptions `yaml:"footer"`
}

type CellColors struct {
	Project *ColorOptions `yaml:"project"`
	Synced  *ColorOptions `yaml:"synced"`
	Tag     *ColorOptions `yaml:"tag"`
	Desc    *ColorOptions `yaml:"desc"`
	RelPath *ColorOptions `yaml:"rel_path"`
	Path    *ColorOptions `yaml:"path"`
	Url     *ColorOptions `yaml:"url"`
	Task    *ColorOptions `yaml:"task"`
	Output  *ColorOptions `yaml:"output"`
}

type TableColor struct {
	Border *BorderColors `yaml:"border"`
	Header *CellColors   `yaml:"header"`
	Row    *CellColors   `yaml:"row"`
}

type Table struct {
	// Stylable via YAML
	Name    string        `yaml:"name"`
	Style   string        `yaml:"style"`
	Color   *TableColor   `yaml:"color"`
	Format  *TableFormat  `yaml:"format"`
	Options *TableOptions `yaml:"options"`

	// Not stylable via YAML
	Box table.BoxStyle `yaml:"-"`
}

type Tree struct {
	Style string `yaml:"style"`
}

type Text struct {
	Prefix       bool     `yaml:"prefix"`
	PrefixColors []string `yaml:"prefix_colors"`
	Header       bool     `yaml:"header"`
	HeaderChar   string   `yaml:"header_char"`
	HeaderPrefix string   `yaml:"header_prefix"`
}

type Theme struct {
	Name  string `yaml:"name"`
	Table Table  `yaml:"table"`
	Tree  Tree   `yaml:"tree"`
	Text  Text   `yaml:"text"`

	context     string
	contextLine int
}

type Row struct {
	Columns []string
}

type TableOutput struct {
	Headers []string
	Rows    []Row
}

func (t *Theme) GetContext() string {
	return t.context
}

func (t *Theme) GetContextLine() int {
	return t.contextLine
}

func (r Row) GetValue(_ string, i int) string {
	if i < len(r.Columns) {
		return r.Columns[i]
	}

	return ""
}

// Table Box Styles

var StyleBoxLight = table.BoxStyle{
	BottomLeft:       "└",
	BottomRight:      "┘",
	BottomSeparator:  "┴",
	EmptySeparator:   text.RepeatAndTrim(" ", text.RuneCount("┼")),
	Left:             "│",
	LeftSeparator:    "├",
	MiddleHorizontal: "─",
	MiddleSeparator:  "┼",
	MiddleVertical:   "│",
	PaddingLeft:      " ",
	PaddingRight:     " ",
	PageSeparator:    "\n",
	Right:            "│",
	RightSeparator:   "┤",
	TopLeft:          "┌",
	TopRight:         "┐",
	TopSeparator:     "┬",
	UnfinishedRow:    " ≈",
}

var StyleBoxASCII = table.BoxStyle{
	BottomLeft:       "+",
	BottomRight:      "+",
	BottomSeparator:  "+",
	EmptySeparator:   text.RepeatAndTrim(" ", text.RuneCount("+")),
	Left:             "|",
	LeftSeparator:    "+",
	MiddleHorizontal: "-",
	MiddleSeparator:  "+",
	MiddleVertical:   "|",
	PaddingLeft:      " ",
	PaddingRight:     " ",
	PageSeparator:    "\n",
	Right:            "|",
	RightSeparator:   "+",
	TopLeft:          "+",
	TopRight:         "+",
	TopSeparator:     "+",
	UnfinishedRow:    " ~",
}

var DefaultTree = Tree{
	Style: "connected-light",
}

var DefaultText = Text{
	Prefix:       true,
	PrefixColors: []string{"green", "blue", "red", "yellow", "magenta", "cyan"},
	Header:       true,
	HeaderPrefix: "TASK",
	HeaderChar:   "*",
}

var DefaultTable = Table{
	Style: "default",
	Box:   StyleBoxASCII,

	Format: &TableFormat{
		Header: core.Ptr("title"),
		Row:    core.Ptr(""),
	},

	Options: &TableOptions{
		DrawBorder:      core.Ptr(false),
		SeparateColumns: core.Ptr(true),
		SeparateHeader:  core.Ptr(true),
		SeparateRows:    core.Ptr(false),
		SeparateFooter:  core.Ptr(false),
	},

	Color: &TableColor{
		Border: &BorderColors{
			Header: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},

			Row: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},

			RowAlternate: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},

			Footer: &ColorOptions{
				Fg:   core.Ptr(""),
				Bg:   core.Ptr(""),
				Attr: core.Ptr("faint"),
			},
		},

		Header: &CellColors{
			Project: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Synced: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Tag: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Desc: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			RelPath: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Path: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Url: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Task: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
			Output: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr("bold"),
			},
		},

		Row: &CellColors{
			Project: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Synced: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Tag: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Desc: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			RelPath: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Path: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Url: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Task: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
			Output: &ColorOptions{
				Fg:    core.Ptr(""),
				Bg:    core.Ptr(""),
				Align: core.Ptr(""),
				Attr:  core.Ptr(""),
			},
		},
	},
}

// Populates ThemeList
func (c *Config) GetThemeList() ([]Theme, []ResourceErrors[Theme]) {
	var themes []Theme
	count := len(c.Themes.Content)

	themeErrors := []ResourceErrors[Theme]{}
	foundErrors := false
	for i := 0; i < count; i += 2 {
		theme := &Theme{
			Name:        c.Themes.Content[i].Value,
			context:     c.Path,
			contextLine: c.Themes.Content[i].Line,
		}

		err := c.Themes.Content[i+1].Decode(theme)
		if err != nil {
			foundErrors = true
			themeError := ResourceErrors[Theme]{Resource: theme, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			themeErrors = append(themeErrors, themeError)
			continue
		}

		themes = append(themes, *theme)
	}

	// Loop through themes and set default values
	for i := range themes {
		// TEXT
		if themes[i].Text.PrefixColors == nil {
			themes[i].Text.PrefixColors = DefaultText.PrefixColors
		}

		// TABLE
		if themes[i].Table.Style == "ascii" {
			themes[i].Table.Box = StyleBoxASCII
		} else {
			themes[i].Table.Style = "default"
			themes[i].Table.Box = DefaultTable.Box
		}

		// Format
		if themes[i].Table.Format == nil {
			themes[i].Table.Format = DefaultTable.Format
		} else {
			if themes[i].Table.Format.Header == nil {
				themes[i].Table.Format.Header = DefaultTable.Format.Header
			}

			if themes[i].Table.Format.Row == nil {
				themes[i].Table.Format.Row = DefaultTable.Format.Row
			}
		}

		if themes[i].Table.Options == nil {
			themes[i].Table.Options = DefaultTable.Options
		} else {
			if themes[i].Table.Options.DrawBorder == nil {
				themes[i].Table.Options.DrawBorder = DefaultTable.Options.DrawBorder
			}

			if themes[i].Table.Options.SeparateColumns == nil {
				themes[i].Table.Options.SeparateColumns = DefaultTable.Options.SeparateColumns
			}

			if themes[i].Table.Options.SeparateHeader == nil {
				themes[i].Table.Options.SeparateHeader = DefaultTable.Options.SeparateHeader
			}

			if themes[i].Table.Options.SeparateRows == nil {
				themes[i].Table.Options.SeparateRows = DefaultTable.Options.SeparateRows
			}

			if themes[i].Table.Options.SeparateFooter == nil {
				themes[i].Table.Options.SeparateFooter = DefaultTable.Options.SeparateFooter
			}
		}

		// Colors
		if themes[i].Table.Color == nil {
			themes[i].Table.Color = DefaultTable.Color
		} else {
			// Border
			if themes[i].Table.Color.Border == nil {
				themes[i].Table.Color.Border = DefaultTable.Color.Border
			} else {
				// Header
				if themes[i].Table.Color.Border.Header == nil {
					themes[i].Table.Color.Border.Header = DefaultTable.Color.Border.Header
				} else {
					if themes[i].Table.Color.Border.Header.Fg == nil {
						themes[i].Table.Color.Border.Header.Fg = DefaultTable.Color.Border.Header.Fg
					}
					if themes[i].Table.Color.Border.Header.Bg == nil {
						themes[i].Table.Color.Border.Header.Bg = DefaultTable.Color.Border.Header.Bg
					}
					if themes[i].Table.Color.Border.Header.Attr == nil {
						themes[i].Table.Color.Border.Header.Attr = DefaultTable.Color.Border.Header.Attr
					}
				}

				// Row
				if themes[i].Table.Color.Border.Row == nil {
					themes[i].Table.Color.Border.Row = DefaultTable.Color.Border.Row
				} else {
					if themes[i].Table.Color.Border.Row.Fg == nil {
						themes[i].Table.Color.Border.Row.Fg = DefaultTable.Color.Border.Row.Fg
					}
					if themes[i].Table.Color.Border.Row.Bg == nil {
						themes[i].Table.Color.Border.Row.Bg = DefaultTable.Color.Border.Row.Bg
					}
					if themes[i].Table.Color.Border.Row.Attr == nil {
						themes[i].Table.Color.Border.Row.Attr = DefaultTable.Color.Border.Row.Attr
					}
				}

				// RowAlternate
				if themes[i].Table.Color.Border.RowAlternate == nil {
					themes[i].Table.Color.Border.RowAlternate = DefaultTable.Color.Border.RowAlternate
				} else {
					if themes[i].Table.Color.Border.RowAlternate.Fg == nil {
						themes[i].Table.Color.Border.RowAlternate.Fg = DefaultTable.Color.Border.RowAlternate.Fg
					}
					if themes[i].Table.Color.Border.RowAlternate.Bg == nil {
						themes[i].Table.Color.Border.RowAlternate.Bg = DefaultTable.Color.Border.RowAlternate.Bg
					}
					if themes[i].Table.Color.Border.RowAlternate.Attr == nil {
						themes[i].Table.Color.Border.RowAlternate.Attr = DefaultTable.Color.Border.RowAlternate.Attr
					}
				}

				// Footer
				if themes[i].Table.Color.Border.Footer == nil {
					themes[i].Table.Color.Border.Footer = DefaultTable.Color.Border.Footer
				} else {
					if themes[i].Table.Color.Border.Footer.Fg == nil {
						themes[i].Table.Color.Border.Footer.Fg = DefaultTable.Color.Border.Footer.Fg
					}
					if themes[i].Table.Color.Border.Footer.Bg == nil {
						themes[i].Table.Color.Border.Footer.Bg = DefaultTable.Color.Border.Footer.Bg
					}
					if themes[i].Table.Color.Border.Footer.Attr == nil {
						themes[i].Table.Color.Border.Footer.Attr = DefaultTable.Color.Border.Footer.Attr
					}
				}

			}

			// Header
			if themes[i].Table.Color.Header == nil {
				themes[i].Table.Color.Header = DefaultTable.Color.Header
			} else {
				// Project
				if themes[i].Table.Color.Header.Project == nil {
					themes[i].Table.Color.Header.Project = DefaultTable.Color.Header.Project
				} else {
					if themes[i].Table.Color.Header.Project.Fg == nil {
						themes[i].Table.Color.Header.Project.Fg = DefaultTable.Color.Header.Project.Fg
					}
					if themes[i].Table.Color.Header.Project.Bg == nil {
						themes[i].Table.Color.Header.Project.Bg = DefaultTable.Color.Header.Project.Bg
					}
					if themes[i].Table.Color.Header.Project.Align == nil {
						themes[i].Table.Color.Header.Project.Align = DefaultTable.Color.Header.Project.Align
					}
					if themes[i].Table.Color.Header.Project.Attr == nil {
						themes[i].Table.Color.Header.Project.Attr = DefaultTable.Color.Header.Project.Attr
					}
				}

				// Synced
				if themes[i].Table.Color.Header.Synced == nil {
					themes[i].Table.Color.Header.Synced = DefaultTable.Color.Header.Synced
				} else {
					if themes[i].Table.Color.Header.Synced.Fg == nil {
						themes[i].Table.Color.Header.Synced.Fg = DefaultTable.Color.Header.Synced.Fg
					}
					if themes[i].Table.Color.Header.Synced.Bg == nil {
						themes[i].Table.Color.Header.Synced.Bg = DefaultTable.Color.Header.Synced.Bg
					}
					if themes[i].Table.Color.Header.Synced.Align == nil {
						themes[i].Table.Color.Header.Synced.Align = DefaultTable.Color.Header.Synced.Align
					}
					if themes[i].Table.Color.Header.Synced.Attr == nil {
						themes[i].Table.Color.Header.Synced.Attr = DefaultTable.Color.Header.Synced.Attr
					}
				}

				// Tag
				if themes[i].Table.Color.Header.Tag == nil {
					themes[i].Table.Color.Header.Tag = DefaultTable.Color.Header.Tag
				} else {
					if themes[i].Table.Color.Header.Tag.Fg == nil {
						themes[i].Table.Color.Header.Tag.Fg = DefaultTable.Color.Header.Tag.Fg
					}
					if themes[i].Table.Color.Header.Tag.Bg == nil {
						themes[i].Table.Color.Header.Tag.Bg = DefaultTable.Color.Header.Tag.Bg
					}
					if themes[i].Table.Color.Header.Tag.Align == nil {
						themes[i].Table.Color.Header.Tag.Align = DefaultTable.Color.Header.Tag.Align
					}
					if themes[i].Table.Color.Header.Tag.Attr == nil {
						themes[i].Table.Color.Header.Tag.Attr = DefaultTable.Color.Header.Tag.Attr
					}
				}

				// Desc
				if themes[i].Table.Color.Header.Desc == nil {
					themes[i].Table.Color.Header.Desc = DefaultTable.Color.Header.Desc
				} else {
					if themes[i].Table.Color.Header.Desc.Fg == nil {
						themes[i].Table.Color.Header.Desc.Fg = DefaultTable.Color.Header.Desc.Fg
					}
					if themes[i].Table.Color.Header.Desc.Bg == nil {
						themes[i].Table.Color.Header.Desc.Bg = DefaultTable.Color.Header.Desc.Bg
					}
					if themes[i].Table.Color.Header.Desc.Align == nil {
						themes[i].Table.Color.Header.Desc.Align = DefaultTable.Color.Header.Desc.Align
					}
					if themes[i].Table.Color.Header.Desc.Attr == nil {
						themes[i].Table.Color.Header.Desc.Attr = DefaultTable.Color.Header.Desc.Attr
					}
				}

				// RelPath
				if themes[i].Table.Color.Header.RelPath == nil {
					themes[i].Table.Color.Header.RelPath = DefaultTable.Color.Header.RelPath
				} else {
					if themes[i].Table.Color.Header.RelPath.Fg == nil {
						themes[i].Table.Color.Header.RelPath.Fg = DefaultTable.Color.Header.RelPath.Fg
					}
					if themes[i].Table.Color.Header.RelPath.Bg == nil {
						themes[i].Table.Color.Header.RelPath.Bg = DefaultTable.Color.Header.RelPath.Bg
					}
					if themes[i].Table.Color.Header.RelPath.Align == nil {
						themes[i].Table.Color.Header.RelPath.Align = DefaultTable.Color.Header.RelPath.Align
					}
					if themes[i].Table.Color.Header.RelPath.Attr == nil {
						themes[i].Table.Color.Header.RelPath.Attr = DefaultTable.Color.Header.RelPath.Attr
					}
				}

				// Path
				if themes[i].Table.Color.Header.Path == nil {
					themes[i].Table.Color.Header.Path = DefaultTable.Color.Header.Path
				} else {
					if themes[i].Table.Color.Header.Path.Fg == nil {
						themes[i].Table.Color.Header.Path.Fg = DefaultTable.Color.Header.Path.Fg
					}
					if themes[i].Table.Color.Header.Path.Bg == nil {
						themes[i].Table.Color.Header.Path.Bg = DefaultTable.Color.Header.Path.Bg
					}
					if themes[i].Table.Color.Header.Path.Align == nil {
						themes[i].Table.Color.Header.Path.Align = DefaultTable.Color.Header.Path.Align
					}
					if themes[i].Table.Color.Header.Path.Attr == nil {
						themes[i].Table.Color.Header.Path.Attr = DefaultTable.Color.Header.Path.Attr
					}
				}

				// Url
				if themes[i].Table.Color.Header.Url == nil {
					themes[i].Table.Color.Header.Url = DefaultTable.Color.Header.Url
				} else {
					if themes[i].Table.Color.Header.Url.Fg == nil {
						themes[i].Table.Color.Header.Url.Fg = DefaultTable.Color.Header.Url.Fg
					}
					if themes[i].Table.Color.Header.Url.Bg == nil {
						themes[i].Table.Color.Header.Url.Bg = DefaultTable.Color.Header.Url.Bg
					}
					if themes[i].Table.Color.Header.Url.Align == nil {
						themes[i].Table.Color.Header.Url.Align = DefaultTable.Color.Header.Url.Align
					}
					if themes[i].Table.Color.Header.Url.Attr == nil {
						themes[i].Table.Color.Header.Url.Attr = DefaultTable.Color.Header.Url.Attr
					}
				}

				// Task
				if themes[i].Table.Color.Header.Task == nil {
					themes[i].Table.Color.Header.Task = DefaultTable.Color.Header.Task
				} else {
					if themes[i].Table.Color.Header.Task.Fg == nil {
						themes[i].Table.Color.Header.Task.Fg = DefaultTable.Color.Header.Task.Fg
					}
					if themes[i].Table.Color.Header.Task.Bg == nil {
						themes[i].Table.Color.Header.Task.Bg = DefaultTable.Color.Header.Task.Bg
					}
					if themes[i].Table.Color.Header.Task.Align == nil {
						themes[i].Table.Color.Header.Task.Align = DefaultTable.Color.Header.Task.Align
					}
					if themes[i].Table.Color.Header.Task.Attr == nil {
						themes[i].Table.Color.Header.Task.Attr = DefaultTable.Color.Header.Task.Attr
					}
				}

				// Output
				if themes[i].Table.Color.Header.Output == nil {
					themes[i].Table.Color.Header.Output = DefaultTable.Color.Header.Output
				} else {
					if themes[i].Table.Color.Header.Output.Fg == nil {
						themes[i].Table.Color.Header.Output.Fg = DefaultTable.Color.Header.Output.Fg
					}
					if themes[i].Table.Color.Header.Output.Bg == nil {
						themes[i].Table.Color.Header.Output.Bg = DefaultTable.Color.Header.Output.Bg
					}
					if themes[i].Table.Color.Header.Output.Align == nil {
						themes[i].Table.Color.Header.Output.Align = DefaultTable.Color.Header.Output.Align
					}
					if themes[i].Table.Color.Header.Output.Attr == nil {
						themes[i].Table.Color.Header.Output.Attr = DefaultTable.Color.Header.Output.Attr
					}
				}
			}

			// Row
			if themes[i].Table.Color.Row == nil {
				themes[i].Table.Color.Row = DefaultTable.Color.Row
			} else {
				// Project
				if themes[i].Table.Color.Row.Project == nil {
					themes[i].Table.Color.Row.Project = DefaultTable.Color.Row.Project
				} else {
					if themes[i].Table.Color.Row.Project.Fg == nil {
						themes[i].Table.Color.Row.Project.Fg = DefaultTable.Color.Row.Project.Fg
					}
					if themes[i].Table.Color.Row.Project.Bg == nil {
						themes[i].Table.Color.Row.Project.Bg = DefaultTable.Color.Row.Project.Bg
					}
					if themes[i].Table.Color.Row.Project.Align == nil {
						themes[i].Table.Color.Row.Project.Align = DefaultTable.Color.Row.Project.Align
					}
					if themes[i].Table.Color.Row.Project.Attr == nil {
						themes[i].Table.Color.Row.Project.Attr = DefaultTable.Color.Row.Project.Attr
					}
				}

				// Synced
				if themes[i].Table.Color.Row.Synced == nil {
					themes[i].Table.Color.Row.Synced = DefaultTable.Color.Row.Synced
				} else {
					if themes[i].Table.Color.Row.Synced.Fg == nil {
						themes[i].Table.Color.Row.Synced.Fg = DefaultTable.Color.Row.Synced.Fg
					}
					if themes[i].Table.Color.Row.Synced.Bg == nil {
						themes[i].Table.Color.Row.Synced.Bg = DefaultTable.Color.Row.Synced.Bg
					}
					if themes[i].Table.Color.Row.Synced.Align == nil {
						themes[i].Table.Color.Row.Synced.Align = DefaultTable.Color.Row.Synced.Align
					}
					if themes[i].Table.Color.Row.Synced.Attr == nil {
						themes[i].Table.Color.Row.Synced.Attr = DefaultTable.Color.Row.Synced.Attr
					}
				}

				// Tag
				if themes[i].Table.Color.Row.Tag == nil {
					themes[i].Table.Color.Row.Tag = DefaultTable.Color.Row.Tag
				} else {
					if themes[i].Table.Color.Row.Tag.Fg == nil {
						themes[i].Table.Color.Row.Tag.Fg = DefaultTable.Color.Row.Tag.Fg
					}
					if themes[i].Table.Color.Row.Tag.Bg == nil {
						themes[i].Table.Color.Row.Tag.Bg = DefaultTable.Color.Row.Tag.Bg
					}
					if themes[i].Table.Color.Row.Tag.Align == nil {
						themes[i].Table.Color.Row.Tag.Align = DefaultTable.Color.Row.Tag.Align
					}
					if themes[i].Table.Color.Row.Tag.Attr == nil {
						themes[i].Table.Color.Row.Tag.Attr = DefaultTable.Color.Row.Tag.Attr
					}
				}

				// Desc
				if themes[i].Table.Color.Row.Desc == nil {
					themes[i].Table.Color.Row.Desc = DefaultTable.Color.Row.Desc
				} else {
					if themes[i].Table.Color.Row.Desc.Fg == nil {
						themes[i].Table.Color.Row.Desc.Fg = DefaultTable.Color.Row.Desc.Fg
					}
					if themes[i].Table.Color.Row.Desc.Bg == nil {
						themes[i].Table.Color.Row.Desc.Bg = DefaultTable.Color.Row.Desc.Bg
					}
					if themes[i].Table.Color.Row.Desc.Align == nil {
						themes[i].Table.Color.Row.Desc.Align = DefaultTable.Color.Row.Desc.Align
					}
					if themes[i].Table.Color.Row.Desc.Attr == nil {
						themes[i].Table.Color.Row.Desc.Attr = DefaultTable.Color.Row.Desc.Attr
					}
				}

				// RelPath
				if themes[i].Table.Color.Row.RelPath == nil {
					themes[i].Table.Color.Row.RelPath = DefaultTable.Color.Row.RelPath
				} else {
					if themes[i].Table.Color.Row.RelPath.Fg == nil {
						themes[i].Table.Color.Row.RelPath.Fg = DefaultTable.Color.Row.RelPath.Fg
					}
					if themes[i].Table.Color.Row.RelPath.Bg == nil {
						themes[i].Table.Color.Row.RelPath.Bg = DefaultTable.Color.Row.RelPath.Bg
					}
					if themes[i].Table.Color.Row.RelPath.Align == nil {
						themes[i].Table.Color.Row.RelPath.Align = DefaultTable.Color.Row.RelPath.Align
					}
					if themes[i].Table.Color.Row.RelPath.Attr == nil {
						themes[i].Table.Color.Row.RelPath.Attr = DefaultTable.Color.Row.RelPath.Attr
					}
				}

				// Path
				if themes[i].Table.Color.Row.Path == nil {
					themes[i].Table.Color.Row.Path = DefaultTable.Color.Row.Path
				} else {
					if themes[i].Table.Color.Row.Path.Fg == nil {
						themes[i].Table.Color.Row.Path.Fg = DefaultTable.Color.Row.Path.Fg
					}
					if themes[i].Table.Color.Row.Path.Bg == nil {
						themes[i].Table.Color.Row.Path.Bg = DefaultTable.Color.Row.Path.Bg
					}
					if themes[i].Table.Color.Row.Path.Align == nil {
						themes[i].Table.Color.Row.Path.Align = DefaultTable.Color.Row.Path.Align
					}
					if themes[i].Table.Color.Row.Path.Attr == nil {
						themes[i].Table.Color.Row.Path.Attr = DefaultTable.Color.Row.Path.Attr
					}
				}

				// Url
				if themes[i].Table.Color.Row.Url == nil {
					themes[i].Table.Color.Row.Url = DefaultTable.Color.Row.Url
				} else {
					if themes[i].Table.Color.Row.Url.Fg == nil {
						themes[i].Table.Color.Row.Url.Fg = DefaultTable.Color.Row.Url.Fg
					}
					if themes[i].Table.Color.Row.Url.Bg == nil {
						themes[i].Table.Color.Row.Url.Bg = DefaultTable.Color.Row.Url.Bg
					}
					if themes[i].Table.Color.Row.Url.Align == nil {
						themes[i].Table.Color.Row.Url.Align = DefaultTable.Color.Row.Url.Align
					}
					if themes[i].Table.Color.Row.Url.Attr == nil {
						themes[i].Table.Color.Row.Url.Attr = DefaultTable.Color.Row.Url.Attr
					}
				}

				// Task
				if themes[i].Table.Color.Row.Task == nil {
					themes[i].Table.Color.Row.Task = DefaultTable.Color.Row.Task
				} else {
					if themes[i].Table.Color.Row.Task.Fg == nil {
						themes[i].Table.Color.Row.Task.Fg = DefaultTable.Color.Row.Task.Fg
					}
					if themes[i].Table.Color.Row.Task.Bg == nil {
						themes[i].Table.Color.Row.Task.Bg = DefaultTable.Color.Row.Task.Bg
					}
					if themes[i].Table.Color.Row.Task.Align == nil {
						themes[i].Table.Color.Row.Task.Align = DefaultTable.Color.Row.Task.Align
					}
					if themes[i].Table.Color.Row.Task.Attr == nil {
						themes[i].Table.Color.Row.Task.Attr = DefaultTable.Color.Row.Task.Attr
					}
				}

				// Output
				if themes[i].Table.Color.Row.Output == nil {
					themes[i].Table.Color.Row.Output = DefaultTable.Color.Row.Output
				} else {
					if themes[i].Table.Color.Row.Output.Fg == nil {
						themes[i].Table.Color.Row.Output.Fg = DefaultTable.Color.Row.Output.Fg
					}
					if themes[i].Table.Color.Row.Output.Bg == nil {
						themes[i].Table.Color.Row.Output.Bg = DefaultTable.Color.Row.Output.Bg
					}
					if themes[i].Table.Color.Row.Output.Align == nil {
						themes[i].Table.Color.Row.Output.Align = DefaultTable.Color.Row.Output.Align
					}
					if themes[i].Table.Color.Row.Output.Attr == nil {
						themes[i].Table.Color.Row.Output.Attr = DefaultTable.Color.Row.Output.Attr
					}
				}
			}
		}
	}

	if foundErrors {
		return themes, themeErrors
	}

	return themes, nil
}

func (c Config) GetTheme(name string) (*Theme, error) {
	for _, theme := range c.ThemeList {
		if name == theme.Name {
			return &theme, nil
		}
	}

	return nil, &core.ThemeNotFound{Name: name}
}

func (c Config) GetThemeNames() []string {
	names := []string{}
	for _, theme := range c.ThemeList {
		names = append(names, theme.Name)
	}

	return names
}
