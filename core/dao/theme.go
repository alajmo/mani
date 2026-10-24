package dao

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/alajmo/mani/core"
	"github.com/gookit/color"
	"gopkg.in/yaml.v3"
)

type ColorOptions struct {
	Fg     *string `yaml:"fg"`
	Bg     *string `yaml:"bg"`
	Align  *string `yaml:"align"`
	Attr   *string `yaml:"attr"`
	Format *string `yaml:"format"`
}

type Theme struct {
	Name   string `yaml:"name"`
	Table  Table  `yaml:"table"`
	Tree   Tree   `yaml:"tree"`
	Stream Stream `yaml:"stream"`
	Block  Block  `yaml:"block"`
	TUI    TUI    `yaml:"tui"`
	Color  *bool  `yaml:"color"`

	context     string
	contextLine int
	node        *yaml.Node // definition in the config, used to merge themes defined in multiple configs
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

// Populates ThemeList
func (c *Config) ParseThemes() ([]Theme, []ResourceErrors[Theme]) {
	var themes []Theme
	pairs, err := sectionPairs(&c.Themes, "themes")
	if err != nil {
		return themes, []ResourceErrors[Theme]{{Resource: &Theme{context: c.Path, contextLine: -1}, Errors: []error{err}}}
	}

	themeErrors := []ResourceErrors[Theme]{}
	foundErrors := false
	for _, p := range pairs {
		theme := &Theme{
			Name:        p.key.Value,
			context:     c.Path,
			contextLine: p.key.Line,
			node:        p.value,
		}

		err := p.value.Decode(theme)
		if err != nil {
			foundErrors = true
			themeError := ResourceErrors[Theme]{Resource: theme, Errors: decodeErrors(err)}
			themeErrors = append(themeErrors, themeError)
			continue
		}

		themes = append(themes, *theme)
	}

	// Loop through themes and set default values
	for i := range themes {
		// Color
		if themes[i].Color == nil {
			themes[i].Color = new(true)
		}

		// Stream
		LoadStreamTheme(&themes[i].Stream)

		// Table
		LoadTableTheme(&themes[i].Table)

		// Tree
		LoadTreeTheme(&themes[i].Tree)

		// Block
		LoadBlockTheme(&themes[i].Block)

		// TUI
		LoadTUITheme(&themes[i].TUI)
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

// Merges default with user theme.
// Converts colors to hex, and align, attr, and format to its backend representation (single character).
func MergeThemeOptions(userOption *ColorOptions, defaultOption *ColorOptions) *ColorOptions {
	if userOption == nil {
		return &ColorOptions{
			Fg:     convertToHex(defaultOption.Fg),
			Bg:     convertToHex(defaultOption.Bg),
			Attr:   convertToAttr(defaultOption.Attr),
			Align:  convertToAlign(defaultOption.Align),
			Format: convertToFormat(defaultOption.Format),
		}
	}
	result := &ColorOptions{}

	if userOption.Fg == nil {
		result.Fg = convertToHex(defaultOption.Fg)
	} else {
		result.Fg = convertToHex(userOption.Fg)
	}

	if userOption.Bg == nil {
		result.Bg = convertToHex(defaultOption.Bg)
	} else {
		result.Bg = convertToHex(userOption.Bg)
	}

	if userOption.Attr == nil {
		result.Attr = convertToAttr(defaultOption.Attr)
	} else {
		result.Attr = convertToAttr(userOption.Attr)
	}

	if userOption.Align == nil {
		result.Align = convertToAlign(defaultOption.Align)
	} else {
		result.Align = convertToAlign(userOption.Align)
	}

	if userOption.Format == nil {
		result.Format = convertToFormat(defaultOption.Format)
	} else {
		result.Format = convertToFormat(userOption.Format)
	}

	return result
}

// Used for gookit/color printing stream
func StyleFg(colr string) color.RGBColor {
	// User provided
	if colr != "" {
		return color.HEX(colr)
	}

	// Default Fg color
	return color.Normal.RGB()
}

func StyleFormat(text string, format string) string {
	switch format {
	case "l":
		return strings.ToLower(text)
	case "u":
		return strings.ToUpper(text)
	case "t":
		caser := cases.Title(language.English)
		return caser.String(text)
	}

	return text
}

// Used for gookit/color printing tables/blocks
func StyleString(text string, opts ColorOptions, useColors bool) string {
	if !useColors {
		return text
	}

	// Format
	switch *opts.Format {
	case "l":
		text = strings.ToLower(text)
	case "u":
		text = strings.ToUpper(text)
	case "t":
		caser := cases.Title(language.English)
		text = caser.String(text)
	}

	// Fg
	var fgStr string
	if *opts.Fg != "" {
		fgStr = color.HEX(*opts.Fg).Sprint(text)
	} else {
		fgStr = text
	}

	// Attr
	attr := color.OpReset
	switch *opts.Attr {
	case "b":
		attr = color.OpBold
	case "i":
		attr = color.OpItalic
	case "u":
		attr = color.OpUnderscore
	}

	styledString := attr.Sprint(fgStr)

	return styledString
}

func convertToHex(s *string) *string {
	if s == nil || len(*s) == 0 {
		return new("-")
	}

	// Assume it's hex already
	if (*s)[0] == '#' {
		return s
	}

	// Named color
	hex := "#" + color.RGBFromString(*s).Hex()
	return &hex
}

func convertToAttr(attr *string) *string {
	if attr == nil || len(*attr) == 0 {
		return new("-")
	}

	attrStr := strings.ToLower(*attr)
	switch attrStr {
	case "b", "bold":
		return new("b")
	case "i", "italic":
		return new("i")
	case "u", "underline":
		return new("u")
	}

	return new("-")
}

func convertToAlign(align *string) *string {
	if align == nil || len(*align) == 0 {
		return new("")
	}

	alignStr := strings.ToLower(*align)
	switch alignStr {
	case "l", "left":
		return new("l")
	case "c", "center":
		return new("c")
	case "r", "right":
		return new("r")
	}

	return new("")
}

func convertToFormat(format *string) *string {
	if format == nil || len(*format) == 0 {
		return new("")
	}

	formatStr := strings.ToLower(*format)
	switch formatStr {
	case "t", "title":
		return new("t")
	case "l", "lower":
		return new("l")
	case "u", "upper":
		return new("u")
	}

	return new("")
}
