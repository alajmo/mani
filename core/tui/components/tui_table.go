package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

type TUITable struct {
	Table           *tview.Table
	SelectEnabled   bool
	IsRowSelected   func(name string) bool
	ToggleSelected  func()
	SelectAllRows   func()
	DeSelectAllRows func()
	ClearFilters    func()
	DescribeRow     func()
	EditRow         func(name string)
}

func (t *TUITable) CreateTable() {
	table := tview.NewTable()
	table.SetFixed(1, 0)           // Fixed header
	table.Select(1, 0)             // Select first row
	table.SetEvaluateAllRows(true) // Avoid resizing of headers when scrolling
	table.SetBorder(true).SetBorderPadding(0, 0, 2, 2)
	table.SetSelectable(true, false) // Only rows can be selected
	table.SetBackgroundColor(misc.THEME.BG)
	// table.SetSelectedStyle(tcell.StyleDefault)

	t.IsRowSelected = func(name string) bool { return false }
	t.EditRow = func(projectName string) {}
	t.ToggleSelected = func() {}
	t.SelectAllRows = func() {}
	t.DeSelectAllRows = func() {}
	t.DescribeRow = func() {}

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'o': // Edit/Open file in editor
				row, _ := table.GetSelection()
				name := table.GetCell(row, 0).Text
				t.EditRow(name)
				return nil
			case ' ': // Toggle item (space)
				if t.SelectEnabled {
					t.ToggleSelected()
				}
				return nil
			case 'd': // Open description modal
				t.DescribeRow()
				return nil
			}
		}
		return event
	})

	// Event Listeners
	table.SetSelectionChangedFunc(func(row, column int) {
		t.UpdateCellStyles()
	})

	table.SetFocusFunc(func() {
		table.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
	})
	table.SetBlurFunc(func() {
		misc.PreviousPage = table
		table.SetBorderColor(misc.THEME.BORDER_COLOR)
	})

	t.Table = table
}

func (t *TUITable) UpdateCellStyles() {
	focusedRow, _ := t.Table.GetSelection()
	if focusedRow == 0 {
		return
	}

	for row := 1; row < t.Table.GetRowCount(); row++ {
		isSelected := false
		name := t.Table.GetCell(row, 0).Text

		if t.IsRowSelected(name) {
			isSelected = true
		}

		isFocused := row == focusedRow
		var style tcell.Style
		var selectedStyle tcell.Style

		if isFocused {
			if isSelected {
				style = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
				selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
			} else {
				style = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED).Background(misc.THEME.BG_FOCUSED)
				selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED).Background(misc.THEME.BG_FOCUSED)
			}
		} else {
			if isSelected {
				style = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
				selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
			} else {
				style = tcell.StyleDefault.Foreground(misc.THEME.FG).Background(misc.THEME.BG)
				selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG).Background(misc.THEME.BG)
			}
		}

		// style = focusedSelectedStyle

		for col := 0; col < t.Table.GetColumnCount(); col++ {
			t.Table.GetCell(row, col).SetStyle(style)
			t.Table.GetCell(row, col).SetSelectedStyle(selectedStyle)
		}
	}
}

func CreateTableHeader(header string) *tview.TableCell {
	return tview.NewTableCell(header).
		SetTextColor(misc.THEME.TABLE_HEADER_FG).
		SetAttributes(tcell.AttrBold).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)
}
