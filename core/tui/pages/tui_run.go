package pages

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/jinzhu/copier"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"
)

func CreateRunPage(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	data := views.CreateProjectsData(projects, projectTags, projectPaths)
	execTable := createExecTable()

	helpInfo := createRunInfo()
	projectsView := createSelecRuntProjectsView(&data)
	execView := createRunRunProjectsView(execTable)

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

	focusableElements := updateRunProjectSelectProject(data)

	execPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			name, _ := pages.GetFrontPage()
			if name == "exec-run" {
				pages.SwitchToPage("exec-projects")
				focusableElements = updateRunProjectSelectProject(data)
			} else {
				pages.SwitchToPage("exec-run")
				focusableElements = updateRunProject(data, execTable)
			}

			misc.App.SetFocus(focusableElements[0])
			return nil
		case tcell.KeyCtrlR:
			name, _ := pages.GetFrontPage()
			if name == "exec-projects" {
				pages.SwitchToPage("exec-run")
				focusableElements = updateRunProject(data, execTable)
			}

			misc.App.SetFocus(focusableElements[0])

			// cmd := execInput.GetText()
			// runTasks(execTable, cmd, data.ProjectsSelected)
			return nil
		}

		switch event.Key() {
		case tcell.KeyTab:
			misc.FocusNext(focusableElements)
			return nil
		case tcell.KeyBacktab:
			misc.FocusPrevious(focusableElements)
			return nil
			// TODO: Capture if on input box, then disable
		case tcell.KeyRune:
			// If NewInputField
			if _, ok := misc.App.GetFocus().(*tview.InputField); ok {
				return event
			}

			name, _ := pages.GetFrontPage()
			if name == "exec-projects" {
				switch event.Rune() {
				case 'f': // Clear filters
					data.Emitter.PublishAndWait(misc.Event{Name: "clear_filters", Data: ""})
					data.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
					return nil
				case 'a': // Select all
					data.Emitter.Publish(misc.Event{Name: "select_all_projects", Data: ""})
					return nil
				case 'c': // Unselect all all
					data.Emitter.Publish(misc.Event{Name: "deselect_all_projects", Data: ""})
					return nil
				case '1':
					misc.App.SetFocus(data.ProjectsTable)
				case '2':
					return nil
					misc.App.SetFocus(data.ProjectsTagsPane)
					return nil
				case '3':
					misc.App.SetFocus(data.ProjectsPathsPane)
					return nil
				case '4':
					misc.App.SetFocus(data.ProjectsSelectedPane)
					return nil
				}
			}

			if name == "exec-run" {
				switch event.Rune() {
				case '1': // Unselect all all
					// misc.App.SetFocus()
					return nil
				case '2':
					misc.App.SetFocus(execTable.Grid)
					return nil
				}
			}
		}

		return event
	})

	return execPage
}

func createRunTable() components.TUIGrid {
	grid := components.TUIGrid{}
	grid.CreateGrid()

	data := dao.TableOutput{
		Headers: []string{"Project", "Output"},
		Rows:    []dao.Row{},
	}

	updateExecTable(&grid, data)

	return grid
}

func updateRunProjectTable(g *components.TUIGrid, data dao.TableOutput) {
	g.Grid.Clear()
	g.Grid.SetGap(1, 1)
	g.Grid.SetColumns(16, 0) // First column fixed size 16, second column expands

	// Set up headers
	headers := []string{"Project", "Output"}
	for col, header := range headers {
		cell := components.CreateGridHeader(header)
		g.Grid.AddItem(cell, 0, col, 1, 1, 0, 0, false)
	}

	// Calculate row heights and populate the table
	rowHeights := []int{1} // Start with header row height
	for row, task := range data.Rows {
		cell1 := tview.NewTextView().SetText(task.Columns[0]).SetWordWrap(false)
		cell2 := tview.NewTextView().SetText(task.Columns[1]).SetWordWrap(false)

		g.Grid.AddItem(cell1, row+1, 0, 1, 1, 0, 0, false)
		g.Grid.AddItem(cell2, row+1, 1, 1, 1, 0, 0, false)

		height1 := misc.CalculateTextHeight(task.Columns[0])
		height2 := misc.CalculateTextHeight(task.Columns[1])
		rowHeight := misc.Max(height1, height2)
		rowHeights = append(rowHeights, rowHeight)
	}

	g.Grid.SetRows(rowHeights...)
}

func createRunInfo() *tview.TextView {
	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[green]<Ctrl-r>[white] Run, [blue]<Ctrl-s>[white] Switch view"))
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func createSelecRuntProjectsView(data *views.TUIProjects) *tview.Flex {
	// Tasks
	// tasksTable := views.CreateProjectsTable(data, true)
	// tasksSelectedList := views.CreateProjectsSelectedList(data)

	// Table
	projectsTable := views.CreateProjectsTable(data, true)
	tagsList := views.CreateProjectsTagsList(data)
	pathsList := views.CreateProjectsPathsList(data)
	selectedList := views.CreateProjectsSelectedList(data)

	data.ProjectsTable = projectsTable.Table
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
		// AddItem(tasksTable.Table, 0, 1, true).
		AddItem(projectsTable.Table, 0, 1, true).
		AddItem(data.ProjectsContextPage, 30, 1, false)

	return page
}

func updateRunProjectSelectProject(data views.TUIProjects) []tview.Primitive {
	focusableElements := []tview.Primitive{data.ProjectsTable}

	if len(data.ProjectTags) > 0 {
		focusableElements = append(focusableElements, data.ProjectsTagsPane)
	}
	if len(data.ProjectPaths) > 0 {
		focusableElements = append(focusableElements, data.ProjectsPathsPane)
	}
	focusableElements = append(focusableElements, data.ProjectsSelectedPane)

	return focusableElements
}

func updateRunProject(
	data views.TUIProjects,
	execTable components.TUIGrid,
) []tview.Primitive {
	focusableElements := []tview.Primitive{execTable.Grid}
	return focusableElements
}

func createRunRunProjectsView(execTable components.TUIGrid) *tview.Flex {
	// Run
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(execTable.Grid, 0, 8, true),
			0, 1, true)

	return page
}

func runTasks(table components.TUIGrid, cmd string, projects []dao.Project) {
	task := dao.Task{Name: "output", Cmd: cmd}
	taskErrors := make([]dao.ResourceErrors[dao.Task], 1)
	task.ParseTask(*misc.Config, &taskErrors[0])

	task.SpecData.Output = "table"

	var tasks []dao.Task
	for range projects {
		t := dao.Task{}
		err := copier.Copy(&t, &task)
		core.CheckIfError(err)
		tasks = append(tasks, t)
	}

	var runFlags core.RunFlags
	runFlags.Silent = true
	var setRunFlags core.SetRunFlags

	target := exec.Exec{Projects: projects, Tasks: tasks, Config: *misc.Config}
	data, err := target.RunTUI([]string{}, &runFlags, &setRunFlags)
	core.CheckIfError(err)

	updateExecTable(&table, data)
}
