package views

import (
	"strings"

	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

type TUIProjects struct {
	// UI
	ProjectsPage         *tview.Flex
	ProjectsTable        *tview.Table
	ProjectsContextPage  *tview.Flex
	ProjectsTagsPane     *tview.List
	ProjectsPathsPane    *tview.List
	ProjectsSelectedPane *tview.List

	// Data
	Projects         []dao.Project
	ProjectsFiltered []dao.Project
	ProjectsSelected []dao.Project
	ProjectTags      []string
	ProjectPaths     []string

	// Misc
	ProjectsTagsFiltered  map[string]bool
	ProjectsPathsFiltered map[string]bool
}

func CreateProjectsData(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) TUIProjects {
	data := TUIProjects{
		projects:         projects,
		projectTags:      projectTags,
		projectPaths:     projectPaths,
		projectsFiltered: projects,
		projectsSelected: []dao.Project{},

		projectsPathsFiltered: make(map[string]bool),
		projectsTagsFiltered:  make(map[string]bool),
	}

	for _, projectPath := range data.projectPaths {
		data.projectsPathsFiltered[projectPath] = false
	}
	for _, tag := range data.projectTags {
		data.projectsTagsFiltered[tag] = false
	}

	return data
}

func CreateProjectsTable(data *TUIProjects) TUITable {
	table := TUITable{}
	table.createTable()
	// TUI.projectsTable = table.Table
	// TUI.previousPage = table.Table

	// Methods
	table.IsRowSelected = func(name string) bool {
		return isProjectSelected(data.projectsSelected, name)
	}
	table.EditRow = func(projectName string) {
		editProject(projectName)
	}
	table.ToggleSelected = func() {
		i, _ := table.Table.GetSelection()
		projectName := table.Table.GetCell(i, 0).Text
		isSelected := isProjectSelected(data.projectsSelected, projectName)
		if isSelected {
			data.projectsSelected = removeProject(data.projectsSelected, projectName)
		} else {
			project := getProject(data.projects, projectName)
			data.projectsSelected = append(data.projectsSelected, project)
		}
		TUI.emitter.Publish(Event{Name: "toggle_selected_project", Data: projectName})
		table.updateCellStyles()
	}
	table.SelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			projectName := table.Table.GetCell(i, 0).Text
			if !isProjectSelected(data.projectsSelected, projectName) {
				project := getProject(data.projects, projectName)
				data.projectsSelected = append(data.projectsSelected, project)
			}
		}
		TUI.emitter.Publish(Event{Name: "update_all_selected_projects", Data: ""})
		table.updateCellStyles()
	}
	table.DeSelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			projectName := table.Table.GetCell(i, 0).Text
			data.projectsSelected = removeProject(data.projectsSelected, projectName)
		}
		TUI.emitter.Publish(Event{Name: "update_all_selected_projects", Data: ""})
		table.updateCellStyles()
	}
	table.DescribeRow = func() {
		row, _ := table.Table.GetSelection()
		if row > 0 {
			showProjectDescModal(data.projects[row-1])
		}
	}

	// Events
	TUI.emitter.Subscribe("filter_projects", func(e Event) {
		table.filterProjects(data)
	})
	TUI.emitter.Subscribe("remove_selected_projects", func(e Event) {
		table.updateProjectTable(data)
	})
	TUI.emitter.Subscribe("select_all_projects", func(e Event) {
		table.SelectAllRows()
	})
	TUI.emitter.Subscribe("deselect_all_projects", func(e Event) {
		table.DeSelectAllRows()
	})

	table.updateProjectTable(data)

	return table
}

func CreateProjectsTagsList(data *TUIProjects) TUIList {
	list := TUIList{Title: "Tags", Items: data.projectsTagsFiltered}
	list.createList()
	data.projectsTagsPane = list.List

	// Methods
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		list.handleSelectItem(i, mainText, secondaryText)
		TUI.emitter.Publish(Event{Name: "filter_projects", Data: ""})
	}

	// Events
	TUI.emitter.Subscribe("clear_filters", func(e Event) {
		list.clearItems(data.projectsTagsFiltered)
	})

	return list
}

func CreateProjectsPathsList(data *TUIProjects) TUIList {
	list := TUIList{Title: "Paths", Items: data.projectsPathsFiltered}
	list.createList()
	data.projectsPathsPane = list.List

	// Methods
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		list.handleSelectItem(i, mainText, secondaryText)
		TUI.emitter.Publish(Event{Name: "filter_projects", Data: ""})
	}

	// Events
	TUI.emitter.Subscribe("clear_filters", func(e Event) {
		list.clearItems(data.projectsPathsFiltered)
	})

	return list
}

func CreateProjectsSelectedList(data *TUIProjects) TUIList {
	list := TUIList{Title: "Selected", Items: make(map[string]bool)}
	list.createList()
	data.projectsSelectedPane = list.List

	// Methods
	updateSelectedProjects := func() {
		list.List.Clear()
		for _, project := range data.projectsSelected {
			list.List.AddItem(project.Name, project.Name, 0, nil)
		}

		if list.List.HasFocus() {
			list.setActive(true)
		} else {
			list.setActive(false)
		}
	}
	toggleSelectedProject := func(projectName string) {
		items := list.List.FindItems(projectName, projectName, false, false)
		if len(items) == 0 {
			list.List.AddItem(projectName, projectName, 0, nil)
		} else {
			list.List.RemoveItem(items[0])
		}

		if list.List.HasFocus() {
			list.setActive(true)
		} else {
			list.setActive(false)
		}
	}
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		projectName, _ := list.List.GetItemText(i)
		data.projectsSelected = removeProject(data.projectsSelected, projectName)
		toggleSelectedProject(projectName)

		TUI.emitter.Publish(Event{Name: "remove_selected_projects", Data: ""})
	}

	// Events
	TUI.emitter.Subscribe("toggle_selected_project", func(e Event) {
		toggleSelectedProject(e.Data.(string))
	})

	TUI.emitter.Subscribe("update_all_selected_projects", func(e Event) {
		updateSelectedProjects()
	})

	return list
}

func (t *TUITable) updateProjectTable(data *TUIProjects) {
	t.Table.Clear()

	// Set up headers
	headers := []string{"Name", "Description", "Tags"}
	for col, header := range headers {
		t.Table.SetCell(0, col, createTableHeader(header))
	}

	// Populate the table with project data
	for row, project := range data.projectsFiltered {
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

func (t *TUITable) filterProjects(data *TUIProjects) {
	projectTags := []string{}
	for key, filtered := range data.projectsTagsFiltered {
		if filtered {
			projectTags = append(projectTags, key)
		}
	}

	projectPaths := []string{}
	for key, filtered := range data.projectsPathsFiltered {
		if filtered {
			projectPaths = append(projectPaths, key)
		}
	}

	if len(projectPaths) > 0 || len(projectTags) > 0 {
		projects, _ := TUI.config.FilterProjects(false, false, []string{}, projectPaths, projectTags)
		data.projectsFiltered = projects
	} else {
		data.projectsFiltered = data.projects
	}

	t.updateProjectTable(data)
	t.Table.ScrollToBeginning()
	t.Table.Select(1, 0)
}

func showProjectDescModal(project dao.Project) {
	description := print.PrintProjectBlocks([]dao.Project{project})
	openModal("project-description-modal", description, project.Name, 80, 30)
}

func editProject(projectName string) {
	TUI.app.Suspend(func() {
		err := TUI.config.EditProject(projectName)
		if err != nil {
			return
		}
	})
}
