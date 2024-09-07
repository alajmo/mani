package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var version = "dev"

var TUI = struct {
	config *dao.Config

	app          *tview.Application
	navView      *tview.Flex
	pages        *tview.Pages
	mainPage     *tview.Pages
	search       *tview.InputField
	previousPage tview.Primitive

	// Nav
	projectBtn *tview.Button
	taskBtn    *tview.Button
	runBtn     *tview.Button
	helpBtn    *tview.Button

	// Projects
	projectsPage         *tview.Flex
	projectsTable        *tview.Table
	projectsContextPage  *tview.Flex
	projectsTagsView     *tview.List
	projectsPathsView    *tview.List
	projectsSelectedView *tview.List

	projectsTagsFiltered  map[string]bool
	projectsPathsFiltered map[string]bool
	projectTags           []string
	projectPaths          []string
	projects              []dao.Project
	projectsFiltered      []dao.Project
	projectsSelected      []dao.Project

	// Tasks
	tasksPage         *tview.Flex
	tasksTable        *tview.Table
	taskssFilterPage  *tview.Flex
	tasksTagsView     *tview.List
	tasksSelectedView *tview.TextView

	tasksAll      []dao.Project
	tasksFiltered []dao.Project

	// Run
	runPage *tview.Flex

	// Misc
	helpModal *tview.Modal
}{}

func RunTui(config *dao.Config, args []string) {
	// Setup data
	TUI.config = config
	TUI.projects = config.ProjectList
	TUI.projectTags = config.GetTags()
	TUI.projectPaths = config.GetProjectPaths()
	TUI.projectsTagsFiltered = make(map[string]bool)
	TUI.projectsPathsFiltered = make(map[string]bool)

	// Set Styles
	setupStyles()

	// Create pages
	TUI.app = tview.NewApplication()
	TUI.pages = tview.NewPages()

	// Create pages
	createPages()

	// Handle input
	handleInput()

	// Run TUI
	if err := TUI.app.SetRoot(TUI.pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func createPages() {
	createNav()
	createProjectsPage()
	createTasksPage()
	createRunPage()

	TUI.mainPage = tview.NewPages().
		AddPage("projects", TUI.projectsPage, true, true).
		AddPage("tasks", TUI.tasksPage, true, false).
		AddPage("run", TUI.runPage, true, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.navView, 1, 1, false).
		AddItem(TUI.mainPage, 0, 1, true)
	TUI.pages.AddPage("main", mainLayout, true, true)

	switchToPage("projects")
}

func createNav() {
	// Buttons
	TUI.projectBtn = createButton("Projects")
	TUI.projectBtn.SetSelectedFunc(func() {
		switchToPage("projects")
	})

	TUI.taskBtn = createButton("Tasks")
	TUI.taskBtn.SetSelectedFunc(func() {
		switchToPage("tasks")
	})

	TUI.runBtn = createButton("Run")
	TUI.runBtn.SetSelectedFunc(func() {
		switchToPage("run")
	})

	TUI.helpBtn = createButton("Help")
	TUI.helpBtn.SetSelectedFunc(func() {
		showHelpModal()
	})

	// Left
	left := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(TUI.projectBtn, 8, 0, false).
		AddItem(tview.NewTextView().SetText("  |  "), 5, 0, false).
		AddItem(TUI.taskBtn, 5, 0, false).
		AddItem(tview.NewTextView().SetText("  |"), 5, 0, false).
		AddItem(TUI.runBtn, 4, 0, false)

	// Right
	right := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(TUI.helpBtn, 5, 0, false)

	// Nav
	TUI.navView = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(left, 0, 1, false).
		AddItem(nil, 0, 1, false).
		AddItem(right, 4, 0, false)
	TUI.navView.SetBorderPadding(0, 0, 1, 1)
}

func setupStyles() {
	// Foreground / Background
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault

	// Borders Colors
	tview.Styles.BorderColor = tcell.ColorWhite
	tview.Styles.BorderColor = tcell.ColorWhite

	// Border style
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft
}
