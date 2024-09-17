package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var version = "dev"

var THEME = struct {
	FG                  tcell.Color
	BG                  tcell.Color
	FG_FOCUSED          tcell.Color
	BG_FOCUSED          tcell.Color
	FG_FOCUSED_SELECTED tcell.Color
	BG_FOCUSED_SELECTED tcell.Color

	BORDER_COLOR       tcell.Color
	BORDER_COLOR_FOCUS tcell.Color

	TITLE        tcell.Color
	TITLE_ACTIVE tcell.Color

	TABLE_HEADER_FG tcell.Color

	SEARCH_BG tcell.Color
	SEARCH_FG tcell.Color

	BTN_FG        tcell.Color
	BTN_BG        tcell.Color
	BTN_FG_ACTIVE tcell.Color
	BTN_BG_ACTIVE tcell.Color
}{
	FG:                  tcell.ColorDefault,
	BG:                  tcell.ColorDefault,
	FG_FOCUSED:          tcell.ColorWhite,
	BG_FOCUSED:          tcell.Color235,
	FG_FOCUSED_SELECTED: tcell.ColorBlue,
	BG_FOCUSED_SELECTED: tcell.Color235,

	BORDER_COLOR:       tcell.ColorWhite,
	BORDER_COLOR_FOCUS: tcell.ColorYellow,

	TITLE:        tcell.ColorDefault,
	TITLE_ACTIVE: tcell.ColorYellow,

	TABLE_HEADER_FG: tcell.ColorYellow,

	SEARCH_BG: tcell.ColorDefault,
	SEARCH_FG: tcell.ColorBlue,

	BTN_FG:        tcell.ColorWhite,
	BTN_BG:        tcell.ColorDefault,
	BTN_FG_ACTIVE: tcell.ColorYellow,
	BTN_BG_ACTIVE: tcell.ColorDefault,
}

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

	// Projects
	projectsPage         *tview.Flex
	projectsTable        *tview.Table
	projectsContextPage  *tview.Flex
	projectsTagsPane     *tview.List
	projectsPathsPane    *tview.List
	projectsSelectedPane *tview.List

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
	runContextPage    *tview.Flex
	tasksSelectedPane *tview.List

	tasks         []dao.Task
	tasksFiltered []dao.Task
	tasksSelected []dao.Task

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
	TUI.projects = config.ProjectList
	TUI.tasks = config.TaskList

	TUI.projectTags = config.GetTags()
	TUI.projectsTagsFiltered = make(map[string]bool)
	for _, tag := range TUI.projectTags {
		TUI.projectsTagsFiltered[tag] = false
	}

	TUI.projectPaths = config.GetProjectPaths()
	TUI.projectsPathsFiltered = make(map[string]bool)
	for _, projectPath := range TUI.projectPaths {
		TUI.projectsPathsFiltered[projectPath] = false
	}

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
	createSearchInput()
	createProjectsPage()
	createTasksPage()
	createRunPage()
  createExecPage()

	TUI.mainPage = tview.NewPages().
		AddPage("projects", TUI.projectsPage, true, true).
		AddPage("tasks", TUI.tasksPage, true, false).
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
