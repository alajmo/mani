package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

type TUIGrid struct {
	Grid    *tview.Flex
	Headers *tview.Grid
	Rows    *tview.Grid
	Title   string
	Border  bool
}

func (g *TUIGrid) CreateGrid() {
	headers := tview.NewGrid()
	headers.SetBorderPadding(4, 4, 4, 4).SetBorder(g.Border)
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
	if g.Title != "" {
		rows.SetTitle(fmt.Sprintf("[::b] %s ", g.Title))
	}

	rows.SetBorder(g.Border)
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
		misc.PreviousPage = rows
		g.SetActiveGrid(true)
	})
	rows.SetBlurFunc(func() {
		g.SetActiveGrid(false)
	})

	g.Headers = headers
	g.Rows = rows

	g.Headers.SetMinSize(1, 1)
	g.Rows.SetMinSize(1, 1)

	g.Grid = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(headers, 4, 0, false).
		AddItem(rows, 2000, 1, true)

	// AddItem(k, 0, 1, false)
}

func (t *TUIGrid) Update() {
	t.Headers.Clear()
	t.Headers.Box = tview.NewBox()
	t.Headers.SetGap(1, 1)
	t.Headers.SetBorders(true)
	t.Headers.SetBorderPadding(1, 1, 1, 1)
	t.Headers.SetColumns(16, 0)
	t.Headers.SetRows(6, 0)

	t.Rows.Clear()
	t.Rows.Box = tview.NewBox()
	t.Rows.SetGap(1, 1)
	t.Rows.SetBorders(true)
	t.Rows.SetBorderPadding(1, 1, 1, 1)
	t.Rows.SetColumns(16, 0)
	t.Rows.SetRows(40, 0)
}

func (t *TUIGrid) Populate() {
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

func (g *TUIGrid) SetActiveGrid(active bool) {
	if active {
		g.Rows.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		if g.Title != "" {
			g.Rows.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, g.Title))
		}
	} else {
		g.Rows.SetBorderColor(misc.THEME.BORDER_COLOR)
		if g.Title != "" {
			g.Rows.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR, g.Title))
		}
	}
}
