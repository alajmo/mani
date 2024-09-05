package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func setupProjectPage(config *dao.Config, projects []dao.Project) *tview.Flex {
	// Poulate project data
	TUI.projectsFiltered = projects
	TUI.projectsTable = createProjectTable(projects)
	TUI.previousPage = TUI.projectsTable

	// Project tags
	TUI.projectsTagsView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
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
		SetMainTextColor(tcell.ColorWhite)
	TUI.projectsPathsView.SetTitle("Paths").SetBorder(true)
	populatePathsList(config)

	// Selected projects
	TUI.projectsSelectedView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		SetMainTextColor(tcell.ColorWhite)
	TUI.projectsSelectedView.SetTitle("Selected").SetBorder(true)
	populateSelectedList(config)

	// Projects context
	TUI.projectsContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.projectsTagsView, 0, 1, true).
		AddItem(TUI.projectsPathsView, 0, 1, true).
		AddItem(TUI.projectsSelectedView, 0, 1, true)

	// Project search
	TUI.inputSearch = tview.NewInputField().
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
		AddItem(TUI.inputSearch, 1, 0, false)

	// Callbacks
	TUI.projectsTable.SetFocusFunc(func() {
		TUI.projectsTable.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsTable.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsTable
		TUI.projectsTable.SetBorderColor(tcell.ColorWhite)
	})

	TUI.projectsTagsView.SetFocusFunc(func() {
		TUI.projectsTagsView.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsTagsView.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsTagsView
		TUI.projectsTagsView.SetBorderColor(tcell.ColorWhite)
	})

	TUI.projectsPathsView.SetFocusFunc(func() {
		TUI.projectsPathsView.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsPathsView.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsPathsView
		TUI.projectsPathsView.SetBorderColor(tcell.ColorWhite)
	})

	TUI.projectsSelectedView.SetFocusFunc(func() {
		TUI.projectsSelectedView.SetBorderColor(tcell.ColorYellow)
	})
	TUI.projectsSelectedView.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsSelectedView
		TUI.projectsSelectedView.SetBorderColor(tcell.ColorWhite)
	})

	return projectsPage
}

func createProjectTable(projects []dao.Project) *tview.Table {
	table := tview.NewTable()
	table.SetFixed(1, 0)           // Fixed header
	table.SetEvaluateAllRows(true) // Avoid resizing of headers when scrolling
	table.SetBorder(true).SetBorderPadding(0, 0, 2, 2)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorDefault)

	// Add headers + rows
	updateProjectTable(table)

	// Select first row
	table.Select(1, 0)

	selectedRows := make(map[int]bool)

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
		updateCellStyles(table, selectedRows)
	}

	// Event Listeners
	table.SetSelectionChangedFunc(func(row, column int) {
		updateCellStyles(table, selectedRows)
	})
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ': // Space: Toggle select project
				row, _ := table.GetSelection()
				if row > 0 {
					selectedRows[row] = !selectedRows[row]
				}

				updateSelectedProjectsDisplay(selectedRows)
				return nil
			case 'd': // Open up project description modal
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

	updateCellStyles(table, selectedRows)

	return table
}

func updateProjectTable(table *tview.Table) {
	table.Clear()

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
	for row, project := range TUI.projectsFiltered {
		table.SetCell(row+1, 0, tview.NewTableCell(project.Name))
		table.SetCell(row+1, 1, tview.NewTableCell(project.Desc))
		tagsString := ""
		if len(project.Tags) > 0 {
			tagsString = strings.Join(project.Tags, ", ")
		}
		table.SetCell(row+1, 2, tview.NewTableCell(tagsString))
	}
}

func updateCellStyles(table *tview.Table, selectedRows map[int]bool) {
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
			table.GetCell(row, col).SetStyle(style)
		}
	}
}

func updateSelectedProjectsDisplay(selectedRows map[int]bool) {
	if TUI.projectsTable == nil {
		return
	}

	for row, project := range TUI.projectsFiltered {
		exists := TUI.projectsSelectedView.FindItems(project.Name, project.Name, true, false)
		if selectedRows[row] && len(exists) == 0 {
			TUI.projectsSelectedView.AddItem(project.Name, project.Name, 0, nil)
		}

		if !selectedRows[row] && len(exists) > 0 {
			TUI.projectsSelectedView.RemoveItem(row)
		}
	}
}

func populateTagList(config *dao.Config) {
	for _, project := range TUI.projectsAll {
		for _, tag := range project.Tags {
			TUI.projectsTagsFiltered[tag] = false
		}
	}

	TUI.projectsTagsView.Clear()

	var tags []string
	for tag := range TUI.projectsTagsFiltered {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	for _, tag := range tags {
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

	var paths []string
	for path := range TUI.projectsPathsFiltered {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		TUI.projectsPathsView.AddItem(path, path, 0, nil)
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

func populateSelectedList(config *dao.Config) {
	TUI.projectsSelectedView.Clear()

	TUI.projectsSelectedView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		numItems := TUI.projectsSelectedView.GetItemCount()
		currentItem := TUI.projectsSelectedView.GetCurrentItem()

		switch event.Rune() {
		case 'g':
			TUI.projectsSelectedView.SetCurrentItem(0)
			return nil
		case 'G':
			TUI.projectsSelectedView.SetCurrentItem(numItems - 1)
			return nil
		case ' ':
			TUI.projectsSelectedView.SetCurrentItem(numItems - 1)
			return nil
		case 'j':
			nextItem := currentItem + 1
			if nextItem < numItems {
				TUI.projectsSelectedView.SetCurrentItem(nextItem)
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

	updateProjectTable(TUI.projectsTable)
}

func searchInTable(table *tview.Table, query string, lastFoundRow, lastFoundCol *int, direction int) {
	rowCount := table.GetRowCount()
	colCount := table.GetColumnCount()
	startRow := *lastFoundRow

	if startRow == -1 {
		startRow = 0
	} else {
		startRow += direction
	}

	searchRow := startRow
	for i := 0; i < rowCount; i++ {
		if searchRow < 0 {
			searchRow = rowCount - 1
		} else if searchRow >= rowCount {
			searchRow = 0
		}

		for col := 0; col < colCount; col++ {
			if cell := table.GetCell(searchRow, col); cell != nil {
				if strings.Contains(strings.ToLower(cell.Text), strings.ToLower(query)) {
					table.Select(searchRow, col)
					*lastFoundRow, *lastFoundCol = searchRow, col
					return
				}
			}
		}

		searchRow += direction
	}

	*lastFoundRow, *lastFoundCol = -1, -1
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
