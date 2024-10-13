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
	spec := &views.TUISpec{
		Output:            "text",
		ClearBeforeRun:    true,
		Parallel:          false,
		IgnoreErrors:      false,
		IgnoreNonExisting: false,
	}
	data := views.CreateProjectsData(projects, projectTags, projectPaths, []string{"Project"}, false)
	tableView, streamView := createExecTable()

	projectInfo := createProjectInfo()
	cmdInfo := createExecInfo()
	cmdView := createExecInput()
	specView := views.CreateSpecView(data.Emitter, spec)

	// Pages
	projectsView := createSelectProjectsView(&data, projectInfo, cmdView)
	execView := createRunProjectsView(&data, cmdInfo, cmdView, streamView, tableView)
	pages := tview.NewPages().
		AddPage("exec-projects", projectsView, true, true).
		AddPage("exec-run", execView, true, false)

	// Select projects
	page := tview.NewFlex()
	page = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := updateSelectProject(data, cmdView)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			name, _ := pages.GetFrontPage()
			if name == "exec-run" {
				pages.SwitchToPage("exec-projects")
				focusableElements = updateSelectProject(data, cmdView)
			} else {
				pages.SwitchToPage("exec-run")

				if spec.Output == "text" {
					focusableElements = updateRunText(cmdView, streamView)
				} else {
					focusableElements = updateRunTable(cmdView, tableView)
				}
			}

			misc.App.SetFocus(focusableElements[0].Primitive)
			return nil
		case tcell.KeyCtrlR:
			name, _ := pages.GetFrontPage()
			if name == "exec-projects" {
				pages.SwitchToPage("exec-run")
				if spec.Output == "text" {
					focusableElements = updateRunText(cmdView, streamView)
				} else {
					focusableElements = updateRunTable(cmdView, tableView)
				}
			}

			misc.App.SetFocus(focusableElements[0].Primitive)

			cmd := cmdView.GetText()
			runTask(tableView, streamView, cmd, data.ProjectsSelected, spec)
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
		case tcell.KeyCtrlO:
			components.OpenModal("spec-modal", "Options", specView, 50, 10)
			return nil
		case tcell.KeyCtrlX:
			streamView.Clear()
			return nil
		case tcell.KeyRune:
			// If TextArea is in focus
			if _, ok := misc.App.GetFocus().(*tview.TextArea); ok {
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
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					i := int(event.Rune()-'0') - 1
					if i < len(focusableElements) {
						misc.App.SetFocus(focusableElements[i].Box)
					}
					return nil
				}
			}

			if name == "exec-run" {
				switch event.Rune() {
				case '1': // Unselect all all
					misc.App.SetFocus(cmdView)
					return nil
				case '2':
					misc.App.SetFocus(tableView.Grid)
					return nil
				}
			}
		}

		return event
	})

	return page
}

func createExecTable() (components.TUIGrid, *tview.TextView) {
	grid := components.TUIGrid{Border: true}
	grid.CreateGrid()
	data := dao.TableOutput{
		Headers: []string{"Project", "Output"},
		Rows:    []dao.Row{},
	}
	updateExecTable(&grid, data)

	streamView := components.CreateTextView("Output")

	return grid, streamView
}

func updateExecTable(g *components.TUIGrid, data dao.TableOutput) {
	// g.Grid.Clear()
	// g.Grid.SetGap(1, 1)
	// g.Grid.SetColumns(16, 0) // First column fixed size 16, second column expands

	// // Set up headers
	// headers := []string{"Project", "Output"}
	// for col, header := range headers {
	// 	cell := components.CreateGridHeader(header)
	// 	g.Grid.AddItem(cell, 0, col, 1, 1, 0, 0, false)
	// }

	// // Calculate row heights and populate the table
	// rowHeights := []int{1} // Start with header row height
	// for row, task := range data.Rows {
	// 	cell1 := tview.NewTextView().SetText(task.Columns[0]).SetWordWrap(false)
	// 	cell2 := tview.NewTextView().SetText(task.Columns[1]).SetWordWrap(false)

	// 	g.Grid.AddItem(cell1, row+1, 0, 1, 1, 0, 0, false)
	// 	g.Grid.AddItem(cell2, row+1, 1, 1, 1, 0, 0, false)

	// 	height1 := misc.CalculateTextHeight(task.Columns[0])
	// 	height2 := misc.CalculateTextHeight(task.Columns[1])
	// 	rowHeight := misc.Max(height1, height2)
	// 	rowHeights = append(rowHeights, rowHeight)
	// }

	// g.Grid.SetRows(rowHeights...)
}

func createExecInput() *tview.TextArea {
	textInput := tview.NewTextArea()
	textInput.SetBorder(true)
	textInput.SetWrap(false)
	textInput.SetTitle("Command")
	textInput.SetTitleAlign(tview.AlignCenter)
	textInput.SetBackgroundColor(misc.THEME.BG)
	textInput.SetTitleColor(misc.THEME.FG)
	textInput.SetBorderPadding(0, 0, 1, 1)

	textInput.SetFocusFunc(func() {
		misc.PreviousPage = textInput
		setActive(textInput, true)
	})

	textInput.SetBlurFunc(func() {
		setActive(textInput, false)
	})

	return textInput
}

func setActive(textInput *tview.TextArea, active bool) {
	title := "Command"

	if active {
		textInput.Box.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		textInput.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, title))
	} else {
		textInput.Box.SetBorderColor(misc.THEME.BORDER_COLOR)
		textInput.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR, title))
	}
}

func createSelectProjectsView(
	data *views.TUIProjects,
	info *tview.TextView,
	execInput *tview.TextArea,
) *tview.Flex {
	// Table
	projectsTable := views.CreateProjectsTable(data, true, "Projects")

	// Projects context
	tagsList := views.CreateProjectsTagsList(data)
	pathsList := views.CreateProjectsPathsList(data)

	data.ProjectsTable = projectsTable.Table
	data.ProjectsContextPage = tview.NewFlex().SetDirection(tview.FlexRow)
	if tagsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(tagsList.List, 0, 1, false)
	}
	if pathsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(pathsList.List, 0, 1, false)
	}

	bottom := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(projectsTable.Table, 0, 1, false).
		AddItem(data.ProjectsContextPage, 30, 1, false)

	// Container
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(info, 1, 0, false).
		AddItem(execInput, 8, 0, true).
		AddItem(bottom, 0, 1, false)

	return page
}

func updateSelectProject(
	data views.TUIProjects,
	execInput *tview.TextArea,
) []*misc.TUIItem {
	focusableElements := []*misc.TUIItem{
		misc.GetTUIItem("Command", execInput, execInput.Box),
		misc.GetTUIItem("Projects", data.ProjectsTable, data.ProjectsTable.Box),
	}

	if len(data.ProjectTags) > 0 {
		focusableElements = append(
			focusableElements,
			misc.GetTUIItem(
				fmt.Sprintf("Tags (%d)", len(data.ProjectTags)),
				data.ProjectsTagsPane,
				data.ProjectsTagsPane.Box,
			))
	}
	if len(data.ProjectPaths) > 0 {
		focusableElements = append(
			focusableElements,
			misc.GetTUIItem(
				fmt.Sprintf("Paths (%d)", len(data.ProjectPaths)),
				data.ProjectsPathsPane,
				data.ProjectsPathsPane.Box,
			))
	}

	return focusableElements
}

func updateRunText(
	execInput *tview.TextArea,
	streamView *tview.TextView,
) []*misc.TUIItem {
	focusableElements := []*misc.TUIItem{
		misc.GetTUIItem("Command", execInput, execInput.Box),
		misc.GetTUIItem("Output", streamView, streamView.Box),
	}
	return focusableElements
}

func updateRunTable(
	execInput *tview.TextArea,
	execTable components.TUIGrid,
) []*misc.TUIItem {
	focusableElements := []*misc.TUIItem{
		misc.GetTUIItem("Command", execInput, execInput.Box),
		misc.GetTUIItem("Output", execTable.Grid, execTable.Grid.Box),
	}
	return focusableElements
}

func createRunProjectsView(
	data *views.TUIProjects,
	info *tview.TextView,
	execInput *tview.TextArea,
	streamView *tview.TextView,
	execTable components.TUIGrid,
) *tview.Flex {
	pages := tview.NewPages().
		AddPage("exec-text", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(streamView, 0, 1, true), true, true).
		AddPage("exec-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(execTable.Grid, 0, 8, true), true, false)

	data.Emitter.Subscribe("toggle_output", func(e misc.Event) {
		pages.SwitchToPage(e.Data.(string))
	})

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(info, 1, 0, true).
		AddItem(execInput, 8, 0, false).
		AddItem(pages, 0, 1, false)

	return page
}

func runTask(
	table components.TUIGrid,
	streamView *tview.TextView,
	cmd string,
	projects []dao.Project,
	spec *views.TUISpec,
) {
	if len(projects) < 1 {
		return
	}
	// Task
	task := dao.Task{Name: "", Cmd: cmd}
	taskErrors := make([]dao.ResourceErrors[dao.Task], 1)
	task.ParseTask(*misc.Config, &taskErrors[0])
	task.SpecData.Output = spec.Output
	task.SpecData.Parallel = spec.Parallel
	task.SpecData.IgnoreErrors = spec.IgnoreErrors

	// Flags
	runFlags := core.RunFlags{Silent: true}
	var setRunFlags core.SetRunFlags

	// Preprocess
	var tasks []dao.Task
	for range projects {
		t := dao.Task{}
		err := copier.Copy(&t, &task)
		core.CheckIfError(err)
		tasks = append(tasks, t)
	}

	if spec.ClearBeforeRun {
		streamView.Clear()
	}
	// Run
	target := exec.Exec{Projects: projects, Tasks: tasks, Config: *misc.Config}
	ansiWriter := tview.ANSIWriter(streamView)
	data, err := target.RunTUI([]string{}, &runFlags, &setRunFlags, "text", ansiWriter, ansiWriter)
	core.CheckIfError(err)

	streamView.ScrollToEnd()

	// Update table
	updateExecTable(&table, data)
}
