package pages

import (
	"github.com/alajmo/mani/core/tui/components"
	"github.com/rivo/tview"
)

func CreateRunPage() *tview.Flex {
	runPage := tview.NewFlex()

	return runPage
	// // Data
	// TUI.tasksFiltered = TUI.tasks

	// // tasksTable := createRunTable(TUI.tasks)

	// // Context
	// TUI.runContextPage = tview.NewFlex().
	// 	SetDirection(tview.FlexRow)
	// // TUI.tasksContextPage.AddItem(selectedList.List, 0, 1, true)

	// TUI.tasksPage = tview.NewFlex().
	// 	SetDirection(tview.FlexRow).
	// 	AddItem(
	// 		tview.NewFlex().SetDirection(tview.FlexColumn).
	// 			// AddItem(tasksTable.Table, 0, 1, true).
	// 			AddItem(TUI.runContextPage, 30, 1, false),
	// 		0, 1, true).
	// 	AddItem(TUI.search, 1, 0, false)

	// // focusableElements := []tview.Primitive{tasksTable.Table}
	// // focusableElements = append(focusableElements, selectedList.List)
}

func createRunTable() components.TUITable {
	table := components.TUITable{}
	// table.createTable()
	// TUI.tasksTable = table.Table
	// TUI.previousPage = TUI.tasksTable

	// // Methods
	// table.IsRowSelected = func(name string) bool {
	// 	return isTaskSelected(TUI.tasksSelected, name)
	// }
	// table.EditRow = func(taskName string) {
	// 	editTask(taskName)
	// }
	// table.ToggleSelected = func() {
	// 	i, _ := table.Table.GetSelection()
	// 	taskName := table.Table.GetCell(i, 0).Text
	// 	isSelected := isTaskSelected(TUI.tasksSelected, taskName)
	// 	if isSelected {
	// 		TUI.tasksSelected = removeTask(TUI.tasksSelected, taskName)
	// 	} else {
	// 		task := getTask(TUI.tasks, taskName)
	// 		TUI.tasksSelected = append(TUI.tasksSelected, task)
	// 	}
	// 	TUI.emitter.Publish(Event{Name: "toggle_selected_task", Data: taskName})
	// 	table.updateCellStyles()
	// }
	// table.SelectAllRows = func() {
	// 	for i := 1; i < table.Table.GetRowCount(); i++ {
	// 		taskName := table.Table.GetCell(i, 0).Text
	// 		if !isTaskSelected(TUI.tasksSelected, taskName) {
	// 			task := getTask(TUI.tasks, taskName)
	// 			TUI.tasksSelected = append(TUI.tasksSelected, task)
	// 		}
	// 	}
	// 	TUI.emitter.Publish(Event{Name: "update_all_selected_tasks", Data: ""})
	// 	table.updateCellStyles()
	// }
	// table.DeSelectAllRows = func() {
	// 	for i := 1; i < table.Table.GetRowCount(); i++ {
	// 		taskName := table.Table.GetCell(i, 0).Text
	// 		TUI.tasksSelected = removeTask(TUI.tasksSelected, taskName)
	// 	}
	// 	TUI.emitter.Publish(Event{Name: "update_all_selected_tasks", Data: ""})
	// 	table.updateCellStyles()
	// }
	// table.DescribeRow = func() {
	// 	row, _ := table.Table.GetSelection()
	// 	if row > 0 {
	// 		showTaskDescModal(tasks[row-1])
	// 	}
	// }

	// // Events
	// TUI.emitter.Subscribe("filter_tasks", func(e Event) {
	// 	table.filterTasks()
	// })
	// TUI.emitter.Subscribe("remove_selected_task", func(e Event) {
	// 	table.updateTasksTable()
	// })
	// TUI.emitter.Subscribe("select_all_tasks", func(e Event) {
	// 	table.SelectAllRows()
	// })
	// TUI.emitter.Subscribe("deselect_all_tasks", func(e Event) {
	// 	table.DeSelectAllRows()
	// })

	// table.updateTasksTable()

	return table
}
