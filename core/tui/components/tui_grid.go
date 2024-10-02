package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

type TUIGrid struct {
	Grid    *tview.Flex
	Headers *tview.Grid
	Rows    *tview.Grid
	Border  bool
}

func (t *TUIGrid) CreateGrid() {
	headers := tview.NewGrid()
	headers.SetBorderPadding(4, 4, 4, 4).SetBorder(t.Border)
	headers.SetBackgroundColor(misc.THEME.BG)
	headers.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			}
		}
		return event
	})
	headers.SetFocusFunc(func() {
		headers.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
	})
	headers.SetBlurFunc(func() {
		misc.PreviousPage = headers
		headers.SetBorderColor(misc.THEME.BORDER_COLOR)
	})

	// Rows
	rows := tview.NewGrid()
	rows.SetBorder(t.Border)
	rows.SetBackgroundColor(misc.THEME.BG)
	rows.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			}
		}
		return event
	})
	rows.SetFocusFunc(func() {
		rows.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
	})
	rows.SetBlurFunc(func() {
		misc.PreviousPage = rows
		rows.SetBorderColor(misc.THEME.BORDER_COLOR)
	})

	// k := tview.NewTextView().
	// 	SetText("HIIIIIIIIIIIII").
	// 	SetWordWrap(true).
	// 	SetTextAlign(tview.AlignLeft)

	t.Headers = headers
	t.Rows = rows

	t.Headers.SetMinSize(1, 1)
	t.Rows.SetMinSize(1, 1)

	t.Grid = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(headers, 1, 0, false).
		AddItem(rows, 200, 1, true)

	// AddItem(k, 0, 1, false)
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
