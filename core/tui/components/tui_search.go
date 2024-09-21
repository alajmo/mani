package components

import (
	"strings"

  "github.com/rivo/tview"

  "github.com/alajmo/mani/core/tui/misc"
)

func CreateSearchInput() *tview.InputField {
	search := tview.NewInputField().
		SetLabel("").
		SetFieldWidth(30).
		SetFieldBackgroundColor(misc.THEME.SEARCH_BG).
		SetFieldTextColor(misc.THEME.SEARCH_FG)

	return search
}

func ShowSearch() {
	misc.Search.SetLabel("search: ")
	misc.Search.SetText("")
	misc.App.SetFocus(misc.Search)
}

func EmptySearch() {
	misc.Search.SetLabel("")
	misc.Search.SetText("")
}

func SearchInList(list *tview.List, query string, lastFoundIndex *int, direction int) {
	itemCount := list.GetItemCount()
	startIndex := *lastFoundIndex

	if startIndex == -1 {
		startIndex = 0
	} else {
		startIndex += direction
	}

	searchIndex := startIndex
	for i := 0; i < itemCount; i++ {
		if searchIndex < 0 {
			searchIndex = itemCount - 1
		} else if searchIndex >= itemCount {
			searchIndex = 0
		}

		mainText, secondaryText := list.GetItemText(searchIndex)
		if strings.Contains(strings.ToLower(mainText), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(secondaryText), strings.ToLower(query)) {
			list.SetCurrentItem(searchIndex)
			*lastFoundIndex = searchIndex
			return
		}

		searchIndex += direction
	}

	*lastFoundIndex = -1
}

func SearchInTable(table *tview.Table, query string, lastFoundRow, lastFoundCol *int, direction int) {
	rowCount := table.GetRowCount()
	colCount := table.GetColumnCount()
	startRow := *lastFoundRow

	if startRow == -1 {
		startRow = 0
	} else {
		startRow += direction
	}

	searchRow := startRow
	for i := 0; i < rowCount; i++ {
		if searchRow < 0 {
			searchRow = rowCount - 1
		} else if searchRow >= rowCount {
			searchRow = 0
		}

		for col := 0; col < colCount; col++ {
			if cell := table.GetCell(searchRow, col); cell != nil {
				if strings.Contains(strings.ToLower(cell.Text), strings.ToLower(query)) {
					table.Select(searchRow, col)
					*lastFoundRow, *lastFoundCol = searchRow, col
					return
				}
			}
		}

		searchRow += direction
	}

	*lastFoundRow, *lastFoundCol = -1, -1
}
