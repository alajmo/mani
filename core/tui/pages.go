package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/pages"
	"github.com/alajmo/mani/core/tui/views"
	"github.com/rivo/tview"
)

func createPages(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
	tasks []dao.Task,
) *tview.Pages {
	appPages := tview.NewPages()
	navPane := createNav()
	search := components.CreateSearch()
	misc.Search = search

	projectsPage := pages.CreateProjectsPage(projects, projectTags, projectPaths)
	tasksPage := pages.CreateTasksPage(tasks)
	runPage := pages.CreateRunPage(tasks, projects, projectTags, projectPaths)
	execPage := pages.CreateExecPage(projects, projectTags, projectPaths)

	misc.MainPage = tview.NewPages().
		AddPage("run", runPage, true, true).
		AddPage("exec", execPage, true, false).
		AddPage("projects", projectsPage, true, false).
		AddPage("tasks", tasksPage, true, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(navPane, 2, 1, false).
		AddItem(misc.MainPage, 0, 1, true)
	appPages.AddPage("main", mainLayout, true, true)

	SwitchToPage("run")

	return appPages
}

func createNav() *tview.Flex {
	// Buttons
	misc.ProjectBtn = components.CreateButton("Projects")
	misc.ProjectBtn.SetSelectedFunc(func() {
		SwitchToPage("projects")
		misc.App.SetFocus(*misc.ProjectsLastFocus)
	})

	misc.TaskBtn = components.CreateButton("Tasks")
	misc.TaskBtn.SetSelectedFunc(func() {
		SwitchToPage("tasks")
		misc.App.SetFocus(*misc.TasksLastFocus)
	})

	misc.RunBtn = components.CreateButton("Run")
	misc.RunBtn.SetSelectedFunc(func() {
		SwitchToPage("run")
		misc.App.SetFocus(*misc.RunLastFocus)
	})

	misc.ExecBtn = components.CreateButton("Exec")
	misc.ExecBtn.SetSelectedFunc(func() {
		SwitchToPage("exec")
		misc.App.SetFocus(*misc.ExecLastFocus)
	})

	misc.HelpBtn = components.CreateButton("Help")
	misc.HelpBtn.SetSelectedFunc(func() {
		views.ShowHelpModal()
	})

	// Left
	left := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(misc.RunBtn, 7, 0, false).      // 3 size + 2 padding
		AddItem(misc.ExecBtn, 8, 0, false).     // 4 size + 2 padding
		AddItem(misc.ProjectBtn, 12, 0, false). // 8 size + 2 padding
		AddItem(misc.TaskBtn, 9, 0, false)      // 5 size + 2 padding

	// Right
	right := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(misc.HelpBtn, 5, 0, false)

		// Nav
	navPane := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(left, 0, 1, false).
		AddItem(nil, 0, 1, false).
		AddItem(right, 4, 0, false)
	navPane.SetBorderPadding(0, 1, 1, 1)

	return navPane
}

func SwitchToPage(pageName string) {
	misc.MainPage.SwitchToPage(pageName)

	switch pageName {
	case "projects":
		components.SetActiveButtonStyle(misc.ProjectBtn)

		components.SetInactiveButtonStyle(misc.HelpBtn)
		components.SetInactiveButtonStyle(misc.RunBtn)
		components.SetInactiveButtonStyle(misc.TaskBtn)
		components.SetInactiveButtonStyle(misc.ExecBtn)
	case "tasks":
		components.SetActiveButtonStyle(misc.TaskBtn)

		components.SetInactiveButtonStyle(misc.HelpBtn)
		components.SetInactiveButtonStyle(misc.ProjectBtn)
		components.SetInactiveButtonStyle(misc.RunBtn)
		components.SetInactiveButtonStyle(misc.ExecBtn)
	case "run":
		components.SetActiveButtonStyle(misc.RunBtn)

		components.SetInactiveButtonStyle(misc.HelpBtn)
		components.SetInactiveButtonStyle(misc.ProjectBtn)
		components.SetInactiveButtonStyle(misc.TaskBtn)
		components.SetInactiveButtonStyle(misc.ExecBtn)
	case "exec":
		components.SetActiveButtonStyle(misc.ExecBtn)

		components.SetInactiveButtonStyle(misc.HelpBtn)
		components.SetInactiveButtonStyle(misc.ProjectBtn)
		components.SetInactiveButtonStyle(misc.TaskBtn)
		components.SetInactiveButtonStyle(misc.RunBtn)
	}

	_, page := misc.MainPage.GetFrontPage()
	misc.App.SetFocus(page)
}

func setupStyles() {
	// Foreground / Background
	tview.Styles.PrimaryTextColor = misc.STYLE_DEFAULT.Fg
	tview.Styles.PrimitiveBackgroundColor = misc.STYLE_DEFAULT.Bg

	// Borders Colors
	tview.Styles.BorderColor = misc.STYLE_BORDER.Fg

	// Border style
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft
}
