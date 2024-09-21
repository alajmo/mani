package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"
)

func CreateTasksPage(tasks []dao.Task) *tview.Flex {
	data := views.CreateTasksData(tasks)

	tasksTable := views.CreateTasksTable(&data)
	selectedList := views.CreateTasksSelectedList(&data)

	// Context
	data.RunContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow)
	data.RunContextPage.AddItem(selectedList.List, 0, 1, true)

	data.TasksPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tasksTable.Table, 0, 1, true).
				AddItem(data.RunContextPage, 30, 1, false),
			0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := []tview.Primitive{tasksTable.Table}
	focusableElements = append(focusableElements, selectedList.List)

	currentFocus := 0
	// Handle global shortcuts
	data.TasksPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyTab:
			currentFocus = (currentFocus + 1) % len(focusableElements)
			misc.App.SetFocus(focusableElements[currentFocus])
			return nil
		case tcell.KeyBacktab:
			currentFocus = (currentFocus - 1 + len(focusableElements)) % len(focusableElements)
			misc.App.SetFocus(focusableElements[currentFocus])
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case '1': // Table focus
				misc.App.SetFocus(tasksTable.Table)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '2': // Tags focus
				misc.App.SetFocus(selectedList.List)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			// case '3': // Paths focus
			// 	misc.App.SetFocus(pathsList.List)
			// 	currentFocus = getCurrentFocusIndex(focusableElements)
			// 	return nil
			// case '4': // Selected focus
			// 	misc.App.SetFocus(selectedList.List)
			// 	currentFocus = getCurrentFocusIndex(focusableElements)
			// 	return nil
			case 'a': // Select all
				misc.Emitter.Publish(misc.Event{Name: "select_all_tasks", Data: ""})
				return nil
			case 'c': // Unselect all all
				misc.Emitter.Publish(misc.Event{Name: "deselect_all_tasks", Data: ""})
				return nil
			case 'f': // Clear filters
				misc.Emitter.PublishAndWait(misc.Event{Name: "clear_filters", Data: ""})
				misc.Emitter.Publish(misc.Event{Name: "filter_tasks", Data: ""})
				return nil
			}
		}
		return event
	})

	return data.TasksPage
}
