package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createTasksPage(tasks []dao.Task) tview.Flex {
	data := CreateTasksData(tasks)

	tasksTable := createTasksTable(&data)
	selectedList := createTasksSelectedList(&data)

	// Context
	data.runContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow)
	data.runContextPage.AddItem(selectedList.List, 0, 1, true)

	data.tasksPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tasksTable.Table, 0, 1, true).
				AddItem(data.runContextPage, 30, 1, false),
			0, 1, true).
		AddItem(TUI.search, 1, 0, false)

	focusableElements := []tview.Primitive{tasksTable.Table}
	focusableElements = append(focusableElements, selectedList.List)

	currentFocus := 0
	// Handle global shortcuts
	data.tasksPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
				TUI.app.SetFocus(tasksTable.Table)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			case '2': // Tags focus
				TUI.app.SetFocus(selectedList.List)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			// case '3': // Paths focus
			// 	TUI.app.SetFocus(pathsList.List)
			// 	currentFocus = getCurrentFocusIndex(focusableElements)
			// 	return nil
			// case '4': // Selected focus
			// 	TUI.app.SetFocus(selectedList.List)
			// 	currentFocus = getCurrentFocusIndex(focusableElements)
			// 	return nil
			case 'a': // Select all
				TUI.emitter.Publish(Event{Name: "select_all_tasks", Data: ""})
				return nil
			case 'c': // Unselect all all
				TUI.emitter.Publish(Event{Name: "deselect_all_tasks", Data: ""})
				return nil
			case 'f': // Clear filters
				TUI.emitter.PublishAndWait(Event{Name: "clear_filters", Data: ""})
				TUI.emitter.Publish(Event{Name: "filter_tasks", Data: ""})
				return nil
			}
		}
		return event
	})

	return *data.tasksPage
}
