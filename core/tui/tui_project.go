package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createProjectsPage() {
	// Poulate project data
	TUI.projectsFiltered = TUI.projects
	TUI.projectsTable = createProjectTable(TUI.projects)
	TUI.previousPage = TUI.projectsTable

	// Project tags
	TUI.projectsTagsView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		SetMainTextColor(tcell.ColorWhite)
	TUI.projectsTagsView.SetTitle("[::b] Tags ").SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	// TODO: Add number of tags

	populateTagList()

	// Project paths
	TUI.projectsPathsView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		SetMainTextColor(tcell.ColorWhite)
	TUI.projectsPathsView.SetTitle("[::b] Paths ").SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	populatePathsList()
	// TODO: Add number of tags
	// TODO: Add number of paths

	// Selected projects
	TUI.projectsSelectedView = tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		SetMainTextColor(tcell.ColorWhite)
	TUI.projectsSelectedView.SetTitle("[::b] Selected ").SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	populateSelectedList()
	// TODO: Add number of selected

	// Projects context
	TUI.projectsContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.projectsTagsView, 0, 1, true).
		AddItem(TUI.projectsPathsView, 0, 1, true).
		AddItem(TUI.projectsSelectedView, 0, 1, true)

	// Project search
	TUI.search = tview.NewInputField().
		SetLabel("").
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorBlue)

	TUI.projectsPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(TUI.projectsTable, 0, 1, true).
				AddItem(TUI.projectsContextPage, 30, 1, false),
			0, 1, true).
		AddItem(TUI.search, 1, 0, false)

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
		TUI.projectsTagsView.SetTitle("[yellow::b] Tags ")

	})
	TUI.projectsTagsView.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsTagsView
		TUI.projectsTagsView.SetBorderColor(tcell.ColorWhite)
		TUI.projectsTagsView.SetTitle("[white::b] Tags ")
	})

	TUI.projectsPathsView.SetFocusFunc(func() {
		TUI.projectsPathsView.SetBorderColor(tcell.ColorYellow)
		TUI.projectsPathsView.SetTitle("[yellow::b] Paths ")
	})
	TUI.projectsPathsView.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsPathsView
		TUI.projectsPathsView.SetBorderColor(tcell.ColorWhite)
		TUI.projectsPathsView.SetTitle("[white::b] Paths ")
	})

	TUI.projectsSelectedView.SetFocusFunc(func() {
		TUI.projectsSelectedView.SetBorderColor(tcell.ColorYellow)
		TUI.projectsSelectedView.SetTitle("[yellow::b] Selected ")
	})
	TUI.projectsSelectedView.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsSelectedView
		TUI.projectsSelectedView.SetBorderColor(tcell.ColorWhite)
		TUI.projectsSelectedView.SetTitle("[white::b] Selected ")
	})
}

func createProjectTable(projects []dao.Project) *tview.Table {
	table := createTable()

	// Add headers + rows
	updateProjectTable(table)

	// Callbacks
	isAllSelected := func() bool {
		for i := 1; i < table.GetRowCount(); i++ {
			projectName := table.GetCell(i, 0).Text
			if !isProjectSelected(TUI.projectsSelected, projectName) {
				return false
			}
		}
		return true
	}
	toggleAllRows := func() {
		allSelected := isAllSelected()
		if allSelected {
			// De-select all
			for i := 1; i < table.GetRowCount(); i++ {
				projectName := table.GetCell(i, 0).Text
				TUI.projectsSelected = removeProject(TUI.projectsSelected, projectName)
			}
		} else {
			// Select all
			for i := 1; i < table.GetRowCount(); i++ {
				projectName := table.GetCell(i, 0).Text
				if !isProjectSelected(TUI.projectsSelected, projectName) {
					project := getProject(TUI.projects, projectName)
					TUI.projectsSelected = append(TUI.projectsSelected, project)
				}
			}
		}

		updateSelectedProjectsDisplay()
		updateCellStyles(table)
	}

	// Event Listeners
	table.SetSelectionChangedFunc(func(row, column int) {
		updateCellStyles(table)
	})
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ': // space: Toggle select project
				i, _ := table.GetSelection()
				projectName := table.GetCell(i, 0).Text
				isSelected := isProjectSelected(TUI.projectsSelected, projectName)
				if isSelected {
					TUI.projectsSelected = removeProject(TUI.projectsSelected, projectName)
				} else {
					project := getProject(TUI.projects, projectName)
					TUI.projectsSelected = append(TUI.projectsSelected, project)
				}

				// TODOZ: Callback here to re-render
				updateSelectedProjectsDisplay()
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

	updateCellStyles(table)

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

	updateCellStyles(table)
}

func updateCellStyles(table *tview.Table) {
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
		isSelected := false
		projectName := table.GetCell(row, 0).Text
		if isProjectSelected(TUI.projectsSelected, projectName) {
			isSelected = true

		}

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

func updateSelectedProjectsDisplay() {
	TUI.projectsSelectedView.Clear()

	for _, project := range TUI.projectsSelected {
		TUI.projectsSelectedView.AddItem(project.Name, project.Name, 0, nil)
	}
}

func populateTagList() {
	for _, tag := range TUI.projectTags {
		TUI.projectsTagsFiltered[tag] = false
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

		filterProjects()
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

func populatePathsList() {
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

		filterProjects()
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

func populateSelectedList() {
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
			projectName, _ := TUI.projectsSelectedView.GetItemText(currentItem)
			TUI.projectsSelected = removeProject(TUI.projectsSelected, projectName)
			// TODOZ: Callback here to re-render
			updateSelectedProjectsDisplay()
			updateProjectTable(TUI.projectsTable)
			TUI.projectsTable.ScrollToBeginning()
			TUI.projectsTable.Select(1, 0)
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
				TUI.projectsSelectedView.SetCurrentItem(nextItem)
			}
			return nil
		}

		return event
	})
}

func filterProjects() {
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
		projects, _ := TUI.config.FilterProjects(false, false, []string{}, projectPaths, projectTags)
		TUI.projectsFiltered = projects
	} else {
		TUI.projectsFiltered = TUI.projects
	}

	updateProjectTable(TUI.projectsTable)
	TUI.projectsTable.ScrollToBeginning()
	TUI.projectsTable.Select(1, 0)
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
		description += printList("Tags: \n", project.Tags)
	}

	if len(project.EnvList) > 0 {
		description += printList("Env: \n", project.EnvList)
	}

	openModal("project-description", description, project.Name, 80)
}

func isProjectSelected(projects []dao.Project, projectName string) bool {
	for _, project := range projects {
		if project.Name == projectName {
			return true
		}
	}

	return false
}

func removeProject(projects []dao.Project, projectName string) []dao.Project {
	for index, project := range projects {
		if project.Name == projectName {
			return append(projects[:index], projects[index+1:]...)
		}
	}
	return projects
}

func getProject(projects []dao.Project, projectName string) dao.Project {
	for index, project := range projects {
		if project.Name == projectName {
			return projects[index]
		}
	}

	return dao.Project{}
}
