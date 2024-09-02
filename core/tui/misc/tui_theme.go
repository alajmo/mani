package misc

import (
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Default
var STYLE_DEFAULT StyleOption

// Border
var STYLE_BORDER StyleOption
var STYLE_BORDER_FOCUS StyleOption

// Title
var STYLE_TITLE StyleOption
var STYLE_TITLE_ACTIVE StyleOption

// Table Header
var STYLE_TABLE_HEADER StyleOption

// Item
var STYLE_ITEM StyleOption
var STYLE_ITEM_FOCUSED StyleOption
var STYLE_ITEM_SELECTED StyleOption

// Button
var STYLE_BUTTON StyleOption
var STYLE_BUTTON_ACTIVE StyleOption

// Search
var STYLE_SEARCH_LABEL StyleOption
var STYLE_SEARCH_TEXT StyleOption

// Filter
var STYLE_FILTER_LABEL StyleOption
var STYLE_FILTER_TEXT StyleOption

// Shortcut
var STYLE_SHORTCUT_LABEL StyleOption
var STYLE_SHORTCUT_TEXT StyleOption

type StyleOption struct {
	Fg    tcell.Color
	Bg    tcell.Color
	Attr  tcell.AttrMask
	Align int

	FgStr     string
	BgStr     string
	AttrStr   string
	AlignStr  string
	FormatStr string

	Style tcell.Style
}

func LoadStyles(tui *dao.TUI) {
	// Default
	STYLE_DEFAULT = initStyle(tui.Default)

	// Border
	STYLE_BORDER = initStyle(tui.Border)
	STYLE_BORDER_FOCUS = initStyle(tui.BorderFocus)

	// Title
	STYLE_TITLE = initStyle(tui.Title)
	STYLE_TITLE_ACTIVE = initStyle(tui.TitleActive)

	// Table Header
	STYLE_TABLE_HEADER = initStyle(tui.TableHeader)

	// Item
	STYLE_ITEM = initStyle(tui.Item)
	STYLE_ITEM_FOCUSED = initStyle(tui.ItemFocused)
	STYLE_ITEM_SELECTED = initStyle(tui.ItemSelected)

	// Button
	STYLE_BUTTON = initStyle(tui.Button)
	STYLE_BUTTON_ACTIVE = initStyle(tui.ButtonActive)

	// Search
	STYLE_SEARCH_LABEL = initStyle(tui.SearchLabel)
	STYLE_SEARCH_TEXT = initStyle(tui.SearchText)

	// Filter
	STYLE_FILTER_LABEL = initStyle(tui.FilterLabel)
	STYLE_FILTER_TEXT = initStyle(tui.FilterText)

	// Shortcut
	STYLE_SHORTCUT_LABEL = initStyle(tui.ShortcutLabel)
	STYLE_SHORTCUT_TEXT = initStyle(tui.ShortcutText)
}

func initStyle(opts *dao.ColorOptions) StyleOption {
	fg := tcell.GetColor(*opts.Fg)
	bg := tcell.GetColor(*opts.Bg)
	attr := getAttr(*opts.Attr)

	style := StyleOption{
		Fg:    fg,
		Bg:    bg,
		Attr:  attr,
		Align: getAlign(opts.Align),

		FgStr:     *opts.Fg,
		BgStr:     *opts.Bg,
		AttrStr:   *opts.Attr,
		FormatStr: *opts.Format,

		Style: tcell.StyleDefault.Foreground(fg).Background(bg).Attributes(attr),
	}

	return style
}

func Colorize(value string, opts dao.ColorOptions) string {
	return " [-:-:-]" + fmt.Sprintf("[%s:%s:%s]%s", *opts.Fg, *opts.Bg, *opts.Attr, value) + "[-:-:-] "
}

func ColorizeTitle(value string, opts dao.ColorOptions) string {
	return " [-:-:-]" + fmt.Sprintf("[%s:%s:%s] %s ", *opts.Fg, *opts.Bg, *opts.Attr, value) + "[-:-:-] "
}

func getAttr(attrStr string) tcell.AttrMask {
	var attr tcell.AttrMask
	switch attrStr {
	case "b", "bold":
		attr = tcell.AttrBold
	case "d", "dim":
		attr = tcell.AttrDim
	case "i", "italic":
		attr = tcell.AttrItalic
	case "u", "underline":
		attr = tcell.AttrUnderline
	default:
		attr = tcell.AttrNone
	}

	return attr
}

func getAlign(alignStr *string) int {
	if alignStr == nil {
		return tview.AlignLeft
	}

	lowerAlign := strings.ToLower(*alignStr)
	switch lowerAlign {
	case "l", "left":
		return tview.AlignLeft
	case "r", "right":
		return tview.AlignRight
	case "b", "bottom":
		return tview.AlignBottom
	case "t", "top":
		return tview.AlignTop
	case "c", "center":
		return tview.AlignCenter
	}

	return tview.AlignLeft
}

func PadString(name string) string {
	return " " + strings.TrimSpace(name) + " "
}
