package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"
)

func CreateTasksPage(tasks []dao.Task) *tview.Flex {
	data := views.CreateTasksData(tasks, []string{"Name", "Description"}, true)

	tasksTable := views.CreateTasksTable(&data, false, "")
	// selectedList := views.CreateTasksSelectedList(&data)

	// Context
	// data.RunContextPage = tview.NewFlex().
	// 	SetDirection(tview.FlexRow)
	// data.RunContextPage.AddItem(selectedList.List, 0, 1, true)

	data.TasksPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tasksTable.Table, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := []tview.Primitive{tasksTable.Table}

	// Handle global shortcuts
	data.TasksPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyTab:
			misc.FocusNext(focusableElements)
			return nil
		case tcell.KeyBacktab:
			misc.FocusPrevious(focusableElements)
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case '1': // Table focus
				misc.App.SetFocus(tasksTable.Table)
				return nil
				// case 'f': // Clear filters
				// 	data.Emitter.PublishAndWait(misc.Event{Name: "clear_filters", Data: ""})
				// 	data.Emitter.Publish(misc.Event{Name: "filter_tasks", Data: ""})
				// 	return nil
			}
		}
		return event
	})

	return data.TasksPage
}
