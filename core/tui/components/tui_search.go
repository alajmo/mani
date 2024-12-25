package components

import (
	"strings"

	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

func CreateSearch() *tview.InputField {
	search := tview.NewInputField().
		SetLabel("").
		SetLabelStyle(misc.STYLE_SEARCH_LABEL.Style).
		SetFieldStyle(misc.STYLE_SEARCH_TEXT.Style)
	return search
}

func ShowSearch() {
	misc.Search.SetLabel(misc.Colorize("Search:", *misc.TUITheme.SearchLabel))
	misc.Search.SetText("")
	misc.App.SetFocus(misc.Search)
}

func EmptySearch() {
	misc.Search.SetLabel("")
	misc.Search.SetText("")
}

func SearchInTable(table *tview.Table, query string, lastFoundRow, lastFoundCol *int, direction int) {
	query = strings.ToLower(query)
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
				if strings.Contains(strings.ToLower(strings.TrimSpace(cell.Text)), query) {
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

func SearchInTree(tree *TTree, query string, lastFoundIndex *int, direction int) {
	query = strings.ToLower(query)
	itemCount := len(tree.List)
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

		name := strings.ToLower(tree.List[searchIndex].DisplayName)
		if strings.Contains(name, query) {
			tree.Tree.SetCurrentNode(tree.List[searchIndex].TreeNode)
			*lastFoundIndex = searchIndex
			return
		}

		searchIndex += direction
	}

	*lastFoundIndex = -1

}

func SearchInList(list *tview.List, query string, lastFoundIndex *int, direction int) {
	query = strings.ToLower(query)
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
		if strings.Contains(strings.ToLower(mainText), query) ||
			strings.Contains(strings.ToLower(secondaryText), query) {
			list.SetCurrentItem(searchIndex)
			*lastFoundIndex = searchIndex
			return
		}

		searchIndex += direction
	}

	*lastFoundIndex = -1
}
