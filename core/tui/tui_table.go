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
	UpdateTable     func()
}

func (t *TUITable) createTable() {
	table := tview.NewTable()
	table.SetFixed(1, 0)           // Fixed header
	table.Select(1, 0)             // Select first row
	table.SetEvaluateAllRows(true) // Avoid resizing of headers when scrolling
	table.SetBorder(true).SetBorderPadding(0, 0, 2, 2)
	table.SetSelectable(true, false) // Only rows can be selected
	table.SetBackgroundColor(tcell.ColorDefault)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ': // space: Toggle item
				t.ToggleSelected()
				return nil
			case 'A': // Select all
				t.SelectAllRows()
				t.updateCellStyles()
				return nil
			case 'C': // Unselect all all
				t.DeSelectAllRows()
				t.updateCellStyles()
				return nil
			case 'F': // Clear filter
				t.ClearFilters()
				t.updateCellStyles()
				return nil
			case 'd': // Open description modal
				t.DescribeRow()
				return nil
			}
			// case tcell.KeyCtrlA: // Toggle all rows
			// 	t.ToggleAllRows()
			// 	t.updateCellStyles()
			// 	return nil
		}
		return event
	})

	// Event Listeners
	table.SetSelectionChangedFunc(func(row, column int) {
		t.updateCellStyles()
	})

	t.Table = table
}

func (t *TUITable) updateCellStyles() {
	// Define the four states
	// Focused row and unselected (black background and blue text)
	// Focused row and selected (blue background and yellow text)
	// Unfocused row and selected (blue background and black text)
	// Unfocused row and selected (black background and white text)

	// SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlack)).
	focusedUnselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	focusedSelectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue).Attributes(tcell.AttrBold)
	unfocusedSelectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorRed).Attributes(tcell.AttrBold)
	unfocusedUnselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)

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
