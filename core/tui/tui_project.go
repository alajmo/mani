package tui

import (
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setupProjectPage(config *dao.Config, projects []dao.Project) *tview.Flex {
	// Poulate project data
	TUI.projectsFiltered = projects
	TUI.projectsTable = createProjectTable(projects)

	// Project tags
	TUI.projectsTagsView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		// SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlack).Attributes(tcell.AttrBold)).
		SetMainTextColor(tcell.ColorWhite)
	TUI.projectsTagsView.SetTitle("Tags").SetBorder(true)
	populateTagList(config)

	// Project paths
	TUI.projectsPathsView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		// SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlack).Attributes(tcell.AttrBold)).
		SetMainTextColor(tcell.ColorWhite)
	TUI.projectsPathsView.SetTitle("Paths").SetBorder(true)
	populatePathsList(config)

	// Selected projects
	TUI.projectsSelectedView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true)
	TUI.projectsSelectedView.SetTitle("Selected").
		SetBorder(true)

	// Projects context
	TUI.projectsContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.projectsTagsView, 0, 1, true).
		AddItem(TUI.projectsPathsView, 0, 1, true).
		AddItem(TUI.projectsSelectedView, 0, 1, true)

	// Project search
	TUI.projectsInputSearch = tview.NewInputField().
		SetLabel("").
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorBlue)

	projectsPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(TUI.projectsTable, 0, 1, true).
				AddItem(TUI.projectsContextPage, 30, 1, false),
			0, 1, true).
		AddItem(TUI.projectsInputSearch, 1, 0, false)

	// Callbacks
	TUI.projectsTable.SetFocusFunc(func() {
		TUI.projectsTable.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsTable.SetBlurFunc(func() {
		TUI.projectsTable.SetBorderColor(tcell.ColorWhite)
	})

	TUI.projectsTagsView.SetFocusFunc(func() {
		TUI.projectsTagsView.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsTagsView.SetBlurFunc(func() {
		TUI.projectsTagsView.SetBorderColor(tcell.ColorWhite)
	})

	TUI.projectsPathsView.SetFocusFunc(func() {
		TUI.projectsPathsView.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsPathsView.SetBlurFunc(func() {
		TUI.projectsPathsView.SetBorderColor(tcell.ColorWhite)
	})

	TUI.projectsContextPage.SetFocusFunc(func() {
		TUI.projectsContextPage.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsContextPage.SetBlurFunc(func() {
		TUI.projectsContextPage.SetBorderColor(tcell.ColorWhite)
	})

	return projectsPage
}

func createProjectTable(projects []dao.Project) *tview.Table {
	table := tview.NewTable()
	table.SetBorder(true).SetBorderPadding(0, 0, 2, 2)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorDefault)
	// table.SetSelectedStyle(
	// 	tcell.StyleDefault.
	// 		Background(tcell.ColorBlack).
	// 		Foreground(tcell.ColorWhite))

	updateTableContent()

	// Fixed header
	table.SetFixed(1, 0)

	// Avoid resizing of headers when scrolling
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

	// SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlack)).
	focusedUnselectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
	focusedSelectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue).Attributes(tcell.AttrBold)
	unfocusedSelectedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorRed).Attributes(tcell.AttrBold)
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

	// Callbacks
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

	// Update styles when selection changes
	table.SetSelectionChangedFunc(func(row, column int) {
		updateCellStyles()
	})

	// Key inputs
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

	// Select first row
	table.Select(1, 0)

	// Initial style update
	updateCellStyles()

	return table
}

func updateTableContent() {
	if TUI.projectsTable == nil {
		return
	}

	TUI.projectsTable.Clear()

	// Set up headers
	headers := []string{"Name", "Description", "Tags"}
	for col, header := range headers {
		TUI.projectsTable.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAttributes(tcell.AttrBold).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Populate the table with project data
	for row, project := range TUI.projectsFiltered {
		TUI.projectsTable.SetCell(row+1, 0, tview.NewTableCell(project.Name))
		TUI.projectsTable.SetCell(row+1, 1, tview.NewTableCell(project.Desc))
		tagsString := ""
		if len(project.Tags) > 0 {
			tagsString = strings.Join(project.Tags, ", ")
		}
		TUI.projectsTable.SetCell(row+1, 2, tview.NewTableCell(tagsString))
	}
}

func updateSelectedProjectsDisplay(selectedRows map[int]bool) {
	if TUI.projectsTable == nil {
		return
	}

	var selectedProjects strings.Builder
	// selectedProjects.WriteString("Selected:\n\n")

	for row, project := range TUI.projectsFiltered {
		if selectedRows[row+1] {
			selectedProjects.WriteString(fmt.Sprintf("%s\n", project.Name))
		}
	}

	TUI.projectsSelectedView.SetText(selectedProjects.String())
}

func populateTagList(config *dao.Config) {
	for _, project := range TUI.projectsAll {
		for _, tag := range project.Tags {
			TUI.projectsTagsFiltered[tag] = false
		}
	}

	TUI.projectsTagsView.Clear()
	for tag := range TUI.projectsTagsFiltered {
		TUI.projectsTagsView.AddItem(tag, tag, 0, nil)
	}

	TUI.projectsTagsView.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		TUI.projectsTagsFiltered[secondaryText] = !TUI.projectsTagsFiltered[secondaryText]
		if TUI.projectsTagsFiltered[secondaryText] {
			TUI.projectsTagsView.SetItemText(index, "[blue::b]"+mainText, secondaryText)
		} else {
			TUI.projectsTagsView.SetItemText(index, secondaryText, secondaryText)
		}

		filterProjects(config)
	})

	TUI.projectsTagsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		numItems := TUI.projectsTagsView.GetItemCount()
		currentItem := TUI.projectsTagsView.GetCurrentItem()

		switch event.Rune() {
		case 'g':
			TUI.projectsTagsView.SetCurrentItem(0)
			return nil
		case 'G':
			TUI.projectsTagsView.SetCurrentItem(numItems - 1)
			return nil
		case 'j':
			nextItem := currentItem + 1
			if nextItem < numItems {
				TUI.projectsTagsView.SetCurrentItem(nextItem)
			}
			return nil
		case 'k':
			nextItem := currentItem - 1
			if nextItem >= 0 {
				TUI.projectsTagsView.SetCurrentItem(nextItem)
			}
			return nil
		}

		return event
	})
}

func populatePathsList(config *dao.Config) {
	for _, projectPath := range TUI.projectPaths {
		TUI.projectsPathsFiltered[projectPath] = false
	}

	TUI.projectsPathsView.Clear()
	for tag := range TUI.projectsPathsFiltered {
		TUI.projectsPathsView.AddItem(tag, tag, 0, nil)
	}

	TUI.projectsPathsView.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		TUI.projectsPathsFiltered[secondaryText] = !TUI.projectsPathsFiltered[secondaryText]
		if TUI.projectsPathsFiltered[secondaryText] {
			TUI.projectsPathsView.SetItemText(index, "[blue::b]"+mainText, secondaryText)
		} else {
			TUI.projectsPathsView.SetItemText(index, secondaryText, secondaryText)
		}

		filterProjects(config)
	})

	TUI.projectsPathsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		numItems := TUI.projectsPathsView.GetItemCount()
		currentItem := TUI.projectsPathsView.GetCurrentItem()

		switch event.Rune() {
		case 'g':
			TUI.projectsPathsView.SetCurrentItem(0)
			return nil
		case 'G':
			TUI.projectsPathsView.SetCurrentItem(numItems - 1)
			return nil
		case 'j':
			nextItem := currentItem + 1
			if nextItem < numItems {
				TUI.projectsPathsView.SetCurrentItem(nextItem)
			}
			return nil
		case 'k':
			nextItem := currentItem - 1
			if nextItem >= 0 {
				TUI.projectsPathsView.SetCurrentItem(nextItem)
			}
			return nil
		}

		return event
	})
}

func filterProjects(config *dao.Config) {
	projectPaths := []string{}
	projectTags := []string{}

	for key, value := range TUI.projectsPathsFiltered {
		if value {
			projectPaths = append(projectPaths, key)
		}
	}

	for key, value := range TUI.projectsTagsFiltered {
		if value {
			projectTags = append(projectTags, key)
		}
	}

	if len(projectPaths) > 0 || len(projectTags) > 0 {
		projects, _ := config.FilterProjects(false, false, []string{}, projectPaths, projectTags)
		TUI.projectsFiltered = projects
	} else {
		TUI.projectsFiltered = TUI.projectsAll
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
