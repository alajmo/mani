package pages

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"
)

func CreateRunPage(
	tasks []dao.Task,
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	taskData := views.CreateTasksData(tasks, []string{"Name"}, false)
	projectData := views.CreateProjectsData(projects, projectTags, projectPaths, []string{"Project"}, false)
	// runTable := createRunTable()
	runTable := testTable()

	helpInfo := createRunInfo()
	mainView := createMainView(&taskData, &projectData)
	runView := createRunRunProjectsView(runTable)

	pages := tview.NewPages().
		AddPage("exec-projects", mainView, true, true).
		AddPage("exec-run", runView, true, false)

		// Select projects
	execPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(helpInfo, 1, 0, false).
		AddItem(pages, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := updateRunProjectSelectProject(taskData, projectData)

	execPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			name, _ := pages.GetFrontPage()
			if name == "exec-run" {
				pages.SwitchToPage("exec-projects")
				focusableElements = updateRunProjectSelectProject(taskData, projectData)
			} else {
				pages.SwitchToPage("exec-run")
				focusableElements = updateRunProject(projectData, *runTable)
			}

			misc.App.SetFocus(focusableElements[0].Primitive)
			return nil
		case tcell.KeyCtrlR:
			name, _ := pages.GetFrontPage()
			if name == "exec-projects" {
				pages.SwitchToPage("exec-run")
				focusableElements = updateRunProject(projectData, *runTable)
			}

			misc.App.SetFocus(focusableElements[0].Primitive)

			runTasks(*runTable, taskData.TasksSelected, projectData.ProjectsSelected)
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
					projectData.Emitter.PublishAndWait(misc.Event{Name: "clear_filters", Data: ""})
					projectData.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
					return nil
				case 'a': // Select all
					projectData.Emitter.Publish(misc.Event{Name: "select_all_projects", Data: ""})
					return nil
				case 'c': // Unselect all all
					projectData.Emitter.Publish(misc.Event{Name: "deselect_all_projects", Data: ""})
					return nil
				case '1':
					misc.App.SetFocus(taskData.TasksTable)
					return nil
				case '2':
					misc.App.SetFocus(projectData.ProjectsTable)
					return nil
				case '3':
					misc.App.SetFocus(projectData.ProjectsTagsPane)
					return nil
				case '4':
					misc.App.SetFocus(projectData.ProjectsPathsPane)
					return nil
				}
			}

			if name == "exec-run" {
				switch event.Rune() {
				case '1': // Unselect all all
					// misc.App.SetFocus()
					return nil
				case '2':
					misc.App.SetFocus(runTable.Grid)
					return nil
				}
			}
		}

		return event
	})

	return execPage
}

func createRunInfo() *tview.TextView {
	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[green]<Ctrl-r>[white] Run tasks, [blue]<Ctrl-s>[white] Switch view"))
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func createMainView(tasksData *views.TUITasks, projectData *views.TUIProjects) *tview.Flex {
	// Tasks
	tasksTable := views.CreateTasksTable(tasksData, true, "Tasks")
	tasksData.TasksTable = tasksTable.Table

	// Project
	projectsTable := views.CreateProjectsTable(projectData, true, "Projects")
	tagsList := views.CreateProjectsTagsList(projectData)
	pathsList := views.CreateProjectsPathsList(projectData)

	projectData.ProjectsTable = projectsTable.Table
	projectData.ProjectsContextPage = tview.NewFlex().SetDirection(tview.FlexRow)
	if tagsList.List.GetItemCount() > 0 {
		projectData.ProjectsContextPage.AddItem(tagsList.List, 0, 1, true)
	}
	if pathsList.List.GetItemCount() > 0 {
		projectData.ProjectsContextPage.AddItem(pathsList.List, 0, 1, true)
	}
	projects := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(projectsTable.Table, 0, 2, true).
		AddItem(projectData.ProjectsContextPage, 0, 1, false)

	page := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tasksTable.Table, 0, 1, true).
		AddItem(projects, 0, 2, false)

	return page
}

func updateRunProjectSelectProject(
	tasksData views.TUITasks,
	projectsData views.TUIProjects,
) []*misc.TUIItem {
	focusableElements := []*misc.TUIItem{
		misc.GetTUIItem("Tasks", tasksData.TasksTable, tasksData.TasksTable.Box),
		misc.GetTUIItem("Projects", projectsData.ProjectsTable, projectsData.ProjectsTable.Box),
	}

	if len(projectsData.ProjectTags) > 0 {
		focusableElements = append(focusableElements, misc.GetTUIItem("Tags", projectsData.ProjectsTagsPane, projectsData.ProjectsTagsPane.Box))
	}
	if len(projectsData.ProjectPaths) > 0 {
		focusableElements = append(focusableElements, misc.GetTUIItem("Paths", projectsData.ProjectsPathsPane, projectsData.ProjectsPathsPane.Box))
	}

	return focusableElements
}

func updateRunProject(
	data views.TUIProjects,
	execTable components.TUIGrid,
) []*misc.TUIItem {
	focusableElements := []*misc.TUIItem{misc.GetTUIItem("Output", execTable.Grid, execTable.Grid.Box)}
	return focusableElements
}

func createRunRunProjectsView(execTable *components.TUIGrid) *tview.Flex {
	// Run
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(execTable.Grid, 0, 0, true)

	// page := tview.NewFlex().
	// 	SetDirection(tview.FlexRow).
	// 	AddItem(
	// 		tview.NewFlex().SetDirection(tview.FlexRow).
	// 			AddItem(execTable.Grid, 0, 8, true),
	// 		0, 1, true)

	return page
}

func createRunTable() components.TUIGrid {
	grid := components.TUIGrid{Border: true}
	grid.CreateGrid()

	data := dao.TableOutput{
		Headers: []string{},
		Rows:    []dao.Row{},
	}

	updateRunProjectTable(&grid, data)

	return grid
}

func updateRunProjectTable(g *components.TUIGrid, data dao.TableOutput) {
	// g.Grid.Clear()
	// g.Grid.SetGap(1, 1)
	// g.Grid.SetColumns(16, 0) // First column fixed size 16, second column expands

	// // Set up headers
	// for col, header := range data.Headers {
	// 	cell := components.CreateGridHeader(header)
	// 	g.Grid.AddItem(cell, 0, col, 1, 1, 0, 0, false)
	// }

	// // Calculate row heights and populate the table
	// // rowHeights := []int{1} // Start with header row height
	// for row, task := range data.Rows {
	// 	for col, _ := range data.Headers {
	// 		cell := tview.NewTextView().SetText(task.Columns[col]).SetWordWrap(false)
	// 		g.Grid.AddItem(cell, row+1, col, 1, 1, 0, 0, false)
	// 		// height := misc.CalculateTextHeight(task.Columns[col])
	// 		// rowHeight := misc.Max(height, height)
	// 		// rowHeights = append(rowHeights, rowHeight)

	// 		// cell1 := tview.NewTextView().SetText(task.Columns[0]).SetWordWrap(false)
	// 		// cell2 := tview.NewTextView().SetText(task.Columns[1]).SetWordWrap(false)

	// 		// g.Grid.AddItem(cell1, row+1, 0, 1, 1, 0, 0, false)
	// 		// g.Grid.AddItem(cell2, row+1, 1, 1, 1, 0, 0, false)

	// 		// height1 := misc.CalculateTextHeight(task.Columns[0])
	// 		// height2 := misc.CalculateTextHeight(task.Columns[1])
	// 		// rowHeight := misc.Max(height1, height2)
	// 		// rowHeights = append(rowHeights, rowHeight)
	// 	}
	// }

	// // g.Grid.SetRows(rowHeights...)
}

func runTasks(table components.TUIGrid, tasks []dao.Task, projects []dao.Project) {
	// Preprocess
	var taskNames []string
	for _, task := range tasks {
		taskNames = append(taskNames, task.Name)
	}
	var projectNames []string
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	// Flags
	runFlags := core.RunFlags{
		Silent:   true,
		Projects: projectNames,
	}
	var setRunFlags core.SetRunFlags

	// Run
	var err error
	if len(taskNames) == 1 {
		tasks, projects, err = dao.ParseSingleTask(taskNames[0], &runFlags, &setRunFlags, misc.Config, []string{})
	} else {
		tasks, projects, err = dao.ParseManyTasks(taskNames, &runFlags, &setRunFlags, misc.Config, []string{})
	}
	core.CheckIfError(err)

	target := exec.Exec{Projects: projects, Tasks: tasks, Config: *misc.Config}
	core.CheckIfError(err)
	data, runErr := target.RunTUI([]string{}, &runFlags, &setRunFlags, "table", nil, nil)
	core.CheckIfError(runErr)

	// Update table
	updateRunProjectTable(&table, data)
}

func testTable() *components.TUIGrid {
	// Headers
	grid := &components.TUIGrid{Border: true}
	grid.CreateGrid()
	grid.Update()

	headersData := []string{"Project", "Output 1", "Output 2", "Output 3"}
	// Set up headers
	for col, header := range headersData {
		cell := components.CreateGridHeader(header)
		grid.Headers.AddItem(cell, 0, col, 1, 1, 0, 0, false)
	}

	// Rows
	data := dao.TableOutput{
		Rows: []dao.Row{
			dao.Row{Columns: []string{"hello1", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello2", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello3", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello4", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello5", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello6", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello7\nffffffffffffffffffffff\nkkkkkkkkkkkkkkkkkk\n11111111111111111111", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello8", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello9", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello10", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello11", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello12", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello13", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello14", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello15", "world", "foo", "bar"}},
			dao.Row{Columns: []string{"hello16", "world", "foo", "bar"}},
		},
	}

	rowHeights := []int{}
	// Set up data rows
	for row, task := range data.Rows {
		for col := range headersData {
			cell := tview.NewTextView().
				SetText(task.Columns[col]).
				SetWordWrap(true).
				SetTextAlign(tview.AlignLeft)
			grid.Rows.AddItem(cell, row, col, 1, 1, 0, 0, false)
			rowHeights = append(rowHeights, 2)
		}
	}
	grid.Rows.SetRows(rowHeights...)

	return grid
}
