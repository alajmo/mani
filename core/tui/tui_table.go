package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUITable struct {
	Table           *tview.Table
	IsRowSelected   func(name string) bool
	ToggleSelected  func()
	SelectAllRows   func()
	DeSelectAllRows func()
	ClearFilters    func()
	DescribeRow     func()
	EditRow         func(name string)
}

func (t *TUITable) createTable() {
	table := tview.NewTable()
	table.SetFixed(1, 0)           // Fixed header
	table.Select(1, 0)             // Select first row
	table.SetEvaluateAllRows(true) // Avoid resizing of headers when scrolling
	table.SetBorder(true).SetBorderPadding(0, 0, 2, 2)
	table.SetSelectable(true, false) // Only rows can be selected
	table.SetBackgroundColor(THEME.BG)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'o': // Edit project
				row, _ := table.GetSelection()
				projectName := table.GetCell(row, 0).Text
				t.EditRow(projectName)
				return nil
			case ' ': // Toggle item (space)
				t.ToggleSelected()
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
		t.updateCellStyles()
	})

	table.SetFocusFunc(func() {
		table.SetBorderColor(THEME.BORDER_COLOR_FOCUS)
	})
	table.SetBlurFunc(func() {
		TUI.previousPage = table
		table.SetBorderColor(THEME.BORDER_COLOR)
	})

	t.Table = table
}

func (t *TUITable) updateCellStyles() {
	// Focused row and unselected
	// Focused row and selected
	// Unfocused row and selected
	// Unfocused row and selected
	focusedUnselectedStyle := tcell.StyleDefault.Foreground(THEME.BG_FOCUSED).Background(THEME.FG_FOCUSED)
	focusedSelectedStyle := tcell.StyleDefault.Foreground(THEME.BG_FOCUSED).Background(THEME.FG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
	unfocusedSelectedStyle := tcell.StyleDefault.Foreground(THEME.FG_FOCUSED_SELECTED).Background(THEME.FG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
	unfocusedUnselectedStyle := tcell.StyleDefault.Foreground(THEME.BG).Background(THEME.FG)

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

		if isFocused {
			if isSelected {
				style = focusedSelectedStyle
			} else {
				style = focusedUnselectedStyle
			}
		} else {
			if isSelected {
				style = unfocusedSelectedStyle
			} else {
				style = unfocusedUnselectedStyle
			}
		}
		for col := 0; col < t.Table.GetColumnCount(); col++ {
			t.Table.GetCell(row, col).SetStyle(style)
		}
	}
}

func createTableHeader(header string) *tview.TableCell {
	return tview.NewTableCell(header).
		SetTextColor(THEME.TABLE_HEADER_FG).
		SetAttributes(tcell.AttrBold).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)
}
