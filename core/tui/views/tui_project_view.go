package views

import (
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
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
	ProjectHeaders   []string
	ShowHeaders      bool

	// Misc
	ProjectsTagsFiltered  map[string]bool
	ProjectsPathsFiltered map[string]bool
	Emitter               *misc.EventEmitter
}

func CreateProjectsData(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
	headers []string,
	showHeaders bool,
) TUIProjects {
	data := TUIProjects{
		Projects:         projects,
		ProjectTags:      projectTags,
		ProjectPaths:     projectPaths,
		ProjectsFiltered: projects,
		ProjectsSelected: []dao.Project{},

		ProjectsPathsFiltered: make(map[string]bool),
		ProjectsTagsFiltered:  make(map[string]bool),
		ProjectHeaders:        headers,
		ShowHeaders:           showHeaders,

		Emitter: misc.NewEventEmitter(),
	}

	for _, projectPath := range data.ProjectPaths {
		data.ProjectsPathsFiltered[projectPath] = false
	}
	for _, tag := range data.ProjectTags {
		data.ProjectsTagsFiltered[tag] = false
	}

	return data
}

func CreateProjectsTable(data *TUIProjects, selectEnabled bool, title string) components.TUITable {
	table := components.TUITable{SelectEnabled: selectEnabled, Title: title}
	table.CreateTable()

	// Methods
	table.IsRowSelected = func(name string) bool {
		return misc.IsProjectSelected(data.ProjectsSelected, name)
	}
	table.EditRow = func(projectName string) {
		editProject(projectName)
	}

	table.ToggleSelected = func() {
		i, _ := table.Table.GetSelection()
		projectName := table.Table.GetCell(i, 0).Text
		isSelected := misc.IsProjectSelected(data.ProjectsSelected, projectName)
		if isSelected {
			data.ProjectsSelected = misc.RemoveProject(data.ProjectsSelected, projectName)
		} else {
			project := misc.GetProject(data.Projects, projectName)
			data.ProjectsSelected = append(data.ProjectsSelected, project)
		}
		data.Emitter.Publish(misc.Event{Name: "toggle_selected_project", Data: projectName})
		table.UpdateCellStyles()
	}
	table.SelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			projectName := table.Table.GetCell(i, 0).Text
			if !misc.IsProjectSelected(data.ProjectsSelected, projectName) {
				project := misc.GetProject(data.Projects, projectName)
				data.ProjectsSelected = append(data.ProjectsSelected, project)
			}
		}
		data.Emitter.Publish(misc.Event{Name: "update_all_selected_projects", Data: ""})
		table.UpdateCellStyles()
	}
	table.DeSelectAllRows = func() {
		for i := 1; i < table.Table.GetRowCount(); i++ {
			projectName := table.Table.GetCell(i, 0).Text
			data.ProjectsSelected = misc.RemoveProject(data.ProjectsSelected, projectName)
		}
		data.Emitter.Publish(misc.Event{Name: "update_all_selected_projects", Data: ""})
		table.UpdateCellStyles()
	}
	table.DescribeRow = func() {
		row, _ := table.Table.GetSelection()
		if row > 0 {
			showProjectDescModal(data.Projects[row-1])
		}
	}

	// Events
	data.Emitter.Subscribe("filter_projects", func(e misc.Event) {
		filterProjects(&table, data)
	})
	data.Emitter.Subscribe("remove_selected_projects", func(e misc.Event) {
		updateProjectTable(&table, data)
	})
	data.Emitter.Subscribe("select_all_projects", func(e misc.Event) {
		table.SelectAllRows()
	})
	data.Emitter.Subscribe("deselect_all_projects", func(e misc.Event) {
		table.DeSelectAllRows()
	})

	updateProjectTable(&table, data)

	return table
}

func CreateProjectsTagsList(data *TUIProjects) components.TUIList {
	list := components.TUIList{Title: "Tags", Items: data.ProjectsTagsFiltered}
	list.CreateList()
	data.ProjectsTagsPane = list.List

	// Methods
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		list.HandleSelectItem(i, mainText, secondaryText)
		data.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
	}

	// Events
	data.Emitter.Subscribe("clear_filters", func(e misc.Event) {
		list.ClearItems(data.ProjectsTagsFiltered)
	})

	return list
}

func CreateProjectsPathsList(data *TUIProjects) components.TUIList {
	list := components.TUIList{Title: "Paths", Items: data.ProjectsPathsFiltered}
	list.CreateList()
	data.ProjectsPathsPane = list.List

	// Methods
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		list.HandleSelectItem(i, mainText, secondaryText)
		data.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
	}

	// Events
	data.Emitter.Subscribe("clear_filters", func(e misc.Event) {
		list.ClearItems(data.ProjectsPathsFiltered)
	})

	return list
}

func CreateProjectsSelectedList(data *TUIProjects, title string) components.TUIList {
	list := components.TUIList{Title: title, Items: make(map[string]bool)}
	list.CreateList()
	data.ProjectsSelectedPane = list.List

	// Methods
	updateSelectedProjects := func() {
		list.List.Clear()
		for _, project := range data.ProjectsSelected {
			list.List.AddItem(project.Name, project.Name, 0, nil)
		}

		if list.List.HasFocus() {
			list.SetActive(true)
		} else {
			list.SetActive(false)
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
			list.SetActive(true)
		} else {
			list.SetActive(false)
		}
	}
	list.SelectItem = func(i int, mainText string, secondaryText string) {
		projectName, _ := list.List.GetItemText(i)
		data.ProjectsSelected = misc.RemoveProject(data.ProjectsSelected, projectName)
		toggleSelectedProject(projectName)

		data.Emitter.Publish(misc.Event{Name: "remove_selected_projects", Data: ""})
	}

	// Events
	data.Emitter.Subscribe("toggle_selected_project", func(e misc.Event) {
		toggleSelectedProject(e.Data.(string))
	})

	data.Emitter.Subscribe("update_all_selected_projects", func(e misc.Event) {
		updateSelectedProjects()
	})

	return list
}

func updateProjectTable(t *components.TUITable, data *TUIProjects) {
	t.Table.Clear()

	// Set up headers
	for col, header := range data.ProjectHeaders {
		if data.ShowHeaders {
			t.Table.SetCell(0, col, components.CreateTableHeader(header))
		} else {
			t.Table.SetCell(0, col, components.CreateTableHeader(""))
		}
	}

	// Populate the table with project data
	for row, project := range data.ProjectsFiltered {
		for col, header := range data.ProjectHeaders {
			t.Table.SetCell(row+1, col, tview.NewTableCell(project.GetValue(header, 0)))
		}
	}

	t.UpdateCellStyles()
}

func filterProjects(t *components.TUITable, data *TUIProjects) {
	projectTags := []string{}
	for key, filtered := range data.ProjectsTagsFiltered {
		if filtered {
			projectTags = append(projectTags, key)
		}
	}

	projectPaths := []string{}
	for key, filtered := range data.ProjectsPathsFiltered {
		if filtered {
			projectPaths = append(projectPaths, key)
		}
	}

	if len(projectPaths) > 0 || len(projectTags) > 0 {
		projects, _ := misc.Config.FilterProjects(false, false, []string{}, projectPaths, projectTags)
		data.ProjectsFiltered = projects
	} else {
		data.ProjectsFiltered = data.Projects
	}

	updateProjectTable(t, data)
	t.Table.ScrollToBeginning()
	t.Table.Select(1, 0)
}

func showProjectDescModal(project dao.Project) {
	description := print.PrintProjectBlocks([]dao.Project{project})
	components.OpenTextModal("project-description-modal", description, project.Name, 80, 30)
}

func editProject(projectName string) {
	misc.App.Suspend(func() {
		err := misc.Config.EditProject(projectName)
		if err != nil {
			return
		}
	})
}
