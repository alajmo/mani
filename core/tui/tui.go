package tui

import (
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var TUI = struct {
	app          *tview.Application
	topBar       *tview.TextView
	allPages     *tview.Pages
	viewPage     *tview.Pages
	inputSearch  *tview.InputField
	previousPage tview.Primitive

	// Projects
	projectsTable        *tview.Table
	projectsContextPage  *tview.Flex
	projectsTagsView     *tview.List
	projectsPathsView    *tview.List
	projectsSelectedView *tview.List

	projectsTagsFiltered  map[string]bool
	projectsPathsFiltered map[string]bool
	projectTags           []string
	projectPaths          []string
	projectsAll           []dao.Project
	projectsFiltered      []dao.Project
	projectsSelected      []dao.Project

	// Tasks
	tasksTable        *tview.Table
	taskssFilterPage  *tview.Flex
	tasksTagsView     *tview.List
	tasksSelectedView *tview.TextView
	tasksInputSearch  *tview.InputField

	tasksAll      []dao.Project
	tasksFiltered []dao.Project

	// Output

	// Misc
	helpModal *tview.Modal
}{}

func RunTui(config *dao.Config, args []string) {
	projects := config.ProjectList

	TUI.projectsAll = projects
	TUI.projectTags = config.GetProjectPaths()
	TUI.projectPaths = config.GetProjectPaths()

	// Create TUI
	setupStyles()

	TUI.app = tview.NewApplication()
	TUI.allPages = tview.NewPages()

	TUI.projectsTagsFiltered = make(map[string]bool)
	TUI.projectsPathsFiltered = make(map[string]bool)

	setupPages(config, projects, config.TaskList)
	configureInput()

	// Run TUI
	if err := TUI.app.SetRoot(TUI.allPages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func setupTasksPage(tasks []dao.Task) *tview.Box {
	tasksPage := tview.NewTextView().
		SetText("Tasks content\n\nThis is where task information will be displayed.").
		SetBorder(false).
		SetTitle("Tasks")

	return tasksPage
}

func setupOutputPage() {
	// outputBox := tview.NewTextView().
	// 	SetText("Output content\n\nThis is where output information will be displayed.").
	// 	SetBorder(false).
	// 	SetTitle("Output")
}

func setupPages(config *dao.Config, projects []dao.Project, tasks []dao.Task) {
	TUI.topBar = tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(true).
		SetWrap(false)
	TUI.topBar.SetText("[-:b]  [\"projects\"]Projects[\"\"]  |  [\"tasks\"]Tasks[\"\"]  |  [\"output\"]Output[\"\"]  |  [\"help\"]Help[\"\"]")

	projectPage := setupProjectPage(config, projects)

	TUI.viewPage = tview.NewPages().
		AddPage("projects", projectPage, true, true)

	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.topBar, 1, 1, false).
		AddItem(TUI.viewPage, 0, 1, true)
	TUI.allPages.AddPage("main", mainLayout, true, true)

	TUI.topBar.Highlight("projects")
}

func createHelpModal() *tview.Modal {
	helpText := "Keyboard Shortcuts:\n" +
		"q: Quit\n" +
		"p or 1: Switch to Projects\n" +
		"t or 2: Switch to Tasks\n" +
		"o or 3: Switch to Output\n" +
		"d: View project\n" +
		"?: Show this Help\n" +
		"escape: Close Help"

	modal := tview.NewModal().SetText(helpText)

	modal.SetTitle("Help")
	modal.SetBackgroundColor(tcell.ColorDefault)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetBorderColor(tcell.ColorYellow)
	modal.SetBorderPadding(1, 1, 1, 1)
	modal.Box.SetBackgroundColor(tcell.ColorDefault)

	return modal
}

func configureInput() {
	focusableElements := []tview.Primitive{
		TUI.viewPage,
		TUI.projectsTagsView,
		TUI.projectsPathsView,
		TUI.projectsSelectedView,
	}

	currentFocus := 0
	var lastSearchQuery string
	var lastFoundRow, lastFoundCol int
	searchDirection := 1

	showSearch := func() {
		TUI.inputSearch.SetLabel("search: ")
		TUI.inputSearch.SetText("")
		TUI.app.SetFocus(TUI.inputSearch)
	}

	hideSearch := func() {
		TUI.inputSearch.SetLabel("")
		TUI.inputSearch.SetText("")
	}

	TUI.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Search Input
		if TUI.app.GetFocus() == TUI.inputSearch {
			lastFoundRow, lastFoundCol = -1, -1

			switch event.Key() {
			case tcell.KeyEscape:
				hideSearch()
				return nil

			case tcell.KeyEnter:
				{
					query := TUI.inputSearch.GetText()
					if query == "" {
						return nil
					}

					switch TUI.previousPage {
					case TUI.projectsTable:
						TUI.app.SetFocus(TUI.projectsTable)
						searchInTable(TUI.projectsTable, query, &lastFoundRow, &lastFoundCol, searchDirection)

					case TUI.projectsTagsView:
						TUI.app.SetFocus(TUI.projectsTagsView)
						searchInList(TUI.projectsTagsView, query, &lastFoundRow, searchDirection)

					case TUI.projectsPathsView:
						TUI.app.SetFocus(TUI.projectsPathsView)
						searchInList(TUI.projectsPathsView, query, &lastFoundRow, searchDirection)

					case TUI.projectsSelectedView:
						TUI.app.SetFocus(TUI.projectsSelectedView)
						searchInList(TUI.projectsSelectedView, query, &lastFoundRow, searchDirection)
					}

					return nil
				}
			}
			return event
		}

		// Main
		switch event.Key() {
		case tcell.KeyEscape:
			checkAndHideVisiblePages()
			hideSearch()
			return nil
		case tcell.KeyTab:
			currentFocus = (currentFocus + 1) % len(focusableElements)
			TUI.app.SetFocus(focusableElements[currentFocus])
			return nil
		case tcell.KeyBacktab:
			currentFocus = (currentFocus - 1 + len(focusableElements)) % len(focusableElements)
			TUI.app.SetFocus(focusableElements[currentFocus])
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				TUI.app.Stop()
				return nil
			case 'p', '1':
				TUI.viewPage.SwitchToPage("projects")
				TUI.topBar.Highlight("projects")
				hideSearch()
				return nil
			case 't', '2':
				TUI.viewPage.SwitchToPage("tasks")
				TUI.topBar.Highlight("tasks")
				hideSearch()
				return nil
			case 'o', '3':
				TUI.viewPage.SwitchToPage("output")
				TUI.topBar.Highlight("output")
				hideSearch()
				return nil
			case '?', '4':
				showHelpModal()
				hideSearch()
				return nil
			case '/':
				showSearch()
				return nil
			case 'n':
				{
					query := TUI.inputSearch.GetText()
					if query == "" {
						return nil
					}
					searchDirection = 1

					switch TUI.app.GetFocus() {
					case TUI.projectsTable:
						searchInTable(TUI.projectsTable, query, &lastFoundRow, &lastFoundCol, searchDirection)
						return nil
					case TUI.projectsTagsView:
						searchInList(TUI.projectsTagsView, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsPathsView:
						searchInList(TUI.projectsPathsView, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsSelectedView:
						searchInList(TUI.projectsSelectedView, query, &lastFoundRow, searchDirection)
						return nil
					}
				}

			case 'N':
				{
					query := TUI.inputSearch.GetText()
					if query == "" {
						return nil
					}
					searchDirection = -1

					switch TUI.app.GetFocus() {
					case TUI.projectsTable:
						searchInTable(TUI.projectsTable, query, &lastFoundRow, &lastFoundCol, searchDirection)
						return nil
					case TUI.projectsTagsView:
						searchInList(TUI.projectsTagsView, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsPathsView:
						searchInList(TUI.projectsPathsView, query, &lastFoundRow, searchDirection)
						return nil
					case TUI.projectsSelectedView:
						searchInList(TUI.projectsSelectedView, query, &lastFoundRow, searchDirection)
						return nil
					}
				}
			}
		}

		return event
	})

	TUI.inputSearch.SetChangedFunc(func(text string) {
		if text != lastSearchQuery {
			lastSearchQuery = text
			lastFoundRow, lastFoundCol = -1, -1
			searchDirection = 1

			switch TUI.previousPage {
			case TUI.projectsTable:
				searchInTable(TUI.projectsTable, text, &lastFoundRow, &lastFoundCol, searchDirection)
			case TUI.projectsTagsView:
				searchInList(TUI.projectsTagsView, text, &lastFoundRow, searchDirection)
			case TUI.projectsPathsView:
				searchInList(TUI.projectsPathsView, text, &lastFoundRow, searchDirection)
			case TUI.projectsSelectedView:
				searchInList(TUI.projectsSelectedView, text, &lastFoundRow, searchDirection)
			}
		}
	})
}

func setupStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.BorderColor = tcell.ColorWhite

	tview.Styles.BorderColor = tcell.ColorWhite

	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft

	// tview.Styles.(tcell.ColorDefault)
}

func checkAndHideVisiblePages() {
	frontPageName, _ := TUI.allPages.GetFrontPage()

	if frontPageName == "help" {
		TUI.allPages.HidePage("help")
		return
	}

	if frontPageName == "project-description" {
		TUI.allPages.HidePage("project-description")
		return
	}
}

func openModal(text string, title string, width int) {
	textView := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	textView.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorYellow).
		SetBorderPadding(1, 1, 2, 2)
	textView.SetBackgroundColor(tcell.ColorDefault)
	textView.SetTextColor(tcell.ColorWhite)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().
				AddItem(nil, 0, 1, false).
				AddItem(textView, width, 1, true).
				AddItem(nil, 0, 1, false),
			15, 1, true,
		).
		AddItem(nil, 0, 1, false)
	flex.SetFullScreen(true).SetBackgroundColor(tcell.ColorBlack)
	TUI.allPages.AddPage("help", flex, false, true)
	TUI.app.SetFocus(textView)
}

func showHelpModal() {
	helpText := "Keyboard Shortcuts:\n" +
		"q: Quit\n" +
		"p or 1: Switch to Projects\n" +
		"t or 2: Switch to Tasks\n" +
		"o or 3: Switch to Output\n" +
		"d: View project\n" +
		"?: Show this Help\n" +
		"escape: Close Help"

	openModal(helpText, "Help", 50)
}

func printEnv(env []string) string {
	envStr := "Env: \n"
	for _, env := range env {
		envStr += fmt.Sprintf("%4s%s\n", " ", strings.Replace(strings.TrimSuffix(env, "\n"), "=", ": ", 1))
	}

	return envStr
}

func searchInList(list *tview.List, query string, lastFoundIndex *int, direction int) {
	itemCount := list.GetItemCount()
	startIndex := *lastFoundIndex

	if startIndex == -1 {
		startIndex = 0
	} else {
		startIndex += direction
	}

	searchIndex := startIndex
	for i := 0; i < itemCount; i++ {
		if searchIndex < 0 {
			searchIndex = itemCount - 1
		} else if searchIndex >= itemCount {
			searchIndex = 0
		}

		mainText, secondaryText := list.GetItemText(searchIndex)
		if strings.Contains(strings.ToLower(mainText), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(secondaryText), strings.ToLower(query)) {
			list.SetCurrentItem(searchIndex)
			*lastFoundIndex = searchIndex
			return
		}

		searchIndex += direction
	}

	*lastFoundIndex = -1
}
