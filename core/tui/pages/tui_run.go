package pages

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/exec"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"
)

type TRunPage struct {
	focusable []*misc.TItem
}

func CreateRunPage(
	tasks []dao.Task,
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	r := &TRunPage{}

	// Data
	taskData := views.CreateTasksData(
		tasks,
		[]string{"Name", "Description"},
		1,
		true,
		true,
		true,
	)
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
	streamView, ansiWriter := components.CreateOutputView("[1] Output")
	runInfoView := views.CreateRunInfoVIew()
	execInfoView := views.CreateExecInfoView()
	spec := views.CreateSpecView()

	// Pages
	runPage := r.createSelectPage(taskData, projectData, runInfoView)
	outputPage := r.createOutputPage(execInfoView, streamView)
	pages := tview.NewPages().
		AddPage("exec-projects", runPage, true, true).
		AddPage("exec-run", outputPage, true, false)

	// Main page
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	// Focus
	r.focusable = r.updateRunFocusable(*taskData, *projectData)
	misc.RunLastFocus = &r.focusable[0].Primitive

	// Shortcuts
	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlS:
			r.focusable = r.switchView(pages, taskData, projectData, streamView)
			misc.App.SetFocus(r.focusable[0].Primitive)
			misc.RunLastFocus = &r.focusable[0].Primitive
			return nil
		case tcell.KeyCtrlR:
			r.focusable = r.switchBeforeRun(pages, r.focusable, streamView)
			misc.App.SetFocus(r.focusable[0].Primitive)
			misc.RunLastFocus = &r.focusable[0].Primitive
			r.runTasks(streamView, *taskData, *projectData, spec, ansiWriter)
			return nil
		}
		switch event.Key() {
		case tcell.KeyTab:
			nextPrimitive := misc.FocusNext(r.focusable)
			misc.RunLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(r.focusable)
			misc.RunLastFocus = nextPrimitive
			return nil
		case tcell.KeyCtrlO:
			components.OpenModal("spec-modal", "Options", spec.View, 30, 11)
			return nil
		case tcell.KeyCtrlX:
			streamView.Clear()
			return nil
		case tcell.KeyRune:
			if _, ok := misc.App.GetFocus().(*tview.InputField); ok {
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

					taskData.Emitter.PublishAndWait(misc.Event{Name: "remove_task_filter", Data: ""})
					taskData.Emitter.PublishAndWait(misc.Event{Name: "remove_task_selections", Data: ""})
					taskData.Emitter.Publish(misc.Event{Name: "filter_tasks", Data: ""})
					return nil
				case '1', '2', '3', '4', '5', '6', '7', '8', '9':
					misc.FocusPage(event, r.focusable)
					return nil
				}
			}
		}

		return event
	})

	return page
}

func (r *TRunPage) createSelectPage(
	taskData *views.TTask,
	projectData *views.TProject,
	info *tview.TextView,
) *tview.Flex {
	// Tasks
	isTaskTable := taskData.TaskStyle == "task-table"
	taskPages := tview.NewPages().
		AddPage(
			"task-table",
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(taskData.TaskTableView.Root, 0, 1, true),
			true, isTaskTable,
		).
		AddPage(
			"task-tree",
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(taskData.TaskTreeView.Root, 0, 8, false),
			true, !isTaskTable,
		)
	taskPages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlE:
			if taskData.TaskStyle == "task-table" {
				taskData.TaskStyle = "task-tree"
			} else {
				taskData.TaskStyle = "task-table"
			}
			taskPages.SwitchToPage(taskData.TaskStyle)
			r.focusable = r.updateRunFocusable(*taskData, *projectData)
			misc.App.SetFocus(r.focusable[0].Primitive)
			misc.RunLastFocus = &r.focusable[0].Primitive
			return nil
		}
		return event
	})

	// Projects
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
			r.focusable = r.updateRunFocusable(*taskData, *projectData)
			misc.App.SetFocus(r.focusable[1].Primitive)
			misc.RunLastFocus = &r.focusable[1].Primitive
			return nil
		}
		return event
	})

	projectData.ContextView = tview.NewFlex().SetDirection(tview.FlexRow)
	if projectData.TagView.List.GetItemCount() > 0 {
		projectData.ContextView.AddItem(projectData.TagView.Root, 0, 1, true)
	}
	if projectData.PathView.List.GetItemCount() > 0 {
		projectData.ContextView.AddItem(projectData.PathView.Root, 0, 1, true)
	}
	taskProjects := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(projectPages, 0, 1, true).
		AddItem(projectData.ContextView, 30, 1, false)

	page := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(taskPages, 0, 1, true).
		AddItem(taskProjects, 0, 1, false)

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(page, 0, 1, true).
		AddItem(info, 1, 0, false)
}

func (r *TRunPage) createOutputPage(
	info *tview.TextView,
	streamView *tview.TextView,
) *tview.Flex {
	outputView := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(streamView, 0, 1, false).
		AddItem(info, 1, 0, true)

	return outputView
}

func (r *TRunPage) updateRunFocusable(
	taskData views.TTask,
	projectData views.TProject,
) []*misc.TItem {
	focusable := []*misc.TItem{}

	// Task
	if taskData.TaskStyle == "task-table" {
		focusable = append(
			focusable, misc.GetTUIItem(
				taskData.TaskTableView.Table,
				taskData.TaskTableView.Table.Box,
			))
	} else {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				taskData.TaskTreeView.Tree,
				taskData.TaskTreeView.Tree.Box,
			))
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

	// Project Context
	if len(projectData.ProjectTags) > 0 {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				projectData.TagView.List,
				projectData.TagView.List.Box),
		)
	}
	if len(projectData.ProjectPaths) > 0 {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				projectData.PathView.List,
				projectData.PathView.List.Box),
		)
	}

	return focusable
}

func (r *TRunPage) updateStreamFocusable(streamView *tview.TextView) []*misc.TItem {
	focusable := []*misc.TItem{
		misc.GetTUIItem(streamView, streamView.Box),
	}
	return focusable
}

func (r *TRunPage) switchView(
	pages *tview.Pages,
	taskData *views.TTask,
	projectData *views.TProject,
	streamView *tview.TextView,
) []*misc.TItem {
	name, _ := pages.GetFrontPage()
	var focusable []*misc.TItem
	if name == "exec-run" {
		pages.SwitchToPage("exec-projects")
		focusable = r.updateRunFocusable(*taskData, *projectData)
	} else {
		pages.SwitchToPage("exec-run")
		focusable = r.updateStreamFocusable(streamView)
	}

	return focusable
}

func (r *TRunPage) switchBeforeRun(
	pages *tview.Pages,
	focusable []*misc.TItem,
	streamView *tview.TextView,
) []*misc.TItem {
	name, _ := pages.GetFrontPage()
	if name == "exec-projects" {
		pages.SwitchToPage("exec-run")
		focusable = r.updateStreamFocusable(streamView)
	}

	return focusable
}

func (r *TRunPage) runTasks(
	streamView *tview.TextView,
	taskData views.TTask,
	projectData views.TProject,
	spec *views.TSpec,
	ansiWriter *misc.ThreadSafeWriter,
) {
	// Check if any projects selected
	selectedProjects := []dao.Project{}
	for _, project := range projectData.Projects {
		if projectData.ProjectsSelected[project.Name] {
			selectedProjects = append(selectedProjects, project)
		}
	}
	if len(selectedProjects) < 1 {
		return
	}

	// Task
	var taskNames []string
	for _, task := range taskData.Tasks {
		if taskData.TasksSelected[task.Name] {
			taskNames = append(taskNames, task.Name)
		}
	}
	var projectNames []string
	for _, project := range selectedProjects {
		projectNames = append(projectNames, project.Name)
	}

	// Flags
	runFlags := core.RunFlags{
		Silent: true,

		// Filter
		Cwd:      false,
		All:      false,
		TagsExpr: "",

		Target: "default",
		Spec:   "default",

		Projects:          projectNames,
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

	// Parse Task
	var err error
	var tasks []dao.Task
	var projects []dao.Project
	if len(taskNames) == 1 {
		tasks, projects, err = dao.ParseSingleTask(taskNames[0], &runFlags, &setRunFlags, misc.Config)
	} else {
		tasks, projects, err = dao.ParseManyTasks(taskNames, &runFlags, &setRunFlags, misc.Config)
	}
	if err != nil {
		misc.App.Stop()
	}

	// Run task
	target := exec.Exec{Projects: projects, Tasks: tasks, Config: *misc.Config}

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

	err = target.RunTUI([]string{}, &runFlags, &setRunFlags, spec.Output, ansiWriter, ansiWriter)
	if err != nil {
		misc.App.Stop()
	}

	streamView.ScrollToEnd()
}
