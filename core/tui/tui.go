package tui

import (
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/pages"
)

func RunTui(config *dao.Config, args []string) {
	// Globals
	misc.Config = config

	// Data
	projects := config.ProjectList
	tasks := config.TaskList
	projectTags := config.GetTags()
	projectPaths := config.GetProjectPaths()

	// Styles
	setupStyles()

	// Create pages
	misc.App = tview.NewApplication()
	misc.Pages = tview.NewPages()
	createPages(projects, projectTags, projectPaths, tasks)

	// Global input handling
	HandleInput()

	// Run TUI
	if err := misc.App.SetRoot(misc.Pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func createPages(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
	tasks []dao.Task,
) {
	navPane := createNav()
	search := components.CreateSearchInput()
	misc.Search = search

	projectsPage := pages.CreateProjectsPage(projects, projectTags, projectPaths)
	tasksPage := pages.CreateTasksPage(tasks)
	runPage := pages.CreateRunPage(tasks, projects, projectTags, projectPaths)
	execPage := pages.CreateExecPage(projects, projectTags, projectPaths)

	misc.MainPage = tview.NewPages().
		AddPage("projects", projectsPage, true, true).
		AddPage("tasks", tasksPage, true, false).
		AddPage("run", runPage, true, false).
		AddPage("exec", execPage, true, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(navPane, 1, 1, false).
		AddItem(misc.MainPage, 0, 1, true)
	misc.Pages.AddPage("main", mainLayout, true, true)

	misc.SwitchToPage("projects")
}

func createNav() *tview.Flex {
	// Buttons
	misc.ProjectBtn = misc.CreateButton("Projects")
	misc.ProjectBtn.SetSelectedFunc(func() {
		misc.SwitchToPage("projects")
	})

	misc.TaskBtn = misc.CreateButton("Tasks")
	misc.TaskBtn.SetSelectedFunc(func() {
		misc.SwitchToPage("tasks")
	})

	misc.RunBtn = misc.CreateButton("Run")
	misc.RunBtn.SetSelectedFunc(func() {
		misc.SwitchToPage("run")
	})

	misc.ExecBtn = misc.CreateButton("Exec")
	misc.ExecBtn.SetSelectedFunc(func() {
		misc.SwitchToPage("exec")
	})

	misc.HelpBtn = misc.CreateButton("Help")
	misc.HelpBtn.SetSelectedFunc(func() {
		components.ShowHelpModal()
	})

	// Left
	left := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(misc.ProjectBtn, 8, 0, false).
		AddItem(tview.NewTextView().SetText("  |  "), 5, 0, false).
		AddItem(misc.TaskBtn, 5, 0, false).
		AddItem(tview.NewTextView().SetText("  |"), 5, 0, false).
		AddItem(misc.RunBtn, 4, 0, false).
		AddItem(tview.NewTextView().SetText("  |"), 5, 0, false).
		AddItem(misc.ExecBtn, 4, 0, false)

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
	navPane.SetBorderPadding(0, 0, 1, 1)

	return navPane
}

func setupStyles() {
	// Foreground / Background
	tview.Styles.PrimitiveBackgroundColor = misc.THEME.BG

	// Borders Colors
	tview.Styles.BorderColor = misc.THEME.BORDER_COLOR

	// Border style
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft
}
