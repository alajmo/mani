package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func handleInput() {
	var lastSearchQuery string
	var lastFoundRow, lastFoundCol int
	searchDirection := 1

	TUI.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Search
		if TUI.app.GetFocus() == TUI.search {
			lastFoundRow, lastFoundCol = -1, -1
			switch event.Key() {
			case tcell.KeyEscape:
				emptySearch()
				focusPreviousPage()
				return nil
			case tcell.KeyEnter:
				return handleSearchInput(event, searchDirection, &lastFoundRow, &lastFoundCol)
			}

			return event
		}

		// Modal
		if isModalOpen() {
			switch event.Key() {
			case tcell.KeyEscape:
				closeModal()
				focusPreviousPage()
				return nil
			}
			return event
		}

		// Main
		switch event.Key() {
		case tcell.KeyEscape:
			emptySearch()
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				TUI.app.Stop()
				return nil
			case 'p':
				switchToPage("projects")
				return nil
			case 't':
				switchToPage("tasks")
				return nil
			case 'r':
				switchToPage("run")
				return nil
			case '?':
				showHelpModal()
				return nil
			case '/':
				showSearch()
				return nil
			case 'n':
				searchDirection = 1
				return handleSearchInput(event, searchDirection, &lastFoundRow, &lastFoundCol)
			case 'N':
				searchDirection = -1
				return handleSearchInput(event, searchDirection, &lastFoundRow, &lastFoundCol)
			}
		}

		return event
	})

	TUI.search.SetChangedFunc(func(query string) {
		if query != lastSearchQuery {
			lastSearchQuery = query
			lastFoundRow, lastFoundCol = -1, -1
			searchDirection = 1

			switch prevPage := TUI.previousPage.(type) {
			case *tview.Table:
				searchInTable(prevPage, query, &lastFoundRow, &lastFoundCol, searchDirection)
			case *tview.List:
				searchInList(prevPage, query, &lastFoundRow, searchDirection)
			}
		}
	})
}

func handleSearchInput(event *tcell.EventKey, searchDirection int, lastFoundRow *int, lastFoundCol *int) *tcell.EventKey {
	query := TUI.search.GetText()
	if query == "" {
		return nil
	}

	switch prevPage := TUI.previousPage.(type) {
	case *tview.Table:
		TUI.app.SetFocus(prevPage)
		searchInTable(prevPage, query, lastFoundRow, lastFoundCol, searchDirection)
	case *tview.List:
		TUI.app.SetFocus(prevPage)
		searchInList(prevPage, query, lastFoundRow, searchDirection)
	}

	return nil
}
