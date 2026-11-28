package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"

	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"
)

func HandleInput(app *App) {
	var lastSearchQuery string
	var lastFoundRow, lastFoundCol int
	searchDirection := 1

	misc.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFocus := misc.App.GetFocus()

		switch event.Key() {
		case tcell.KeyF1:
			SwitchToPage("run")
			misc.App.SetFocus(*misc.RunLastFocus)
			return nil
		case tcell.KeyF2:
			SwitchToPage("exec")
			misc.App.SetFocus(*misc.ExecLastFocus)
			return nil
		case tcell.KeyF3:
			SwitchToPage("projects")
			misc.App.SetFocus(*misc.ProjectsLastFocus)
			return nil
		case tcell.KeyF4:
			SwitchToPage("tasks")
			misc.App.SetFocus(*misc.TasksLastFocus)
			return nil
		case tcell.KeyF5:
			go app.Reload()
			return nil
		case tcell.KeyF6:
			misc.App.Sync()
			return nil
		}

		// Modal
		if components.IsModalOpen() {
			switch event.Key() {
			case tcell.KeyEscape:
				components.CloseModal()
				return nil
			case tcell.KeyRune:
				switch event.Rune() {
				case 'q':
					// Close the front modal (like Escape) rather than quitting the
					// app, so a stray 'q' in a modal doesn't kill the TUI.
					components.CloseModal()
					return nil
				}
			}
			return event
		}

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

		// Input
		if _, ok := currentFocus.(*tview.InputField); ok {
			return event
		}
		// TextArea
		if _, ok := currentFocus.(*tview.TextArea); ok {
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
				// Quitting would leave the commands of a run behind, running
				// without anything to show their output. Say so instead:
				// Ctrl-C still quits, for a task that doesn't finish.
				if misc.RunActive() {
					msg := "A task is still running.\nWait for it to finish, or press Ctrl-C to quit and leave its commands running."
					// The modal is sized by the text without colors: make
					// room for the padding around the text as well, which
					// would otherwise hide all but the first line
					size := strings.ReplaceAll(msg, "\n", "      \n") + "      \n.\n."
					components.OpenTextModal("run-active-modal", msg, size, "Running")
					return nil
				}
				misc.App.Stop()
				return nil
			case 'R':
				misc.App.Sync()
				return nil
			case 'p':
				SwitchToPage("projects")
				misc.App.SetFocus(*misc.ProjectsLastFocus)
				return nil
			case 't':
				SwitchToPage("tasks")
				misc.App.SetFocus(*misc.TasksLastFocus)
				return nil
			case 'r':
				SwitchToPage("run")
				misc.App.SetFocus(*misc.RunLastFocus)
				return nil
			case 'e':
				SwitchToPage("exec")
				misc.App.SetFocus(*misc.ExecLastFocus)
				return nil
			case '?':
				views.ShowHelpModal()
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

			switch prevPage := misc.PreviousPane.(type) {
			case *tview.Table:
				components.SearchInTable(prevPage, query, &lastFoundRow, &lastFoundCol, searchDirection)
			case *tview.TreeView:
				if tree, ok := misc.PreviousModel.(*components.TTree); ok {
					components.SearchInTree(tree, query, &lastFoundRow, searchDirection)
				}
			case *tview.List:
				components.SearchInList(prevPage, query, &lastFoundRow, searchDirection)
			}
		}
	})
}

func handleSearchInput(_ *tcell.EventKey, searchDirection int, lastFoundRow *int, lastFoundCol *int) *tcell.EventKey {
	query := misc.Search.GetText()
	if query == "" {
		return nil
	}

	switch prevPage := misc.PreviousPane.(type) {
	case *tview.Table:
		misc.App.SetFocus(prevPage)
		components.SearchInTable(prevPage, query, lastFoundRow, lastFoundCol, searchDirection)
	case *tview.TreeView:
		misc.App.SetFocus(prevPage)
		if tree, ok := misc.PreviousModel.(*components.TTree); ok {
			components.SearchInTree(tree, query, lastFoundRow, searchDirection)
		}
	case *tview.List:
		misc.App.SetFocus(prevPage)
		components.SearchInList(prevPage, query, lastFoundRow, searchDirection)
	}

	return nil
}
