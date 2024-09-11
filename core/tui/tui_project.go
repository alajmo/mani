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
	table := createProjectsTable(TUI.projects)

	// Project tags
	table.createProjectsTagsList()
	table.createProjectsPathsList()
	table.createProjectsSelectedList()

	// Projects context
	TUI.projectsContextPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(TUI.projectsTagsPane, 0, 1, true).
		AddItem(TUI.projectsPathsPane, 0, 1, true).
		AddItem(TUI.projectsSelectedPane, 0, 1, true)

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
}

func createProjectsTable(projects []dao.Project) TUITable {
	table := TUITable{}

	table.IsRowSelected = func(name string) bool {
		return isProjectSelected(TUI.projectsSelected, name)
	}
	table.ToggleSelected = func() {
		i, _ := table.Table.GetSelection()
		projectName := table.Table.GetCell(i, 0).Text
		isSelected := isProjectSelected(TUI.projectsSelected, projectName)
		if isSelected {
			TUI.projectsSelected = removeProject(TUI.projectsSelected, projectName)
		} else {
			project := getProject(TUI.projects, projectName)
			TUI.projectsSelected = append(TUI.projectsSelected, project)
		}
		TUI.emitter.Publish(Event{Name: "update_selected_projects", Data: "Alice"})
	}
	table.ToggleAllRows = func() {
		allSelected := isAllSelected(table.Table)
		if allSelected {
			// De-select all
			for i := 1; i < table.Table.GetRowCount(); i++ {
				projectName := table.Table.GetCell(i, 0).Text
				TUI.projectsSelected = removeProject(TUI.projectsSelected, projectName)
			}
		} else {
			// Select all
			for i := 1; i < table.Table.GetRowCount(); i++ {
				projectName := table.Table.GetCell(i, 0).Text
				if !isProjectSelected(TUI.projectsSelected, projectName) {
					project := getProject(TUI.projects, projectName)
					TUI.projectsSelected = append(TUI.projectsSelected, project)
				}
			}
		}
		TUI.emitter.Publish(Event{Name: "update_selected_projects", Data: "Alice"})
	}
	table.DescribeRow = func() {
		row, _ := table.Table.GetSelection()
		if row > 0 {
			openProjectDescModal(projects[row-1])
		}
	}

	table.createTable()
	table.updateProjectTable()

	TUI.projectsTable = table.Table
	TUI.previousPage = TUI.projectsTable

	TUI.emitter.Subscribe("remove_selected_projects", func(e Event) {
		table.updateProjectTable()
		table.Table.ScrollToBeginning()
		table.Table.Select(1, 0)
	})

	table.Table.SetFocusFunc(func() {
		TUI.projectsTable.SetBorderColor(tcell.ColorYellow)
	})
	table.Table.SetBlurFunc(func() {
		TUI.previousPage = TUI.projectsTable
		TUI.projectsTable.SetBorderColor(tcell.ColorWhite)
	})

	return table
}

func (t *TUITable) createProjectsTagsList() TUIList {
	numTags := len(TUI.projectTags)
	title := "Tags"
	if numTags > 0 {
		title = fmt.Sprintf("Tags (%d)", numTags)
	}

	list := TUIList{}
	list.createList(title)
	list.OnFocus = func() {
		setActive(list.List.Box, title, true)
	}
	list.OnBlur = func() {
		setActive(list.List.Box, title, false)
	}
	list.SelectItem = func() {
	}

	TUI.projectsTagsPane = list.List
	for _, tag := range TUI.projectTags {
		TUI.projectsTagsFiltered[tag] = false
	}

	var tags []string
	for tag := range TUI.projectsTagsFiltered {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	for _, tag := range tags {
		TUI.projectsTagsPane.AddItem(tag, tag, 0, nil)
	}

  // TUI.projectsTagsPane.Clear()
	TUI.projectsTagsPane.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		TUI.projectsTagsFiltered[secondaryText] = !TUI.projectsTagsFiltered[secondaryText]
		// If selected
		if TUI.projectsTagsFiltered[secondaryText] {
			TUI.projectsTagsPane.SetItemText(index, "[blue::b]"+mainText, secondaryText)
		} else {
			TUI.projectsTagsPane.SetItemText(index, secondaryText, secondaryText)
		}

		t.filterProjects()
	})

	return list
}

func (t *TUITable) createProjectsPathsList() TUIList {
	numPaths := len(TUI.projectPaths)
	title := "Paths"
	if numPaths > 0 {
		title = fmt.Sprintf("Paths (%d)", numPaths)
	}

	list := TUIList{}
	list.createList(title)
	list.OnFocus = func() {
		setActive(list.List.Box, title, true)
	}
	list.OnBlur = func() {
		setActive(list.List.Box, title, false)
	}
	list.SelectItem = func() {}

	TUI.projectsPathsPane = list.List
	for _, projectPath := range TUI.projectPaths {
		TUI.projectsPathsFiltered[projectPath] = false
	}

	TUI.projectsPathsPane.Clear()

	var paths []string
	for path := range TUI.projectsPathsFiltered {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		TUI.projectsPathsPane.AddItem(path, path, 0, nil)
	}

	TUI.projectsPathsPane.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		TUI.projectsPathsFiltered[secondaryText] = !TUI.projectsPathsFiltered[secondaryText]
		if TUI.projectsPathsFiltered[secondaryText] {
			TUI.projectsPathsPane.SetItemText(index, "[blue::b]"+mainText, secondaryText)
		} else {
			TUI.projectsPathsPane.SetItemText(index, secondaryText, secondaryText)
		}

		t.filterProjects()
	})

	return list
}

func (t *TUITable) createProjectsSelectedList() TUIList {
	list := TUIList{}
	list.createList("Selected")

	updateSelectedProjects := func() {
		list.List.Clear()
		for _, project := range TUI.projectsSelected {
			list.List.AddItem(project.Name, project.Name, 0, nil)
		}

		numPaths := len(TUI.projectsSelected)
		title := "Selected"
		if numPaths > 0 {
			title = fmt.Sprintf("Selected (%d)", numPaths)
		}

		if list.List.HasFocus() {
			setActive(list.List.Box, title, true)
		} else {
			setActive(list.List.Box, title, false)
		}
	}

	list.OnFocus = func() {
		numPaths := len(TUI.projectsSelected)
		title := "Selected"
		if numPaths > 0 {
			title = fmt.Sprintf("Selected (%d)", numPaths)
		}

		setActive(list.List.Box, title, true)
	}
	list.OnBlur = func() {
		numPaths := len(TUI.projectsSelected)
		title := "Selected"
		if numPaths > 0 {
			title = fmt.Sprintf("Selected (%d)", numPaths)
		}
		setActive(list.List.Box, title, false)
	}
	list.SelectItem = func() {
		currentItem := list.List.GetCurrentItem()
		projectName, _ := list.List.GetItemText(currentItem)

		TUI.projectsSelected = removeProject(TUI.projectsSelected, projectName)
		TUI.emitter.Publish(Event{Name: "remove_selected_projects", Data: "Alice"})
		updateSelectedProjects()
	}

	TUI.projectsSelectedPane = list.List
	TUI.emitter.Subscribe("update_selected_projects", func(e Event) {
		updateSelectedProjects()
	})

	return list
}

func (t *TUITable) updateProjectTable() {
	t.Table.Clear()

	// Set up headers
	headers := []string{"Name", "Description", "Tags"}
	for col, header := range headers {
		t.Table.SetCell(0, col,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetAttributes(tcell.AttrBold).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Populate the table with project data
	for row, project := range TUI.projectsFiltered {
		t.Table.SetCell(row+1, 0, tview.NewTableCell(project.Name))
		t.Table.SetCell(row+1, 1, tview.NewTableCell(project.Desc))
		tagsString := ""
		if len(project.Tags) > 0 {
			tagsString = strings.Join(project.Tags, ", ")
		}
		t.Table.SetCell(row+1, 2, tview.NewTableCell(tagsString))
	}

	t.updateCellStyles()
}

func (t *TUITable) populateTagsList() {
	for _, tag := range TUI.projectTags {
		TUI.projectsTagsFiltered[tag] = false
	}

	TUI.projectsTagsPane.Clear()

	var tags []string
	for tag := range TUI.projectsTagsFiltered {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	for _, tag := range tags {
		TUI.projectsTagsPane.AddItem(tag, tag, 0, nil)
	}

	TUI.projectsTagsPane.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		TUI.projectsTagsFiltered[secondaryText] = !TUI.projectsTagsFiltered[secondaryText]
		if TUI.projectsTagsFiltered[secondaryText] {
			TUI.projectsTagsPane.SetItemText(index, "[blue::b]"+mainText, secondaryText)
		} else {
			TUI.projectsTagsPane.SetItemText(index, secondaryText, secondaryText)
		}

		t.filterProjects()
	})
}

func (t *TUITable) filterProjects() {
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

	t.updateProjectTable()
	TUI.projectsTable.ScrollToBeginning()
	TUI.projectsTable.Select(1, 0)
}

func getProject(projects []dao.Project, projectName string) dao.Project {
	for index, project := range projects {
		if project.Name == projectName {
			return projects[index]
		}
	}
	return dao.Project{}
}

func removeProject(projects []dao.Project, projectName string) []dao.Project {
	for index, project := range projects {
		if project.Name == projectName {
			return append(projects[:index], projects[index+1:]...)
		}
	}
	return projects
}

func isProjectSelected(projects []dao.Project, projectName string) bool {
	for _, project := range projects {
		if project.Name == projectName {
			return true
		}
	}
	return false
}

func isAllSelected(table *tview.Table) bool {
	for i := 1; i < table.GetRowCount(); i++ {
		projectName := table.GetCell(i, 0).Text
		if !isProjectSelected(TUI.projectsSelected, projectName) {
			return false
		}
	}
	return true
}

func openProjectDescModal(project dao.Project) {
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
