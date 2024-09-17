package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createTasksPage() {
	// Data
	TUI.tasksFiltered = TUI.tasks

	tasksTable := createTasksTable(TUI.tasks)
	selectedList := createTasksSelectedList()

	// Context
	TUI.runContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow)
	TUI.runContextPage.AddItem(selectedList.List, 0, 1, true)

	TUI.tasksPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tasksTable.Table, 0, 1, true).
				AddItem(TUI.runContextPage, 30, 1, false),
			0, 1, true).
		AddItem(TUI.search, 1, 0, false)

	focusableElements := []tview.Primitive{tasksTable.Table}
	focusableElements = append(focusableElements, selectedList.List)

	currentFocus := 0
	// Handle global shortcuts
	TUI.tasksPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
}

func createTasksTable(tasks []dao.Task) TUITable {
	table := TUITable{}
	table.createTable()
	TUI.tasksTable = table.Table
	TUI.previousPage = TUI.tasksTable

	// Methods
	table.IsRowSelected = func(name string) bool {
		return isTaskSelected(TUI.tasksSelected, name)
	}
	table.EditRow = func(taskName string) {
		editTask(taskName)
	}
	table.ToggleSelected = func() {
		i, _ := table.Table.GetSelection()
		taskName := table.Table.GetCell(i, 0).Text
		isSelected := isTaskSelected(TUI.tasksSelected, taskName)
		if isSelected {
			TUI.tasksSelected = removeTask(TUI.tasksSelected, taskName)
		} else {
			task := getTask(TUI.tasks, taskName)
			TUI.tasksSelected = append(TUI.tasksSelected, task)
		}
		TUI.emitter.Publish(Event{Name: "toggle_selected_task", Data: taskName})
		table.updateCellStyles()
	}
	table.SelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			taskName := table.Table.GetCell(i, 0).Text
			if !isTaskSelected(TUI.tasksSelected, taskName) {
				task := getTask(TUI.tasks, taskName)
				TUI.tasksSelected = append(TUI.tasksSelected, task)
			}
		}
		TUI.emitter.Publish(Event{Name: "update_all_selected_tasks", Data: ""})
		table.updateCellStyles()
	}
	table.DeSelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			taskName := table.Table.GetCell(i, 0).Text
			TUI.tasksSelected = removeTask(TUI.tasksSelected, taskName)
		}
		TUI.emitter.Publish(Event{Name: "update_all_selected_tasks", Data: ""})
		table.updateCellStyles()
	}
	table.DescribeRow = func() {
		row, _ := table.Table.GetSelection()
		if row > 0 {
			showTaskDescModal(tasks[row-1])
		}
	}

	// Events
	TUI.emitter.Subscribe("filter_tasks", func(e Event) {
		table.filterTasks()
	})
	TUI.emitter.Subscribe("remove_selected_task", func(e Event) {
		table.updateTasksTable()
	})
	TUI.emitter.Subscribe("select_all_tasks", func(e Event) {
		table.SelectAllRows()
	})
	TUI.emitter.Subscribe("deselect_all_tasks", func(e Event) {
		table.DeSelectAllRows()
	})

	table.updateTasksTable()

	return table
}

func createTasksSelectedList() TUIList {
	list := TUIList{Title: "Selected", Items: make(map[string]bool)}
	list.createList()
	TUI.tasksSelectedPane = list.List

	// Methods
	updateSelectedTasks := func() {
		list.List.Clear()
		for _, task := range TUI.tasksSelected {
			list.List.AddItem(task.Name, task.Name, 0, nil)
		}

		if list.List.HasFocus() {
			list.setActive(true)
		} else {
			list.setActive(false)
		}
	}
	toggleSelectedTask := func(taskName string) {
		items := list.List.FindItems(taskName, taskName, false, false)
		if len(items) == 0 {
			list.List.AddItem(taskName, taskName, 0, nil)
		} else {
			list.List.RemoveItem(items[0])
		}

		if list.List.HasFocus() {
			list.setActive(true)
		} else {
			list.setActive(false)
		}
	}
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		taskName, _ := list.List.GetItemText(i)
		TUI.tasksSelected = removeTask(TUI.tasksSelected, taskName)
		toggleSelectedTask(taskName)

		TUI.emitter.Publish(Event{Name: "remove_selected_task", Data: ""})
	}

	// Events
	TUI.emitter.Subscribe("toggle_selected_task", func(e Event) {
		toggleSelectedTask(e.Data.(string))
	})

	TUI.emitter.Subscribe("update_all_selected_tasks", func(e Event) {
		updateSelectedTasks()
	})

	return list
}

func (t *TUITable) updateTasksTable() {
	t.Table.Clear()

	// Set up headers
	headers := []string{"Name", "Description"}
	for col, header := range headers {
		t.Table.SetCell(0, col, createTableHeader(header))
	}

	// Populate the table with task data
	for row, task := range TUI.tasksFiltered {
		t.Table.SetCell(row+1, 0, tview.NewTableCell(task.Name))
		t.Table.SetCell(row+1, 1, tview.NewTableCell(task.Desc))
	}

	t.updateCellStyles()
}

func (t *TUITable) filterTasks() {
	// projectTags := []string{}
	// for key, filtered := range TUI.tasksTagsFiltered {
	// 	if filtered {
	// 		projectTags = append(projectTags, key)
	// 	}
	// }

	// projectPaths := []string{}
	// for key, filtered := range TUI.tasksPathsFiltered {
	// 	if filtered {
	// 		projectPaths = append(projectPaths, key)
	// 	}
	// }

	// if len(projectPaths) > 0 || len(projectTags) > 0 {
	// 	projects, _ := TUI.config.FilterProjects(false, false, []string{}, projectPaths, projectTags)
	// 	TUI.tasksFiltered = projects
	// } else {
	// 	TUI.tasksFiltered = TUI.tasks
	// }

	// t.updateTaskTable()
	// t.Table.ScrollToBeginning()
	// t.Table.Select(1, 0)
}

func showTaskDescModal(task dao.Task) {
	description := print.PrintTaskBlock([]dao.Task{task})
	openModal("task-description-modal", description, task.Name, 80, 30)
}

func editTask(taskName string) {
	TUI.app.Suspend(func() {
		err := TUI.config.EditTask(taskName)
		if err != nil {
			return
		}
	})
}
