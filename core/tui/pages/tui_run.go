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
	spec := &views.TUISpec{
		Output:            "text",
		ClearBeforeRun:    true,
		Parallel:          false,
		IgnoreErrors:      false,
		IgnoreNonExisting: false,
	}
	taskData := views.CreateTasksData(tasks, []string{"Name"}, false)
	projectData := views.CreateProjectsData(projects, projectTags, projectPaths, []string{"Project"}, false)
	tableView, streamView := createExecTable()

	projectInfo := createProjectInfo()
	cmdInfo := createExecInfo()
	specView := views.CreateSpecView(projectData.Emitter, spec)

	taskProjectsView := createMainView(&taskData, &projectData, projectInfo)
	execView, execPages := createRunRunProjectsView(&projectData, cmdInfo, streamView, tableView)

	pages := tview.NewPages().
		AddPage("exec-projects", taskProjectsView, true, true).
		AddPage("exec-run", execView, true, false)

	// Select projects
	execPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := updateRunProjectSelectProject(taskData, projectData)
	misc.RunLastFocus = &focusableElements[0].Primitive

	execPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			name, _ := pages.GetFrontPage()
			if name == "exec-run" {
				pages.SwitchToPage("exec-projects")
				focusableElements = updateRunProjectSelectProject(taskData, projectData)
			} else {
				pages.SwitchToPage("exec-run")

				if spec.Output == "text" {
					focusableElements = updateTaskRunText(streamView)
				} else {
					focusableElements = updateTaskRunTable(tableView)
				}
			}

			misc.App.SetFocus(focusableElements[0].Primitive)
			misc.RunLastFocus = &focusableElements[0].Primitive
			return nil
		case tcell.KeyCtrlR:
			name, _ := pages.GetFrontPage()
			if name == "exec-projects" {
				pages.SwitchToPage("exec-run")

				if spec.Output == "text" {
					focusableElements = updateTaskRunText(streamView)
				} else {
					focusableElements = updateTaskRunTable(tableView)
				}
			}

			misc.App.SetFocus(focusableElements[0].Primitive)
			misc.RunLastFocus = &focusableElements[0].Primitive

			runTasks(tableView, streamView, taskData.TasksSelected, projectData.ProjectsSelected, spec)
			return nil
		}
		switch event.Key() {
		case tcell.KeyTab:
			nextPrimitive := misc.FocusNext(focusableElements)
			misc.RunLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(focusableElements)
			misc.RunLastFocus = nextPrimitive
			return nil
		case tcell.KeyCtrlO:
			components.OpenModal("spec-modal", " Options ", specView, 50, 10)
			return nil
		case tcell.KeyRune:
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
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					i := int(event.Rune()-'0') - 1
					if i < len(focusableElements) {
						misc.App.SetFocus(focusableElements[i].Box)
					}
					return nil
				}
			}
		}

		return event
	})

	projectData.Emitter.Subscribe("toggle_output", func(e misc.Event) {
		page := e.Data.(string)
		execPages.SwitchToPage(page)

		currentPage, _ := pages.GetFrontPage()
		if currentPage == "exec-run" {
			if page == "exec-text" {
				focusableElements = updateTaskRunText(streamView)
			} else {
				focusableElements = updateTaskRunTable(tableView)
			}
			misc.ExecLastFocus = &focusableElements[0].Primitive
			misc.PreviousPage = focusableElements[0].Primitive
		}
	})

	return execPage
}

func createProjectInfo() *tview.TextView {
	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[green]<Ctrl-r>[white] Run command, [blue]<Ctrl-o>[white] Options, [blue]<Ctrl-s>[white] Switch view"))
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func createExecInfo() *tview.TextView {
	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[green]<Ctrl-r>[white] Run command, [green]<Ctrl-x>[white] Clear, [blue]<Ctrl-o>[white] Options, [blue]<Ctrl-s>[white] Switch view"))
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func createMainView(
	tasksData *views.TUITasks,
	projectData *views.TUIProjects,
	info *tview.TextView,
) *tview.Flex {
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

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(info, 1, 0, false).
		AddItem(page, 0, 1, true)

	return root
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
		focusableElements = append(
			focusableElements,
			misc.GetTUIItem(
				fmt.Sprintf("Tags (%d)", len(projectsData.ProjectTags)),
				projectsData.ProjectsTagsPane,
				projectsData.ProjectsTagsPane.Box),
		)
	}
	if len(projectsData.ProjectPaths) > 0 {
		focusableElements = append(
			focusableElements,
			misc.GetTUIItem(
				fmt.Sprintf("Paths (%d)", len(projectsData.ProjectPaths)),
				projectsData.ProjectsPathsPane,
				projectsData.ProjectsPathsPane.Box),
		)
	}

	return focusableElements
}

func updateTaskRunText(
	streamView *tview.TextView,
) []*misc.TUIItem {
	focusableElements := []*misc.TUIItem{
		misc.GetTUIItem("Output", streamView, streamView.Box),
	}
	return focusableElements
}

func updateTaskRunTable(
	execTable components.TUIGrid,
) []*misc.TUIItem {
	focusableElements := []*misc.TUIItem{
		misc.GetTUIItem("Output", execTable.Rows, execTable.Rows.Box),
	}
	return focusableElements
}

func createRunRunProjectsView(
	data *views.TUIProjects,
	info *tview.TextView,
	streamView *tview.TextView,
	execTable components.TUIGrid,
) (*tview.Flex, *tview.Pages) {
	pages := tview.NewPages().
		AddPage("exec-text", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(streamView, 0, 1, true), true, true).
		AddPage("exec-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(execTable.Grid, 0, 8, true), true, false)

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(info, 1, 0, true).
		AddItem(pages, 0, 1, false)

	return page, pages
}

func createRunTable() (components.TUIGrid, *tview.TextView) {
	grid := components.TUIGrid{Border: true}
	grid.CreateGrid()

	data := dao.TableOutput{
		Headers: []string{},
		Rows:    []dao.Row{},
	}

	updateRunProjectTable(&grid, data)
	streamView := components.CreateTextView("Output")

	return grid, streamView
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

func runTasks(
	table components.TUIGrid,
	streamView *tview.TextView,
	tasks []dao.Task,
	projects []dao.Project,
	spec *views.TUISpec,
) {
	if len(projects) < 1 {
		return
	}

	if spec.ClearBeforeRun {
		streamView.Clear()
	}

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
		Silent:       true,
		Projects:     projectNames,
		Output:       spec.Output,
		Parallel:     spec.Parallel,
		IgnoreErrors: spec.IgnoreErrors,
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
	ansiWriter := tview.ANSIWriter(streamView)
	data, err := target.RunTUI([]string{}, &runFlags, &setRunFlags, spec.Output, ansiWriter, ansiWriter)
	core.CheckIfError(err)

	streamView.ScrollToEnd()

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
