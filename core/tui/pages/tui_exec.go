package pages

import (
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

type TExecPage struct {
	focusable []*misc.TItem
}

func CreateExecPage(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	e := &TExecPage{}
	projectData := views.CreateProjectsData(
		projects,
		projectTags,
		projectPaths,
		[]string{"Project", "Description", "Tag"},
		2,
		true,
		true,
		true,
		true,
		true,
	)

	// Views
	streamView, ansiWriter := components.CreateOutputView("[2] Output")
	projectInfo := views.CreateRunInfoVIew()
	cmdInfo := views.CreateExecInfoView()
	cmdView := components.CreateTextArea("[1] Command")
	spec := views.CreateSpecView()

	// Pages
	execPage := e.createSelectPage(projectData, projectInfo, cmdView)
	outputPage := e.createOutputPage(cmdInfo, cmdView, streamView)
	pages := tview.NewPages().
		AddPage("exec-projects", execPage, true, true).
		AddPage("exec-run", outputPage, true, false)

	// Main page
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	// Focus
	e.focusable = e.updateSelectFocusable(*projectData, cmdView)
	misc.ExecLastFocus = &e.focusable[0].Primitive

	// Shortcuts
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			e.focusable = e.switchView(pages, projectData, cmdView, streamView)
			misc.App.SetFocus(e.focusable[0].Primitive)
			misc.ExecLastFocus = &e.focusable[0].Primitive
			return nil
		case tcell.KeyCtrlR:
			e.focusable = e.switchBeforeRun(pages, e.focusable, cmdView, streamView)
			misc.App.SetFocus(e.focusable[0].Primitive)
			misc.ExecLastFocus = &e.focusable[0].Primitive
			cmd := cmdView.GetText()
			e.runCmd(streamView, cmd, projectData.Projects, projectData.ProjectsSelected, spec, ansiWriter)
			return nil
		}

		switch event.Key() {
		case tcell.KeyTab:
			nextPrimitive := misc.FocusNext(e.focusable)
			misc.ExecLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(e.focusable)
			misc.ExecLastFocus = nextPrimitive
			return nil
		case tcell.KeyCtrlO:
			components.OpenModal("spec-modal", "Options", spec.View, 30, 11)
			return nil
		case tcell.KeyCtrlX:
			streamView.Clear()
			return nil
		case tcell.KeyRune:
			if _, ok := misc.App.GetFocus().(*tview.TextArea); ok {
				return event
			}

			name, _ := pages.GetFrontPage()
			if name == "exec-projects" {
				switch event.Rune() {
				case 'C': // Clear filters
					projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_path_filter", Data: ""})
					projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_path_selections", Data: ""})
					projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_project_filter", Data: ""})
					projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_project_selections", Data: ""})
					projectData.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
					return nil
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					misc.FocusPage(event, e.focusable)
					return nil
				}
			}

			if name == "exec-run" {
				switch event.Rune() {
				case '1':
					misc.App.SetFocus(cmdView)
					return nil
				case '2':
					misc.App.SetFocus(streamView)
					return nil
				}
			}
		}

		return event
	})

	return page
}

func (e *TExecPage) createSelectPage(
	projectData *views.TProject,
	infoPane *tview.TextView,
	execInput *tview.TextArea,
) *tview.Flex {
	isProjectTable := projectData.ProjectStyle == "project-table"
	projectPages := tview.NewPages().
		AddPage("project-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(projectData.ProjectTableView.Root, 0, 1, true), true, isProjectTable).
		AddPage("project-tree", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(projectData.ProjectTreeView.Root, 0, 8, false), true, !isProjectTable)
	projectPages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			if projectData.ProjectStyle == "project-table" {
				projectData.ProjectStyle = "project-tree"
			} else {
				projectData.ProjectStyle = "project-table"
			}
			projectPages.SwitchToPage(projectData.ProjectStyle)
			e.focusable = e.updateSelectFocusable(*projectData, execInput)
			misc.App.SetFocus(e.focusable[1].Primitive)
			misc.RunLastFocus = &e.focusable[1].Primitive
			return nil
		}
		return event
	})

	projectData.ContextView = tview.NewFlex().SetDirection(tview.FlexRow)
	if projectData.TagView.List.GetItemCount() > 0 {
		projectData.ContextView.AddItem(projectData.TagView.Root, 0, 1, false)
	}
	if projectData.PathView.List.GetItemCount() > 0 {
		projectData.ContextView.AddItem(projectData.PathView.Root, 0, 1, false)
	}

	bottom := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(projectPages, 0, 1, false).
		AddItem(projectData.ContextView, 30, 1, false)

	// Container
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(execInput, 8, 0, true).
		AddItem(bottom, 0, 1, false).
		AddItem(infoPane, 1, 0, false)

	return page
}

func (e *TExecPage) createOutputPage(
	infoPane *tview.TextView,
	execInput *tview.TextArea,
	streamView *tview.TextView,
) *tview.Flex {
	outputView := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(execInput, 8, 0, true).
		AddItem(streamView, 0, 1, false).
		AddItem(infoPane, 1, 0, false)

	return outputView
}

func (e *TExecPage) updateSelectFocusable(
	projectData views.TProject,
	execInput *tview.TextArea,
) []*misc.TItem {
	focusable := []*misc.TItem{
		misc.GetTUIItem(
			execInput,
			execInput.Box,
		),
	}

	// Project
	if projectData.ProjectStyle == "project-table" {
		focusable = append(
			focusable, misc.GetTUIItem(
				projectData.ProjectTableView.Table,
				projectData.ProjectTableView.Table.Box,
			))
	} else {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				projectData.ProjectTreeView.Tree,
				projectData.ProjectTreeView.Tree.Box,
			))
	}

	if len(projectData.ProjectTags) > 0 {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				projectData.TagView.List,
				projectData.TagView.List.Box,
			))
	}
	if len(projectData.ProjectPaths) > 0 {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				projectData.PathView.List,
				projectData.PathView.List.Box,
			))
	}

	return focusable
}

func (e *TExecPage) updateStreamFocusable(
	execInput *tview.TextArea,
	streamView *tview.TextView,
) []*misc.TItem {
	focusable := []*misc.TItem{
		misc.GetTUIItem(execInput, execInput.Box),
		misc.GetTUIItem(streamView, streamView.Box),
	}
	return focusable
}

func (e *TExecPage) switchView(
	pages *tview.Pages,
	data *views.TProject,
	cmdView *tview.TextArea,
	streamView *tview.TextView,
) []*misc.TItem {
	name, _ := pages.GetFrontPage()
	var focusable []*misc.TItem
	if name == "exec-run" {
		pages.SwitchToPage("exec-projects")
		focusable = e.updateSelectFocusable(*data, cmdView)
	} else {
		pages.SwitchToPage("exec-run")
		focusable = e.updateStreamFocusable(cmdView, streamView)
	}

	return focusable
}

func (e *TExecPage) switchBeforeRun(
	pages *tview.Pages,
	focusable []*misc.TItem,
	cmdView *tview.TextArea,
	streamView *tview.TextView,
) []*misc.TItem {
	name, _ := pages.GetFrontPage()
	if name == "exec-projects" {
		pages.SwitchToPage("exec-run")
		focusable = e.updateStreamFocusable(cmdView, streamView)
	}

	return focusable
}

func (e *TExecPage) runCmd(
	streamView *tview.TextView,
	cmd string,
	projects []dao.Project,
	projectsSelectMap map[string]bool,
	spec *views.TSpec,
	ansiWriter *misc.ThreadSafeWriter,
) {
	// Check if any projects selected
	selectedProjects := []dao.Project{}
	for _, project := range projects {
		if projectsSelectMap[project.Name] {
			selectedProjects = append(selectedProjects, project)
		}
	}
	if len(selectedProjects) < 1 {
		return
	}

	// Task
	task := dao.Task{Name: "", Cmd: cmd}
	taskErrors := make([]dao.ResourceErrors[dao.Task], 1)
	task.ParseTask(*misc.Config, &taskErrors[0])
	task.SpecData.Output = spec.Output
	task.SpecData.Parallel = spec.Parallel
	task.SpecData.IgnoreErrors = spec.IgnoreErrors
	task.SpecData.IgnoreNonExisting = spec.IgnoreNonExisting
	task.SpecData.OmitEmptyRows = spec.OmitEmptyRows
	task.SpecData.OmitEmptyColumns = spec.OmitEmptyColumns

	// Flags
	runFlags := core.RunFlags{
		Silent: true,

		// Target
		Cwd:      false,
		All:      false,
		TagsExpr: "",

		Target: "default",
		Spec:   "default",

		Output:            spec.Output,
		Parallel:          spec.Parallel,
		IgnoreErrors:      spec.IgnoreErrors,
		IgnoreNonExisting: spec.IgnoreNonExisting,
		OmitEmptyRows:     spec.OmitEmptyRows,
		OmitEmptyColumns:  spec.OmitEmptyColumns,
	}
	setRunFlags := core.SetRunFlags{
		Parallel:          spec.Parallel,
		All:               true,
		Cwd:               true,
		IgnoreErrors:      true,
		IgnoreNonExisting: true,
		OmitEmptyRows:     true,
		OmitEmptyColumns:  true,
	}

	// Preprocess
	var tasks []dao.Task
	for range selectedProjects {
		t := dao.Task{}
		err := copier.Copy(&t, &task)
		core.CheckIfError(err)
		tasks = append(tasks, t)
	}

	// Run
	target := exec.Exec{Projects: selectedProjects, Tasks: tasks, Config: *misc.Config}

	if spec.ClearBeforeRun {
		streamView.Clear()
	}

	if spec.Output == "table" {
		text := streamView.GetText(false)
		streamView.SetText(text + "\n")
	} else {
		text := streamView.GetText(false)
		streamView.SetText(text + "\n")
	}

	err := target.RunTUI([]string{}, &runFlags, &setRunFlags, spec.Output, ansiWriter, ansiWriter)
	core.CheckIfError(err)

	streamView.ScrollToEnd()
}
