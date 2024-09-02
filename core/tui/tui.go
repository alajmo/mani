package tui

import (
	"fmt"
	"strings"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var TUI = struct {
	app         *tview.Application
	pages       *tview.Pages
	topBar      *tview.TextView
	contentArea *tview.Pages
	searchInput *tview.InputField
	table       *tview.Table
	rightBox    *tview.Flex
	helpModal   *tview.Modal

	tagList          *tview.List
	selectedView     *tview.TextView
	allProjects      []dao.Project
	filteredProjects []dao.Project
}{}

func RunTui(config *dao.Config, args []string) {
	projects, err := config.FilterProjects(true, true, []string{}, []string{}, []string{})
	core.CheckIfError(err)

	TUI.allProjects = projects
	TUI.filteredProjects = projects

	// Create TUI
	setupStyles()
	TUI.app = tview.NewApplication()
	TUI.pages = tview.NewPages()
	TUI.table = createProjectTable(projects)

	setupComponents()
	configureInput()

	// Run TUI
	if err := TUI.app.SetRoot(TUI.pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func setupProjectPage() {

}

func setupComponents() {
	TUI.topBar = tview.NewTextView().
		SetDynamicColors(false).
		SetRegions(true).
		SetWrap(false)
	TUI.topBar.SetText("[-:b]  [\"projects\"]Projects[\"\"](p)  |  [\"tasks\"]Tasks[\"\"](t)  |  [\"output\"]Output[\"\"](o)  |  [\"help\"]Help[\"\"](?)")

	projectsFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.table, 0, 1, true)

	tasksBox := tview.NewTextView().
		SetText("Tasks content\n\nThis is where task information will be displayed.").
		SetBorder(false).
		SetTitle("Tasks")

	outputBox := tview.NewTextView().
		SetText("Output content\n\nThis is where output information will be displayed.").
		SetBorder(false).
		SetTitle("Output")

	TUI.contentArea = tview.NewPages().
		AddPage("projects", projectsFlex, true, true).
		AddPage("tasks", tasksBox, true, false).
		AddPage("output", outputBox, true, false)

	contentAreaWithBorder := tview.NewFlex().
		AddItem(TUI.contentArea, 0, 1, true)
	contentAreaWithBorder.SetBorder(true).SetBorderPadding(0, 0, 2, 2)

	TUI.selectedView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true)
	TUI.selectedView.SetTitle("Selected Projects").
		SetBorder(false)

	TUI.tagList = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetMainTextColor(tcell.ColorWhite)

	TUI.tagList.SetTitle("Filter by Tags").
		SetBorder(false)

	populateTagList()

	// Right Box
	// TUI.rightBox = tview.NewBox().SetTitle("Filter").SetBorder(true)

	TUI.rightBox = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.selectedView, 0, 1, false).
		AddItem(TUI.tagList, 0, 1, false)

	// TUI.rightBox = tview.NewTextView().
	// 	SetDynamicColors(true).
	// 	SetRegions(true).
	// 	SetWrap(true)

	TUI.rightBox.SetTitle("Filter").
		SetBorder(true)

	// TUI.rightBox = tview.NewTextView().
	// 	SetDynamicColors(true).
	// 	SetRegions(true).
	// 	SetWrap(true).
	// 	SetTitle("Filter").
	// 	SetBorder(true)

	TUI.rightBox.SetFocusFunc(func() {
		TUI.rightBox.SetBorderColor(tcell.ColorYellow)
	})
	TUI.rightBox.SetBlurFunc(func() {
		TUI.rightBox.SetBorderColor(tcell.ColorWhite)
	})

	// Table
	TUI.table.SetFocusFunc(func() {
		contentAreaWithBorder.SetBorderColor(tcell.ColorYellow)
	})
	TUI.table.SetBlurFunc(func() {
		contentAreaWithBorder.SetBorderColor(tcell.ColorWhite)
	})

	// TUI.helpModal = createHelpModal()

	// Create a flex for the search input (left-aligned)
	TUI.searchInput = tview.NewInputField().
		SetLabel("").
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorBlue)
		// SetBackgroundColor(tcell.ColorDefault)

	mainFlex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(TUI.topBar, 1, 1, false).
			AddItem(tview.NewFlex().
				AddItem(contentAreaWithBorder, 0, 3, true).
				AddItem(TUI.rightBox, 30, 1, false),
				0, 1, true).
			AddItem(TUI.searchInput, 1, 0, false),
			0, 20, true)

	TUI.pages.
		AddPage("main", mainFlex, true, true)
		// AddPage("help", TUI.helpModal, true, false)

	TUI.topBar.Highlight("projects")
	TUI.app.SetFocus(TUI.contentArea)
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
		TUI.contentArea,
		TUI.rightBox,
	}

	currentFocus := 0
	var lastSearchQuery string
	var lastFoundRow, lastFoundCol int
	searchDirection := 1 // 1 for forward, -1 for backward

	showSearch := func() {
		TUI.searchInput.SetLabel("search: ")
		TUI.searchInput.SetText("")
		TUI.app.SetFocus(TUI.searchInput)
	}

	hideSearch := func() {
		TUI.searchInput.SetLabel("")
		TUI.searchInput.SetText("")
		TUI.app.SetFocus(TUI.table)
	}

	TUI.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if the search input is currently focused
		if TUI.app.GetFocus() == TUI.searchInput {
			switch event.Key() {
			case tcell.KeyEscape:
				hideSearch()
				return nil
			case tcell.KeyEnter:
				query := TUI.searchInput.GetText()
				if query != "" {
					lastFoundRow, lastFoundCol = -1, -1
					searchDirection = 1
					searchNextInTable(TUI.table, query, &lastFoundRow, &lastFoundCol, searchDirection)
				}
				TUI.app.SetFocus(TUI.table)
				return nil
			}
			// Let all other keys be handled by the input field
			return event
		}

		// Handle other keys when search input is not focused
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
				TUI.contentArea.SwitchToPage("projects")
				TUI.topBar.Highlight("projects")
				hideSearch()
				return nil
			case 't', '2':
				TUI.contentArea.SwitchToPage("tasks")
				TUI.topBar.Highlight("tasks")
				hideSearch()
				return nil
			case 'o', '3':
				TUI.contentArea.SwitchToPage("output")
				TUI.topBar.Highlight("output")
				hideSearch()
				return nil
			case '?', '4':
				// TUI.pages.ShowPage("help")
				showHelpModal()
				hideSearch()
				return nil
			case '/':
				if TUI.contentArea.HasPage("projects") {
					showSearch()
					return nil
				}
			case 'n':
				if TUI.contentArea.HasPage("projects") && TUI.app.GetFocus() == TUI.table {
					query := TUI.searchInput.GetText()
					if query != "" {
						searchDirection = 1
						searchNextInTable(TUI.table, query, &lastFoundRow, &lastFoundCol, searchDirection)
					}
					return nil
				}
			case 'N':
				if TUI.contentArea.HasPage("projects") && TUI.app.GetFocus() == TUI.table {
					query := TUI.searchInput.GetText()
					if query != "" {
						searchDirection = -1
						searchNextInTable(TUI.table, query, &lastFoundRow, &lastFoundCol, searchDirection)
					}
					return nil
				}
			}
		}

		return event
	})

	TUI.searchInput.SetChangedFunc(func(text string) {
		if text != lastSearchQuery {
			lastSearchQuery = text
			lastFoundRow, lastFoundCol = -1, -1
			searchDirection = 1
			searchNextInTable(TUI.table, text, &lastFoundRow, &lastFoundCol, searchDirection)
		}
	})
}

func createProjectTable(projects []dao.Project) *tview.Table {
	table := tview.NewTable().SetBorders(false)
	table.SetSelectable(true, false)

	updateTableContent()

	table.SetFixed(1, 0)
	table.SetEvaluateAllRows(true)

	// Set up headers
	headers := []string{"Name", "Description", "Tags"}
	for col, header := range headers {
		table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAttributes(tcell.AttrBold).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Populate the table with project data
	for row, project := range projects {
		table.SetCell(row+1, 0, tview.NewTableCell(project.Name))
		table.SetCell(row+1, 1, tview.NewTableCell(project.Desc))
		tagsString := ""
		if len(project.Tags) > 0 {
			tagsString = strings.Join(project.Tags, ", ")
		}
		table.SetCell(row+1, 2, tview.NewTableCell(tagsString))
	}

	// Define the four states
	// Focused row and unselected (black background and blue text)
	// Focused row and selected (blue background and yellow text)
	// Unfocused row and selected (blue background and black text)
	// Unfocused row and selected (black background and white text)

	focusedUnselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlack)
	focusedSelectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorYellow).Attributes(tcell.AttrBold)
	unfocusedSelectedStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorRed).Attributes(tcell.AttrBold)
	unfocusedUnselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorDefault).Background(tcell.ColorDefault)

	// Keep track of selected rows
	selectedRows := make(map[int]bool)

	// Function to update cell styles
	updateCellStyles := func() {
		focusedRow, _ := table.GetSelection()
		if focusedRow == 0 {
			return
		}

		for row := 1; row < table.GetRowCount(); row++ {
			isSelected := selectedRows[row]
			isFocused := row == focusedRow
			var style tcell.Style

			if isFocused {
				if isSelected {
					style = focusedSelectedStyle
				} else {
					style = focusedUnselectedStyle
				}
			} else {
				if isSelected {
					style = unfocusedSelectedStyle
				} else {
					style = unfocusedUnselectedStyle
				}
			}

			for col := 0; col < table.GetColumnCount(); col++ {
				// core.DebugPrint(style)
				table.GetCell(row, col).SetStyle(style)
			}
		}

		updateSelectedProjectsDisplay(selectedRows)
	}

	isAllSelected := func() bool {
		for row := 1; row < table.GetRowCount(); row++ {
			if !selectedRows[row] {
				return false
			}
		}
		return true
	}

	toggleAllRows := func() {
		allSelected := isAllSelected()
		for row := 1; row < table.GetRowCount(); row++ {
			selectedRows[row] = !allSelected
		}
		updateCellStyles()
	}

	// Set up key handling for the table
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ':
				row, _ := table.GetSelection()
				if row > 0 {
					selectedRows[row] = !selectedRows[row]
					updateCellStyles()
				}
				return nil
			case 'd':
				row, _ := table.GetSelection()
				if row > 0 {
					showProjectDescriptionModal(projects[row-1])
				}
				return nil
			}
		case tcell.KeyCtrlA:
			toggleAllRows()
			return nil
		}
		return event
	})

	// Update styles when selection changes
	table.SetSelectionChangedFunc(func(row, column int) {
		updateCellStyles()
	})

	table.Select(1, 0)

	// Initial style update
	updateCellStyles()

	return table
}

func updateTableContent() {
	if TUI.table == nil {
		return
	}

	TUI.table.Clear()

	// Set up headers
	headers := []string{"Name", "Description", "Tags"}
	for col, header := range headers {
		TUI.table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAttributes(tcell.AttrBold).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Populate the table with project data
	for row, project := range TUI.filteredProjects {
		TUI.table.SetCell(row+1, 0, tview.NewTableCell(project.Name))
		TUI.table.SetCell(row+1, 1, tview.NewTableCell(project.Desc))
		tagsString := ""
		if len(project.Tags) > 0 {
			tagsString = strings.Join(project.Tags, ", ")
		}
		TUI.table.SetCell(row+1, 2, tview.NewTableCell(tagsString))
	}
}

func updateSelectedProjectsDisplay(selectedRows map[int]bool) {
	if TUI.table == nil {
		return
	}

	var selectedProjects strings.Builder
	selectedProjects.WriteString("Selected Projects:\n\n")

	for row, project := range TUI.filteredProjects {
		if selectedRows[row+1] {
			selectedProjects.WriteString(fmt.Sprintf("- %s\n", project.Name))
		}
	}

	TUI.selectedView.SetText(selectedProjects.String())
}

func populateTagList() {
	tagSet := make(map[string]bool)
	for _, project := range TUI.allProjects {
		for _, tag := range project.Tags {
			tagSet[tag] = false
		}
	}

	TUI.tagList.Clear()
	for tag := range tagSet {
		TUI.tagList.AddItem(tag, "", 0, nil)
	}

	TUI.tagList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		tagSet[mainText] = !tagSet[mainText]
		if tagSet[mainText] {
			TUI.tagList.SetItemText(index, "[yellow]✓ "+mainText, "")
		} else {
			TUI.tagList.SetItemText(index, mainText, "")
		}
		filterProjects(tagSet)
	})
}

func filterProjects(selectedTags map[string]bool) {
	TUI.filteredProjects = []dao.Project{}
	for _, project := range TUI.allProjects {
		include := true
		for tag, selected := range selectedTags {
			if selected {
				tagFound := false
				for _, projectTag := range project.Tags {
					if projectTag == tag {
						tagFound = true
						break
					}
				}
				if !tagFound {
					include = false
					break
				}
			}
		}
		if include {
			TUI.filteredProjects = append(TUI.filteredProjects, project)
		}
	}
	updateTableContent()
}

func searchNextInTable(table *tview.Table, query string, lastFoundRow, lastFoundCol *int, direction int) {
	rowCount := table.GetRowCount()
	colCount := table.GetColumnCount()

	startRow := *lastFoundRow
	if startRow == -1 {
		startRow = 0
	} else {
		startRow += direction
	}

	// Function to check if a row contains the query in any column
	checkRow := func(row int) (bool, int) {
		for col := 0; col < colCount; col++ {
			cell := table.GetCell(row, col)
			if cell == nil {
				continue
			}
			if strings.Contains(strings.ToLower(cell.Text), strings.ToLower(query)) {
				return true, col
			}
		}
		return false, -1
	}

	// Search forward
	if direction > 0 {
		for row := startRow; row < rowCount; row++ {
			found, col := checkRow(row)
			if found {
				table.Select(row, col)
				*lastFoundRow, *lastFoundCol = row, col
				return
			}
		}
		// If not found, start from the beginning
		if startRow != 0 {
			*lastFoundRow, *lastFoundCol = -1, -1
			searchNextInTable(table, query, lastFoundRow, lastFoundCol, direction)
		}
	} else { // Search backward
		for row := startRow; row >= 0; row-- {
			found, col := checkRow(row)
			if found {
				table.Select(row, col)
				*lastFoundRow, *lastFoundCol = row, col
				return
			}
		}
		// If not found, start from the end
		if startRow != rowCount-1 {
			*lastFoundRow, *lastFoundCol = rowCount, -1
			searchNextInTable(table, query, lastFoundRow, lastFoundCol, direction)
		}
	}
}

func setupStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.BorderColor = tcell.ColorWhite
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft
}

func checkAndHideVisiblePages() {
	frontPageName, _ := TUI.pages.GetFrontPage()

	if frontPageName == "help" {
		TUI.pages.HidePage("help")
		return
	}

	if frontPageName == "project-description" {
		TUI.pages.HidePage("project-description")
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
	TUI.pages.AddPage("help", flex, false, true)
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

func showProjectDescriptionModal(project dao.Project) {
	var sync = true
	if !sync {
		sync = false
	}

	description := fmt.Sprintf(`Name: %s
Path: %s
Description: %s
Url: %s
Sync: %v`,
		project.Name,
		project.Path,
		project.Desc,
		project.Url,
		sync,
	)

	if len(project.Tags) > 0 {
		description += fmt.Sprintf("\nTags: %s\n", strings.Join(project.Tags, ", "))
	}

	if len(project.EnvList) > 0 {
		description += printEnv(project.EnvList)
	}

	openModal(description, project.Name, 80)
}

func printEnv(env []string) string {
	envStr := "Env: \n"
	for _, env := range env {
		envStr += fmt.Sprintf("%4s%s\n", " ", strings.Replace(strings.TrimSuffix(env, "\n"), "=", ": ", 1))
	}

	return envStr
}
