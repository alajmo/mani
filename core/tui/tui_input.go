package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
)

func HandleInput() {
	var lastSearchQuery string
	var lastFoundRow, lastFoundCol int
	searchDirection := 1

	misc.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFocus := misc.App.GetFocus()

		// Search
		if currentFocus == misc.Search {
			lastFoundRow, lastFoundCol = -1, -1
			switch event.Key() {
			case tcell.KeyEscape:
				components.EmptySearch()
				misc.FocusPreviousPage()
				return nil
			case tcell.KeyEnter:
				return handleSearchInput(event, searchDirection, &lastFoundRow, &lastFoundCol)
			}

			return event
		}

		// If NewInputField
		if _, ok := currentFocus.(*tview.InputField); ok {
			return event
		}
		if _, ok := currentFocus.(*tview.TextArea); ok {
			return event
		}

		// Modal
		if components.IsModalOpen() {
			switch event.Key() {
			case tcell.KeyEscape:
				components.CloseModal()
				misc.FocusPreviousPage()
				return nil
			case tcell.KeyRune:
				switch event.Rune() {
				case 'q':
					misc.App.Stop()
					return nil
				}
			}
			return event
		}

		// Main
		switch event.Key() {
		case tcell.KeyEscape:
			components.EmptySearch()
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				misc.App.Stop()
				return nil
			case 'p':
				misc.SwitchToPage("projects")
				misc.App.SetFocus(*misc.ProjectsLastFocus)
				return nil
			case 't':
				misc.SwitchToPage("tasks")
				misc.App.SetFocus(*misc.TasksLastFocus)
				return nil
			case 'r':
				misc.SwitchToPage("run")
				misc.App.SetFocus(*misc.RunLastFocus)
				return nil
			case 'e':
				misc.SwitchToPage("exec")
				misc.App.SetFocus(*misc.ExecLastFocus)
				return nil
			case '?':
				components.ShowHelpModal()
				return nil
			case '/':
				components.ShowSearch()
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

	misc.Search.SetChangedFunc(func(query string) {
		if query != lastSearchQuery {
			lastSearchQuery = query
			lastFoundRow, lastFoundCol = -1, -1
			searchDirection = 1

			switch prevPage := misc.PreviousPage.(type) {
			case *tview.Table:
				components.SearchInTable(prevPage, query, &lastFoundRow, &lastFoundCol, searchDirection)
			case *tview.List:
				components.SearchInList(prevPage, query, &lastFoundRow, searchDirection)
			}
		}
	})
}

func handleSearchInput(event *tcell.EventKey, searchDirection int, lastFoundRow *int, lastFoundCol *int) *tcell.EventKey {
	query := misc.Search.GetText()
	if query == "" {
		return nil
	}

	switch prevPage := misc.PreviousPage.(type) {
	case *tview.Table:
		misc.App.SetFocus(prevPage)
		components.SearchInTable(prevPage, query, lastFoundRow, lastFoundCol, searchDirection)
	case *tview.List:
		misc.App.SetFocus(prevPage)
		components.SearchInList(prevPage, query, lastFoundRow, searchDirection)
	}

	return nil
}
