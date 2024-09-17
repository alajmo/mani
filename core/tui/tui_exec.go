package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createExecPage() {
	TUI.runPage = tview.NewFlex()
	// // Data
	// TUI.tasksFiltered = TUI.tasks

	execInput := createExecInput()
	execTable := createExecTable()
	helpInfo := createProjectInfo()

	TUI.execPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(helpInfo, 1, 0, true).
				AddItem(execTable.Table, 0, 8, true).
				AddItem(execInput, 8, 0, true),
			0, 1, true).
		AddItem(TUI.search, 1, 0, false)

	focusableElements := []tview.Primitive{execTable.Table, execInput}

	currentFocus := 0
	// Handle global shortcuts
	TUI.execPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
			case 's': // Select projects
				showSelectProjectsModal()
				return nil
			case '1': // Table focus
				TUI.app.SetFocus(execTable.Table)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			case '2': // Tags focus
				TUI.app.SetFocus(execInput)
				currentFocus = getCurrentFocusIndex(focusableElements)
				return nil
			}
		}
		return event
	})
}

func createExecTable() TUITable {
	table := TUITable{}
	table.createTable()
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

	table.updateExecTable()

	return table
}

func (t *TUITable) updateExecTable() {
	t.Table.Clear()

	// Set up headers
	headers := []string{"Project", "Output"}
	for col, header := range headers {
		t.Table.SetCell(0, col, createTableHeader(header))
	}

	// Populate the table with task data
	// for row, task := range TUI.tasksFiltered {
	// 	t.Table.SetCell(row+1, 0, tview.NewTableCell(task.Name))
	// 	t.Table.SetCell(row+1, 1, tview.NewTableCell(task.Desc))
	// }

	t.updateCellStyles()
}

func createExecInput() *tview.InputField {
	textInput := tview.NewInputField()
	textInput.SetBorder(true)
	// textInput.SetWrap(false)
	textInput.SetTitle("Command")
	textInput.SetTitleAlign(tview.AlignLeft)
	textInput.SetFieldBackgroundColor(THEME.BG)
	textInput.SetFieldTextColor(THEME.FG)
	textInput.SetBorderPadding(0, 0, 1, 1)

	textInput.SetFocusFunc(func() {
		setActive(textInput, true)
	})
	textInput.SetBlurFunc(func() {
		setActive(textInput, false)
	})

	return textInput
}

func createProjectInfo() *tview.TextView {
	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
    SetText(fmt.Sprintf("[blue]<s>[white] the Select projects, [blue]<t>[white] Toggle output"))

		// l.List.SetItemText(i, "[blue::b]"+mainText, secondaryText)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)

	return helpInfo
}

func setActive(textInput *tview.InputField, active bool) {
	title := "Command"

	if active {
		textInput.Box.SetBorderColor(THEME.BORDER_COLOR_FOCUS)
		textInput.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR_FOCUS, title))
	} else {
		textInput.Box.SetBorderColor(THEME.BORDER_COLOR)
		textInput.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR, title))
	}
}

func showSelectProjectsModal() {
	description := "Hi"
	openModal("project-description-modal", description, "Select Projects", 90, 40)
}
