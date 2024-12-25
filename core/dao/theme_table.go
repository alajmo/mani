package dao

import (
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/alajmo/mani/core"
)

var DefaultTable = Table{
	Box: table.StyleDefault.Box,

	Style: "ascii",

	Border: &Border{
		Around:  core.Ptr(false),
		Columns: core.Ptr(true),
		Header:  core.Ptr(true),
		Rows:    core.Ptr(true),
	},

	Header: &ColorOptions{
		Fg:     core.Ptr("#d787ff"),
		Attr:   core.Ptr("bold"),
		Format: core.Ptr(""),
	},

	TitleColumn: &ColorOptions{
		Fg:     core.Ptr("#5f87d7"),
		Attr:   core.Ptr("bold"),
		Format: core.Ptr(""),
	},
}

type Border struct {
	Around  *bool `yaml:"around"`
	Columns *bool `yaml:"columns"`
	Header  *bool `yaml:"header"`
	Rows    *bool `yaml:"rows"`
}

type Table struct {
	// Stylable via YAML
	Style       string        `yaml:"style"`
	Border      *Border       `yaml:"border"`
	Header      *ColorOptions `yaml:"header"`
	TitleColumn *ColorOptions `yaml:"title_column"`

	// Not stylable via YAML
	Box table.BoxStyle `yaml:"-"`
}

func LoadTableTheme(mTable *Table) {
	// Table
	style := strings.ToLower(mTable.Style)
	switch style {
	case "light":
		mTable.Box = table.StyleLight.Box
	case "bold":
		mTable.Box = table.StyleBold.Box
	case "double":
		mTable.Box = table.StyleDouble.Box
	case "rounded":
		mTable.Box = table.StyleRounded.Box
	default: // ascii
		mTable.Box = table.StyleBoxDefault
	}

	// Options
	if mTable.Border == nil {
		mTable.Border = DefaultTable.Border
	} else {
		if mTable.Border.Around == nil {
			mTable.Border.Around = DefaultTable.Border.Around
		}

		if mTable.Border.Columns == nil {
			mTable.Border.Columns = DefaultTable.Border.Columns
		}

		if mTable.Border.Header == nil {
			mTable.Border.Header = DefaultTable.Border.Header
		}

		if mTable.Border.Rows == nil {
			mTable.Border.Rows = DefaultTable.Border.Rows
		}
	}

	// Header
	if mTable.Header == nil {
		mTable.Header = DefaultTable.Header
	} else {
		mTable.Header = MergeThemeOptions(mTable.Header, DefaultTable.Header)
	}

	// Title Column
	if mTable.TitleColumn == nil {
		mTable.TitleColumn = DefaultTable.TitleColumn
	} else {
		mTable.TitleColumn = MergeThemeOptions(mTable.TitleColumn, DefaultTable.TitleColumn)
	}
}
