package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

type TUIGrid struct {
	Grid *tview.Grid
}

func (t *TUIGrid) CreateGrid() {
	grid := tview.NewGrid()
	// grid.SetBorder(true).SetBorderPadding(0, 0, 1, 1)
	grid.SetBorder(true)
	grid.SetBackgroundColor(misc.THEME.BG)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			}
		}
		return event
	})

	grid.SetFocusFunc(func() {
		grid.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
	})
	grid.SetBlurFunc(func() {
		misc.PreviousPage = grid
		grid.SetBorderColor(misc.THEME.BORDER_COLOR)
	})

	t.Grid = grid
}

func CreateGridHeader(header string) *tview.TextView {
	column := tview.NewTextView().SetText(header)
	column.SetTextStyle(
		tcell.StyleDefault.
			Foreground(misc.THEME.TABLE_HEADER_FG).
			Background(misc.THEME.BG).
			Attributes(tcell.AttrBold),
	)
	column.SetTextAlign(tview.AlignLeft)

	// column.SetBorder(true)
	// column.SetBorderPadding(0, 0, 0, 0)
	// column.SetBorderColor(tcell.ColorYellow)

	return column
}
