package print

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/text"
)

var StyleBoxDefault = table.BoxStyle {
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

var StyleBoxASCII = table.BoxStyle {
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

var StyleNoBorders = table.BoxStyle {
	PaddingLeft:      "",
	PaddingRight:     " ",
}

var ManiList = table.Style {
	Name: "table",

	Box: StyleBoxDefault,

	Color: table.ColorOptions {
		// Header: text.Colors{ text.Bold },
	},

	Format: table.FormatOptions {
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
