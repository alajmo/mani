package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createProjectsPage() {
	data := TUIProjects{}

	// Poulate project data
	TUI.projectsFiltered = TUI.projects

	projectsTable := createProjectsTable(TUI.projects)
	tagsList := createProjectsTagsList()
	pathsList := createProjectsPathsList()
	selectedList := createProjectsSelectedList()

	// Projects context
	TUI.projectsContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow)

	if tagsList.List.GetItemCount() > 0 {
		TUI.projectsContextPage.AddItem(tagsList.List, 0, 1, true)
	}
	if pathsList.List.GetItemCount() > 0 {
		TUI.projectsContextPage.AddItem(pathsList.List, 0, 1, true)
	}
	TUI.projectsContextPage.AddItem(selectedList.List, 0, 1, true)

	TUI.projectsPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(projectsTable.Table, 0, 1, true).
				AddItem(TUI.projectsContextPage, 30, 1, false),
			0, 1, true).
		AddItem(TUI.search, 1, 0, false)

	focusableElements := []tview.Primitive{projectsTable.Table}
	if len(TUI.projectTags) > 0 {
		focusableElements = append(focusableElements, tagsList.List)
	}
	if len(TUI.projectPaths) > 0 {
		focusableElements = append(focusableElements, pathsList.List)
	}
	focusableElements = append(focusableElements, selectedList.List)

	currentFocus := 0
	// Handle global shortcuts
	TUI.projectsPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if TUI.app.GetFocus() == TUI.search {
			return event
		}

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
			case '1': // Table focus
				TUI.app.SetFocus(projectsTable.Table)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			case '2': // Tags focus
				TUI.app.SetFocus(tagsList.List)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			case '3': // Paths focus
				TUI.app.SetFocus(pathsList.List)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			case '4': // Selected focus
				TUI.app.SetFocus(selectedList.List)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			case 'a': // Select all
				TUI.emitter.Publish(Event{Name: "select_all_projects", Data: ""})
				return nil
			case 'c': // Unselect all all
				TUI.emitter.Publish(Event{Name: "deselect_all_projects", Data: ""})
				return nil
			case 'f': // Clear filters
				TUI.emitter.PublishAndWait(Event{Name: "clear_filters", Data: ""})
				TUI.emitter.Publish(Event{Name: "filter_projects", Data: ""})
				return nil
			}
		}
		return event
	})
}
