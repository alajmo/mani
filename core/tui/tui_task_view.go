package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/rivo/tview"
)

type TUITasks struct {
	// UI
	tasksPage         *tview.Flex
	tasksTable        *tview.Table
	runContextPage    *tview.Flex
	tasksSelectedPane *tview.List

	// Data
	tasks         []dao.Task
	tasksFiltered []dao.Task
	tasksSelected []dao.Task
}

func CreateTasksData(tasks []dao.Task) TUITasks {
	data := TUITasks{
		tasks:         tasks,
		tasksFiltered: tasks,
		tasksSelected: []dao.Task{},
	}

	return data
}

func createTasksTable(data *TUITasks) TUITable {
	table := TUITable{}
	table.createTable()
	data.tasksTable = table.Table
	TUI.previousPage = data.tasksTable

	// Methods
	table.IsRowSelected = func(name string) bool {
		return isTaskSelected(data.tasksSelected, name)
	}
	table.EditRow = func(taskName string) {
		editTask(taskName)
	}
	table.ToggleSelected = func() {
		i, _ := table.Table.GetSelection()
		taskName := table.Table.GetCell(i, 0).Text
		isSelected := isTaskSelected(data.tasksSelected, taskName)
		if isSelected {
			data.tasksSelected = removeTask(data.tasksSelected, taskName)
		} else {
			task := getTask(data.tasks, taskName)
			data.tasksSelected = append(data.tasksSelected, task)
		}
		TUI.emitter.Publish(Event{Name: "toggle_selected_task", Data: taskName})
		table.updateCellStyles()
	}
	table.SelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			taskName := table.Table.GetCell(i, 0).Text
			if !isTaskSelected(data.tasksSelected, taskName) {
				task := getTask(data.tasks, taskName)
				data.tasksSelected = append(data.tasksSelected, task)
			}
		}
		TUI.emitter.Publish(Event{Name: "update_all_selected_tasks", Data: ""})
		table.updateCellStyles()
	}
	table.DeSelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			taskName := table.Table.GetCell(i, 0).Text
			data.tasksSelected = removeTask(data.tasksSelected, taskName)
		}
		TUI.emitter.Publish(Event{Name: "update_all_selected_tasks", Data: ""})
		table.updateCellStyles()
	}
	table.DescribeRow = func() {
		row, _ := table.Table.GetSelection()
		if row > 0 {
			showTaskDescModal(data.tasks[row-1])
		}
	}

	// Events
	TUI.emitter.Subscribe("filter_tasks", func(e Event) {
		table.filterTasks()
	})
	TUI.emitter.Subscribe("remove_selected_task", func(e Event) {
		table.updateTasksTable(data)
	})
	TUI.emitter.Subscribe("select_all_tasks", func(e Event) {
		table.SelectAllRows()
	})
	TUI.emitter.Subscribe("deselect_all_tasks", func(e Event) {
		table.DeSelectAllRows()
	})

	table.updateTasksTable(data)

	return table
}

func createTasksSelectedList(data *TUITasks) TUIList {
	list := TUIList{Title: "Selected", Items: make(map[string]bool)}
	list.createList()
	data.tasksSelectedPane = list.List

	// Methods
	updateSelectedTasks := func() {
		list.List.Clear()
		for _, task := range data.tasksSelected {
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
		data.tasksSelected = removeTask(data.tasksSelected, taskName)
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

func (t *TUITable) updateTasksTable(data *TUITasks) {
	t.Table.Clear()

	// Set up headers
	headers := []string{"Name", "Description"}
	for col, header := range headers {
		t.Table.SetCell(0, col, createTableHeader(header))
	}

	// Populate the table with task data
	for row, task := range data.tasksFiltered {
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
