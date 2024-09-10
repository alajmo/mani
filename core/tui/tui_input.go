package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func handleInput() {
	focusableElements := []tview.Primitive{
		TUI.mainPage,
		TUI.projectsTagsPane,
		TUI.projectsPathsPane,
		TUI.projectsSelectedPane,
	}

	currentFocus := 0
	var lastSearchQuery string
	var lastFoundRow, lastFoundCol int
	searchDirection := 1

	TUI.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Search Input
		if TUI.app.GetFocus() == TUI.search {
			lastFoundRow, lastFoundCol = -1, -1

			switch event.Key() {
			case tcell.KeyEscape:
				hideSearch()
				return nil

			case tcell.KeyEnter:
				{
					query := TUI.search.GetText()
					if query == "" {
						return nil
					}

					switch TUI.previousPage {
					case TUI.projectsTable:
						TUI.app.SetFocus(TUI.projectsTable)
						searchInTable(TUI.projectsTable, query, &lastFoundRow, &lastFoundCol, searchDirection)

					case TUI.projectsTagsPane:
						TUI.app.SetFocus(TUI.projectsTagsPane)
						searchInList(TUI.projectsTagsPane, query, &lastFoundRow, searchDirection)

					case TUI.projectsPathsPane:
						TUI.app.SetFocus(TUI.projectsPathsPane)
						searchInList(TUI.projectsPathsPane, query, &lastFoundRow, searchDirection)

					case TUI.projectsSelectedPane:
						TUI.app.SetFocus(TUI.projectsSelectedPane)
						searchInList(TUI.projectsSelectedPane, query, &lastFoundRow, searchDirection)
					}

					return nil
				}
			}
			return event
		}

		// TODO: Check if open modal, then only allow escape
		if isModalOpen() {
			switch event.Key() {
			case tcell.KeyEscape:
				closeModal()
				return nil
			}

			return nil
		}

		// Main
		switch event.Key() {
		case tcell.KeyTab:
			currentFocus = (currentFocus + 1) % len(focusableElements)
			TUI.app.SetFocus(focusableElements[currentFocus])
			return nil
		case tcell.KeyBacktab:
			currentFocus = (currentFocus - 1 + len(focusableElements)) % len(focusableElements)
			TUI.app.SetFocus(focusableElements[currentFocus])
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				TUI.app.Stop()
				return nil
			case 'p', '1':
				switchToPage("projects")
				return nil
			case 't', '2':
				switchToPage("tasks")
				return nil
			case 'r', '3':
				switchToPage("run")
				return nil
			case '?', '4':
				showHelpModal()
				return nil
			case '/':
				showSearch()
				return nil
			case 'n':
				{
					query := TUI.search.GetText()
					if query == "" {
						return nil
					}
					searchDirection = 1

					switch TUI.app.GetFocus() {
					case TUI.projectsTable:
						searchInTable(TUI.projectsTable, query, &lastFoundRow, &lastFoundCol, searchDirection)
						return nil
					case TUI.projectsTagsPane:
						searchInList(TUI.projectsTagsPane, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsPathsPane:
						searchInList(TUI.projectsPathsPane, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsSelectedPane:
						searchInList(TUI.projectsSelectedPane, query, &lastFoundRow, searchDirection)
						return nil
					}
				}

			case 'N':
				{
					query := TUI.search.GetText()
					if query == "" {
						return nil
					}
					searchDirection = -1

					switch TUI.app.GetFocus() {
					case TUI.projectsTable:
						searchInTable(TUI.projectsTable, query, &lastFoundRow, &lastFoundCol, searchDirection)
						return nil
					case TUI.projectsTagsPane:
						searchInList(TUI.projectsTagsPane, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsPathsPane:
						searchInList(TUI.projectsPathsPane, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsSelectedPane:
						searchInList(TUI.projectsSelectedPane, query, &lastFoundRow, searchDirection)
						return nil
					}
				}
			}
		}

		return event
	})

	TUI.search.SetChangedFunc(func(text string) {
		if text != lastSearchQuery {
			lastSearchQuery = text
			lastFoundRow, lastFoundCol = -1, -1
			searchDirection = 1

			switch TUI.previousPage {
			case TUI.projectsTable:
				searchInTable(TUI.projectsTable, text, &lastFoundRow, &lastFoundCol, searchDirection)
			case TUI.projectsTagsPane:
				searchInList(TUI.projectsTagsPane, text, &lastFoundRow, searchDirection)
			case TUI.projectsPathsPane:
				searchInList(TUI.projectsPathsPane, text, &lastFoundRow, searchDirection)
			case TUI.projectsSelectedPane:
				searchInList(TUI.projectsSelectedPane, text, &lastFoundRow, searchDirection)
			}
		}
	})
}
