package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/pages"
	"github.com/rivo/tview"
)

var version = "dev"

var TUI = struct {
	config  *dao.Config
	emitter *EventEmitter

	app          *tview.Application
	navPane      *tview.Flex
	pages        *tview.Pages
	mainPage     *tview.Pages
	previousPage tview.Primitive
	search       *tview.InputField

	// Nav
	projectBtn *tview.Button
	taskBtn    *tview.Button
	runBtn     *tview.Button
	execBtn    *tview.Button
	helpBtn    *tview.Button

	// Run
	runPage *tview.Flex

	// Exec
	execPage *tview.Flex

	// Misc
	helpModal *tview.Modal
}{}

func RunTui(config *dao.Config, args []string) {
	// Setup data
	TUI.config = config
	TUI.emitter = NewEventEmitter()

	projects := config.ProjectList
	tasks := config.TaskList
	projectTags := config.GetTags()
	projectPaths := config.GetProjectPaths()

	// Set Styles
	setupStyles()

	// Create pages
	TUI.app = tview.NewApplication()
	TUI.pages = tview.NewPages()

	// Create pages
	createPages(projects, projectTags, projectPaths, tasks)

	// Handle input
	handleInput()

	// Run TUI
	if err := TUI.app.SetRoot(TUI.pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func createPages(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
	tasks []dao.Task,
) {
	createNav()
	createSearchInput()
	projectsPage := pages.CreateProjectsPage(projects, projectTags, projectPaths)
	tasksPage := createTasksPage(tasks)
	createRunPage()
	createExecPage()

	TUI.mainPage = tview.NewPages().
		AddPage("projects", &projectsPage, true, true).
		AddPage("tasks", &tasksPage, true, false).
		AddPage("run", TUI.runPage, true, false).
		AddPage("exec", TUI.execPage, true, false)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.navPane, 1, 1, false).
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

	TUI.execBtn = createButton("Exec")
	TUI.execBtn.SetSelectedFunc(func() {
		switchToPage("exec")
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
		AddItem(TUI.runBtn, 4, 0, false).
		AddItem(tview.NewTextView().SetText("  |"), 5, 0, false).
		AddItem(TUI.execBtn, 4, 0, false)

	// Right
	right := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(TUI.helpBtn, 5, 0, false)

	// Nav
	TUI.navPane = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(left, 0, 1, false).
		AddItem(nil, 0, 1, false).
		AddItem(right, 4, 0, false)
	TUI.navPane.SetBorderPadding(0, 0, 1, 1)
}

func setupStyles() {
	// Foreground / Background
	tview.Styles.PrimitiveBackgroundColor = THEME.BG

	// Borders Colors
	tview.Styles.BorderColor = THEME.BORDER_COLOR

	// Border style
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft
}
