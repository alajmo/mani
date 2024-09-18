package tui

import "github.com/gdamore/tcell/v2"

var THEME = struct {
	FG                  tcell.Color
	BG                  tcell.Color
	FG_FOCUSED          tcell.Color
	BG_FOCUSED          tcell.Color
	FG_FOCUSED_SELECTED tcell.Color
	BG_FOCUSED_SELECTED tcell.Color

	BORDER_COLOR       tcell.Color
	BORDER_COLOR_FOCUS tcell.Color

	TITLE        tcell.Color
	TITLE_ACTIVE tcell.Color

	TABLE_HEADER_FG tcell.Color

	SEARCH_BG tcell.Color
	SEARCH_FG tcell.Color

	BTN_FG        tcell.Color
	BTN_BG        tcell.Color
	BTN_FG_ACTIVE tcell.Color
	BTN_BG_ACTIVE tcell.Color
}{
	FG:                  tcell.ColorDefault,
	BG:                  tcell.ColorDefault,
	FG_FOCUSED:          tcell.ColorWhite,
	BG_FOCUSED:          tcell.Color235,
	FG_FOCUSED_SELECTED: tcell.ColorBlue,
	BG_FOCUSED_SELECTED: tcell.Color235,

	BORDER_COLOR:       tcell.ColorWhite,
	BORDER_COLOR_FOCUS: tcell.ColorYellow,

	TITLE:        tcell.ColorDefault,
	TITLE_ACTIVE: tcell.ColorYellow,

	TABLE_HEADER_FG: tcell.ColorYellow,

	SEARCH_BG: tcell.ColorDefault,
	SEARCH_FG: tcell.ColorBlue,

	BTN_FG:        tcell.ColorWhite,
	BTN_BG:        tcell.ColorDefault,
	BTN_FG_ACTIVE: tcell.ColorYellow,
	BTN_BG_ACTIVE: tcell.ColorDefault,
}
