package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
)

type TTable struct {
	Root   *tview.Flex
	Table  *tview.Table
	Filter *tview.InputField

	Title         string
	FilterValue   *string
	ShowHeaders   bool
	ToggleEnabled bool

	IsRowSelected   func(name string) bool
	ToggleSelectRow func(name string)
	SelectAll       func()
	UnselectAll     func()
	FilterRows      func()
	DescribeRow     func(name string)
	EditRow         func(name string)
}

func (t *TTable) Create() {
	// Init
	table := tview.NewTable()
	table.SetFixed(1, 1)             // Fixed header + name column
	table.Select(1, 0)               // Select first row
	table.SetEvaluateAllRows(true)   // Avoid resizing of headers when scrolling
	table.SetSelectable(true, false) // Only rows can be selected
	table.SetBackgroundColor(misc.STYLE_ITEM.Bg)
	filter := CreateFilter()

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(filter, 1, 0, false)

	root.SetTitleColor(misc.STYLE_TITLE.Fg)
	root.SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorder(true).
		SetBorderPadding(1, 0, 1, 1)

	t.Table = table
	t.Filter = filter
	t.Root = root

	if t.Title != "" {
		misc.SetActive(t.Root.Box, t.Title, false)
	}

	// Methods

	t.IsRowSelected = func(name string) bool { return false }
	t.ToggleSelectRow = func(name string) {}
	t.SelectAll = func() {}
	t.UnselectAll = func() {}
	t.FilterRows = func() {}
	t.DescribeRow = func(_ string) {}
	t.EditRow = func(projectName string) {}

	// Filter
	t.Filter.SetChangedFunc(func(_ string) {
		t.applyFilter()
		t.FilterRows()
	})

	t.Filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFocus := misc.App.GetFocus()
		if currentFocus == filter {
			switch event.Key() {
			case tcell.KeyEscape:
				t.ClearFilter()
				t.FilterRows()
				misc.App.SetFocus(table)
				return nil
			case tcell.KeyEnter:
				t.applyFilter()
				t.FilterRows()
				misc.App.SetFocus(table)
			}
			return event
		}
		return event
	})

	// Input
	t.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if t.ToggleEnabled {
				row, _ := table.GetSelection()
				name := strings.TrimSpace(table.GetCell(row, 0).Text)
				t.ToggleSelectRow(name)
			}
			return nil
		case tcell.KeyCtrlD:
			row, _ := table.GetSelection()
			_, _, _, height := table.GetInnerRect()
			newRow := min(row+height/2, table.GetRowCount()-1)
			table.Select(newRow, 0)
			return nil
		case tcell.KeyCtrlU:
			row, _ := table.GetSelection()
			_, _, _, height := table.GetInnerRect()
			newRow := max(row-height/2, 1)
			table.Select(newRow, 0)
			return nil
		case tcell.KeyCtrlF:
			row, _ := table.GetSelection()
			_, _, _, height := table.GetInnerRect()
			newRow := min(row+height, table.GetRowCount()-1)
			if newRow == 0 {
				newRow = 1 // Skip header
			}
			table.Select(newRow, 0)
			return nil
		case tcell.KeyCtrlB:
			row, _ := table.GetSelection()
			_, _, _, height := table.GetInnerRect()
			newRow := max(row-height, 1)
			table.Select(newRow, 0)
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case ' ': // Toggle item (space)
				if t.ToggleEnabled {
					row, _ := table.GetSelection()
					name := strings.TrimSpace(table.GetCell(row, 0).Text)
					t.ToggleSelectRow(name)
				}
				return nil
			case 'a': // Select all
				if t.ToggleEnabled {
					t.SelectAll()
				}
				return nil
			case 'c': // Unselect all
				if t.ToggleEnabled {
					t.UnselectAll()
				}
				return nil
			case 'f': // Filter rows
				ShowFilter(filter, *t.FilterValue)
				return nil
			case 'F': // Remove filter
				CloseFilter(filter)
				*t.FilterValue = ""
				return nil
			case 'o': // Edit in editor
				row, _ := t.Table.GetSelection()
				name := strings.TrimSpace(t.Table.GetCell(row, 0).Text)
				t.EditRow(name)
				return nil
			case 'd': // Open description modal
				row, _ := t.Table.GetSelection()
				name := strings.TrimSpace(t.Table.GetCell(row, 0).Text)
				t.DescribeRow(name)
				return nil
			}
		}
		return event
	})

	// Events
	t.Table.SetSelectionChangedFunc(func(row, column int) {
		t.UpdateRowStyle()
	})

	t.Table.SetFocusFunc(func() {
		InitFilter(t.Filter, *t.FilterValue)

		misc.PreviousPane = t.Table
		misc.SetActive(t.Root.Box, t.Title, true)
	})

	t.Table.SetBlurFunc(func() {
		misc.PreviousPane = t.Table
		misc.SetActive(t.Root.Box, t.Title, false)
	})
}

func (t *TTable) CreateTableHeader(header string) *tview.TableCell {
	// TODO: format
	return tview.NewTableCell(dao.StyleFormat(header, misc.STYLE_TABLE_HEADER.FormatStr)).
		SetTextColor(misc.STYLE_TABLE_HEADER.Fg).
		SetAttributes(misc.STYLE_TABLE_HEADER.Attr).
		SetAlign(misc.STYLE_TABLE_HEADER.Align).
		SetSelectable(false)
}

func (t *TTable) Update(headers []string, rows [][]string) {
	t.Table.Clear()

	// Add headers and updates style
	for col, header := range headers {
		if t.ShowHeaders {
			t.Table.SetCell(0, col, t.CreateTableHeader(misc.PadString(header)))
		} else {
			t.Table.SetCell(0, col, t.CreateTableHeader(""))
		}
	}

	// Add rows and updates style
	for i := range rows {
		for j := range rows[i] {
			name := misc.PadString(rows[i][j])
			cell := tview.NewTableCell(name)
			t.Table.SetCell(i+1, j, cell)
			t.SetRowSelect(i + 1)
		}
	}
}

func (t *TTable) UpdateRowStyle() {
	for row := 1; row < t.Table.GetRowCount(); row++ {
		t.SetRowSelect(row)
	}
}

func (t *TTable) ToggleSelectCurrentRow(name string) {
	index := -1
	for row := 1; row < t.Table.GetRowCount(); row++ {
		cell := strings.TrimSpace(t.Table.GetCell(row, 0).Text)
		if cell == name {
			index = row
			break
		}
	}
	t.SetRowSelect(index)
}

func (t *TTable) SetRowSelect(row int) {
	// Ignore header row
	focusedRow, _ := t.Table.GetSelection()
	if focusedRow == 0 {
		return
	}

	name := strings.TrimSpace(t.Table.GetCell(row, 0).Text)
	isSelected := t.IsRowSelected(name)
	isFocused := row == focusedRow

	style := tcell.StyleDefault
	if isFocused && isSelected {
		style = style.
			Foreground(misc.STYLE_ITEM_SELECTED.Fg).
			Background(misc.STYLE_ITEM_FOCUSED.Bg).
			Attributes(misc.STYLE_ITEM_SELECTED.Attr)
	} else if isFocused {
		style = style.
			Foreground(misc.STYLE_ITEM_FOCUSED.Fg).
			Background(misc.STYLE_ITEM_FOCUSED.Bg).
			Attributes(misc.STYLE_ITEM_FOCUSED.Attr)
	} else if isSelected {
		style = style.
			Foreground(misc.STYLE_ITEM_SELECTED.Fg).
			Background(misc.STYLE_ITEM_SELECTED.Bg).
			Attributes(misc.STYLE_ITEM_SELECTED.Attr)
	} else {
		style = style.
			Foreground(misc.STYLE_ITEM.Fg).
			Background(misc.STYLE_ITEM.Bg).
			Attributes(misc.STYLE_ITEM.Attr)
	}

	// Apply styles to all cells in the row
	for col := range t.Table.GetColumnCount() {
		cell := t.Table.GetCell(row, col)
		cell.SetStyle(style)
		cell.SetSelectedStyle(style)
	}
}

func (t *TTable) ClearFilter() {
	CloseFilter(t.Filter)
	*t.FilterValue = ""
}

func (t *TTable) applyFilter() {
	*t.FilterValue = t.Filter.GetText()
}
