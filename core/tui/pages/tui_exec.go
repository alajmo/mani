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

func CreateExecPage(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	data := views.CreateProjectsData(projects, projectTags, projectPaths)
	execTable := createExecTable()

	helpInfo := createProjectInfo()
	execInput := createExecInput()
	projectsView := createSelectProjectsView(&data, execInput)
	execView := createRunProjectsView(execTable, execInput)

	pages := tview.NewPages().
		AddPage("exec-projects", projectsView, true, true).
		AddPage("exec-run", execView, true, false)

	// Select projects
	execPage := tview.NewFlex()
	execPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(helpInfo, 1, 0, false).
		AddItem(execInput, 8, 0, true).
		AddItem(pages, 0, 1, false).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := updateSelectProject(data, execInput)

	currentFocus := 0
	execPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			currentFocus = 0

			name, _ := pages.GetFrontPage()
			if name == "exec-run" {
				pages.SwitchToPage("exec-projects")
				focusableElements = updateSelectProject(data, execInput)
			} else {
				pages.SwitchToPage("exec-run")
				focusableElements = updateRun(data, execTable, execInput)
			}

			misc.App.SetFocus(focusableElements[currentFocus])
			return nil
		case tcell.KeyCtrlR:
			name, _ := pages.GetFrontPage()
			if name == "exec-projects" {
				pages.SwitchToPage("exec-run")
				focusableElements = updateRun(data, execTable, execInput)
			}

			currentFocus = 0
			misc.App.SetFocus(focusableElements[currentFocus])

			cmd := execInput.GetText()
			runTask(execTable, cmd, data.ProjectsSelected)
			return nil
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
				case '1': // Unselect all all
					misc.App.SetFocus(execInput)
					currentFocus = misc.GetCurrentFocusIndex(focusableElements)
					return nil
				case '2':
					misc.App.SetFocus(data.ProjectsTable)
					currentFocus = misc.GetCurrentFocusIndex(focusableElements)
					return nil
				case '3':
					misc.App.SetFocus(data.ProjectsTagsPane)
					currentFocus = misc.GetCurrentFocusIndex(focusableElements)
					return nil
				case '4':
					misc.App.SetFocus(data.ProjectsPathsPane)
					currentFocus = misc.GetCurrentFocusIndex(focusableElements)
					return nil
				case '5':
					misc.App.SetFocus(data.ProjectsSelectedPane)
					currentFocus = misc.GetCurrentFocusIndex(focusableElements)
					return nil
				}
			}

			if name == "exec-run" {
				switch event.Rune() {
				case '1': // Unselect all all
					misc.App.SetFocus(execInput)
					currentFocus = misc.GetCurrentFocusIndex(focusableElements)
					return nil
				case '2':
					misc.App.SetFocus(execTable.Grid)
					currentFocus = misc.GetCurrentFocusIndex(focusableElements)
					return nil
				}
			}
		}

		return event
	})

	return execPage
}

func createExecTable() components.TUIGrid {
	grid := components.TUIGrid{}
	grid.CreateGrid()

	data := dao.TableOutput{
		Headers: []string{"Project", "Output"},
		Rows:    []dao.Row{},
	}

	updateExecTable(&grid, data)

	return grid
}

func updateExecTable(g *components.TUIGrid, data dao.TableOutput) {
	g.Grid.Clear()
	// g.Grid.SetGap(1, 0)
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

func createProjectInfo() *tview.TextView {
	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[green]<Ctrl-r>[white] Run, [blue]<Ctrl-s>[white] Reset, [blue]<t>[white] Toggle output"))
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
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

func createSelectProjectsView(data *views.TUIProjects, execInput *tview.InputField) *tview.Flex {
	// Table
	projectsTable := views.CreateProjectsTable(data, true)

	// Projects context
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
		AddItem(projectsTable.Table, 0, 1, true).
		AddItem(data.ProjectsContextPage, 30, 1, false)

	return page
}

func updateSelectProject(
	data views.TUIProjects,
	execInput *tview.InputField,
) []tview.Primitive {
	focusableElements := []tview.Primitive{execInput, data.ProjectsTable}

	if len(data.ProjectTags) > 0 {
		focusableElements = append(focusableElements, data.ProjectsTagsPane)
	}
	if len(data.ProjectPaths) > 0 {
		focusableElements = append(focusableElements, data.ProjectsPathsPane)
	}
	focusableElements = append(focusableElements, data.ProjectsSelectedPane)

	return focusableElements
}

func updateRun(
	data views.TUIProjects,
	execTable components.TUIGrid,
	execInput *tview.InputField,
) []tview.Primitive {
	focusableElements := []tview.Primitive{execInput, execTable.Grid}
	return focusableElements
}

func createRunProjectsView(execTable components.TUIGrid, execInput *tview.InputField) *tview.Flex {
	// Run
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(execTable.Grid, 0, 8, true),
			0, 1, true)

	return page
}

func runTask(table components.TUIGrid, cmd string, projects []dao.Project) {
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
