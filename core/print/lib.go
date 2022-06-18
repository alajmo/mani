package print

import (
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/mani/core"
)

// Format map against go-pretty/table
func GetFormat(s string) text.Format {
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
func GetAlign(s string) text.Align {
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
func GetFg(s string) *text.Color {
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
func GetBg(s string) *text.Color {
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
func GetAttr(s string) *text.Color {
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

	fgVal := GetFg(*fg)
	if *fg != "" && fgVal != nil {
		colors = append(colors, *fgVal)
	}

	bgVal := GetBg(*bg)
	if *bg != "" && bgVal != nil {
		colors = append(colors, *bgVal)
	}

	attrVal := GetAttr(*attr)
	if *attr != "" && attrVal != nil {
		colors = append(colors, *attrVal)
	}

	return colors
}
