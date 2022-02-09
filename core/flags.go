package core

import (
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type TableOutput struct {
	Headers table.Row
	Rows    []table.Row
}

// CMD Flags

type ListFlags struct {
	NoHeaders bool
	NoBorders bool
	Output    string
	Theme     string
	Tree	bool
}

type ProjectFlags struct {
	Tags    []string
	Paths   []string
	Headers []string
	Edit    bool
}

type TagFlags struct {
	Headers []string
}

type TaskFlags struct {
	Headers []string
	Edit    bool
}

type RunFlags struct {
	Edit     bool
	Parallel bool
	DryRun   bool
	Describe bool
	Cwd      bool

	All			bool
	Projects    []string
	Paths       []string
	Tags        []string

	OmitEmpty bool
	Output string
}

type SyncFlags struct {
	Parallel bool
}

type InitFlags struct {
	AutoDiscovery bool
	Vcs string
}

// STYLES

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

var StyleNoBorders = table.BoxStyle{
	PaddingLeft:  "",
	PaddingRight: " ",
}

var ManiList = table.Style{
	Name: "table",

	Box: StyleBoxLight,

	Color: table.ColorOptions{
		// Header: text.Colors{ text.Bold },
	},

	Format: table.FormatOptions{
		Header: text.FormatDefault,
		Row:    text.FormatDefault,
		Footer: text.FormatUpper,
	},

	Options: table.Options{
		DrawBorder:      true,
		SeparateColumns: true,
		SeparateFooter:  false,
		SeparateHeader:  true,
		SeparateRows:    false,
	},
}

var TreeStyle list.Style
