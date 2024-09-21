package pages

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"
)

func CreateExecPage(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	data := views.CreateProjectsData(projects, projectTags, projectPaths)

	helpInfo := createProjectInfo()
	projectsView := createSelectProjectsView(&data)
	execView := createRunProjectsView(&data)

	pages := tview.NewPages().
		AddPage("exec-projects", projectsView, true, true).
		AddPage("exec-run", execView, true, false)

	// Select projects
	execPage := tview.NewFlex()
	execPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(helpInfo, 1, 0, false).
		AddItem(pages, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

		// - Global:
	// s
	// help
	// search
	//
	// - Local:
	// Tab
	// d
	// s

	execPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 's': // Select projects
				name, _ := pages.GetFrontPage()
				if name == "exec-run" {
					pages.SwitchToPage("exec-projects")
				} else {
					pages.SwitchToPage("exec-run")
				}
				return nil
			}
		}
		return event
	})

	return execPage
}

func createExecTable() components.TUITable {
	table := components.TUITable{}
	table.CreateTable()
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

	updateExecTable(&table)

	return table
}

func updateExecTable(t *components.TUITable) {
	t.Table.Clear()

	// Set up headers
	headers := []string{"Project", "Output"}
	for col, header := range headers {
		t.Table.SetCell(0, col, components.CreateTableHeader(header))
	}

	// Populate the table with task data
	// for row, task := range TUI.tasksFiltered {
	// 	t.Table.SetCell(row+1, 0, tview.NewTableCell(task.Name))
	// 	t.Table.SetCell(row+1, 1, tview.NewTableCell(task.Desc))
	// }

	t.UpdateCellStyles()
}

func createExecInput() *tview.InputField {
	textInput := tview.NewInputField()
	textInput.SetBorder(true)
	// textInput.SetWrap(false)
	textInput.SetTitle("Command")
	textInput.SetTitleAlign(tview.AlignLeft)
	textInput.SetFieldBackgroundColor(misc.THEME.BG)
	textInput.SetFieldTextColor(misc.THEME.FG)
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
		SetText(fmt.Sprintf("[blue]<s>[white] Select projects, [blue]<t>[white] Toggle output"))
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)

	return helpInfo
}

func setActive(textInput *tview.InputField, active bool) {
	title := "Command"

	if active {
		textInput.Box.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		textInput.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, title))
	} else {
		textInput.Box.SetBorderColor(misc.THEME.BORDER_COLOR)
		textInput.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR, title))
	}
}

func createSelectProjectsView(data *views.TUIProjects) *tview.Flex {
	// Table
	projectsTable := views.CreateProjectsTable(data)

	// Projects context
	tagsList := views.CreateProjectsTagsList(data)
	pathsList := views.CreateProjectsPathsList(data)
	selectedList := views.CreateProjectsSelectedList(data)
	data.ProjectsContextPage = tview.NewFlex().SetDirection(tview.FlexRow)
	if tagsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(tagsList.List, 0, 1, true)
	}
	if pathsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(pathsList.List, 0, 1, true)
	}
	data.ProjectsContextPage.AddItem(selectedList.List, 0, 1, true)

	// Container
	page := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(projectsTable.Table, 0, 1, true).
		AddItem(data.ProjectsContextPage, 30, 1, false)

	// Focusable elements
	focusableElements := []tview.Primitive{projectsTable.Table}
	if len(data.ProjectTags) > 0 {
		focusableElements = append(focusableElements, tagsList.List)
	}
	if len(data.ProjectPaths) > 0 {
		focusableElements = append(focusableElements, pathsList.List)
	}
	focusableElements = append(focusableElements, selectedList.List)

	currentFocus := 0
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
			case '1':
				misc.App.SetFocus(projectsTable.Table)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '2':
				misc.App.SetFocus(tagsList.List)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '3':
				misc.App.SetFocus(pathsList.List)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '4':
				misc.App.SetFocus(selectedList.List)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			}
		}
		return event
	})

	return page
}

func createRunProjectsView(data *views.TUIProjects) *tview.Flex {
	execInput := createExecInput()
	execTable := createExecTable()

	// Run
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(execInput, 8, 0, true).
				AddItem(execTable.Table, 0, 8, true),
			0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := []tview.Primitive{execInput, execTable.Table}
	currentFocus := 0
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
				misc.App.SetFocus(execInput)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '2': // Tags focus
				misc.App.SetFocus(execTable.Table)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			}
		}
		return event
	})

	return page
}
