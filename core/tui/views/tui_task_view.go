package views

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

type TUITasks struct {
	// UI
	TasksPage         *tview.Flex
	TasksTable        *tview.Table
	RunContextPage    *tview.Flex
	TasksSelectedPane *tview.List

	// Data
	Tasks         []dao.Task
	TasksFiltered []dao.Task
	TasksSelected []dao.Task
	TaskHeaders   []string
	ShowHeaders   bool

	Emitter *misc.EventEmitter
}

func CreateTasksData(tasks []dao.Task, headers []string, showHeaders bool) TUITasks {
	data := TUITasks{
		Tasks:         tasks,
		TasksFiltered: tasks,
		TasksSelected: []dao.Task{},
		TaskHeaders:   headers,
		ShowHeaders:   showHeaders,

		Emitter: misc.NewEventEmitter(),
	}

	return data
}

func CreateTasksTable(data *TUITasks, selectEnabled bool, title string) components.TUITable {
	table := components.TUITable{SelectEnabled: selectEnabled, Title: title}
	table.CreateTable()
	data.TasksTable = table.Table
	misc.PreviousPage = data.TasksTable

	// Methods
	table.IsRowSelected = func(name string) bool {
		return misc.IsTaskSelected(data.TasksSelected, name)
	}
	table.EditRow = func(taskName string) {
		editTask(taskName)
	}
	table.ToggleSelected = func() {
		i, _ := table.Table.GetSelection()
		taskName := table.Table.GetCell(i, 0).Text
		isSelected := misc.IsTaskSelected(data.TasksSelected, taskName)
		if isSelected {
			data.TasksSelected = misc.RemoveTask(data.TasksSelected, taskName)
		} else {
			task := misc.GetTask(data.Tasks, taskName)
			data.TasksSelected = append(data.TasksSelected, task)
		}
		data.Emitter.Publish(misc.Event{Name: "toggle_selected_task", Data: taskName})
		table.UpdateCellStyles()
	}
	table.SelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			taskName := table.Table.GetCell(i, 0).Text
			if !misc.IsTaskSelected(data.TasksSelected, taskName) {
				task := misc.GetTask(data.Tasks, taskName)
				data.TasksSelected = append(data.TasksSelected, task)
			}
		}
		data.Emitter.Publish(misc.Event{Name: "update_all_selected_tasks", Data: ""})
		table.UpdateCellStyles()
	}
	table.DeSelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			taskName := table.Table.GetCell(i, 0).Text
			data.TasksSelected = misc.RemoveTask(data.TasksSelected, taskName)
		}
		data.Emitter.Publish(misc.Event{Name: "update_all_selected_tasks", Data: ""})
		table.UpdateCellStyles()
	}
	table.DescribeRow = func() {
		row, _ := table.Table.GetSelection()
		if row > 0 {
			showTaskDescModal(data.Tasks[row-1])
		}
	}

	// Events
	data.Emitter.Subscribe("filter_tasks", func(e misc.Event) {
		filterTasks(&table)
	})
	data.Emitter.Subscribe("remove_selected_task", func(e misc.Event) {
		UpdateTasksTable(&table, data)
	})
	data.Emitter.Subscribe("select_all_tasks", func(e misc.Event) {
		table.SelectAllRows()
	})
	data.Emitter.Subscribe("deselect_all_tasks", func(e misc.Event) {
		table.DeSelectAllRows()
	})

	UpdateTasksTable(&table, data)

	return table
}

func CreateTasksSelectedList(data *TUITasks, title string) components.TUIList {
	list := components.TUIList{Title: title, Items: make(map[string]bool)}
	list.CreateList()
	data.TasksSelectedPane = list.List

	// Methods
	updateSelectedTasks := func() {
		list.List.Clear()
		for _, task := range data.TasksSelected {
			list.List.AddItem(task.Name, task.Name, 0, nil)
		}

		if list.List.HasFocus() {
			list.SetActive(true)
		} else {
			list.SetActive(false)
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
			list.SetActive(true)
		} else {
			list.SetActive(false)
		}
	}
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		taskName, _ := list.List.GetItemText(i)
		data.TasksSelected = misc.RemoveTask(data.TasksSelected, taskName)
		toggleSelectedTask(taskName)

		data.Emitter.Publish(misc.Event{Name: "remove_selected_task", Data: ""})
	}

	// Events
	data.Emitter.Subscribe("toggle_selected_task", func(e misc.Event) {
		toggleSelectedTask(e.Data.(string))
	})

	data.Emitter.Subscribe("update_all_selected_tasks", func(e misc.Event) {
		updateSelectedTasks()
	})

	return list
}

func UpdateTasksTable(t *components.TUITable, data *TUITasks) {
	t.Table.Clear()

	// Set up headers
	for col, header := range data.TaskHeaders {
		if data.ShowHeaders {
			t.Table.SetCell(0, col, components.CreateTableHeader(header))
		} else {
			t.Table.SetCell(0, col, components.CreateTableHeader(""))
		}
	}

	// Populate the table with task data
	for row, task := range data.TasksFiltered {
		for col, header := range data.TaskHeaders {
			t.Table.SetCell(row+1, col, tview.NewTableCell(task.GetValue(header, 0)))
		}
	}

	t.UpdateCellStyles()
}

func filterTasks(t *components.TUITable) {
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
	description := print.PrintTaskBlock([]dao.Task{task}, true)
	components.OpenTextModal("task-description-modal", description, task.Name, 120, 30)
}

func editTask(taskName string) {
	misc.App.Suspend(func() {
		err := misc.Config.EditTask(taskName)
		if err != nil {
			return
		}
	})
}
