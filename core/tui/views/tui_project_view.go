package views

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
)

type TProject struct {
	// UI
	Page             *tview.Flex
	ContextView      *tview.Flex
	ProjectTableView *components.TTable
	ProjectTreeView  *components.TTree
	TagView          *components.TList
	PathView         *components.TList

	// Project
	Projects           []dao.Project
	ProjectsFiltered   []dao.Project
	ProjectsSelected   map[string]bool
	projectFilterValue *string
	Headers            []string
	ShowHeaders        bool
	ProjectStyle       string

	// Tags
	ProjectTags           []string
	ProjectTagsFiltered   []string
	ProjectTagsSelected   map[string]bool
	projectTagFilterValue *string

	// Paths
	ProjectPaths           []string
	ProjectPathsFiltered   []string
	ProjectPathsSelected   map[string]bool
	projectPathFilterValue *string

	// Misc
	Emitter *misc.EventEmitter
}

func CreateProjectsData(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
	headers []string,
	prefixNumber int,
	showTitle bool,
	showHeaders bool,
	selectEnabled bool,
	showTags bool,
	showPaths bool,
) *TProject {
	p := &TProject{
		Projects:           projects,
		ProjectsFiltered:   projects,
		ProjectsSelected:   make(map[string]bool),
		projectFilterValue: new(string),

		ProjectTags:           projectTags,
		ProjectTagsFiltered:   projectTags,
		ProjectTagsSelected:   make(map[string]bool),
		projectTagFilterValue: new(string),

		ProjectPaths:           projectPaths,
		ProjectPathsFiltered:   projectPaths,
		ProjectPathsSelected:   make(map[string]bool),
		projectPathFilterValue: new(string),

		ProjectStyle: "project-table",
		ShowHeaders:  showHeaders,
		Headers:      headers,

		Emitter: misc.NewEventEmitter(),
	}

	for _, project := range p.Projects {
		p.ProjectsSelected[project.Name] = false
	}
	for _, tag := range p.ProjectTags {
		p.ProjectTagsSelected[tag] = false
	}
	for _, projectPath := range p.ProjectPaths {
		p.ProjectPathsSelected[projectPath] = false
	}

	title := ""
	if showTitle && prefixNumber > 0 {
		title = fmt.Sprintf("[%d] Projects (%d)", prefixNumber, len(projects))
		prefixNumber += 1
	} else if showTitle {
		title = fmt.Sprintf("Projects (%d)", len(projects))
	}

	rows := p.getTableRows()
	projectTable := p.CreateProjectsTable(selectEnabled, title, headers, rows)
	p.ProjectTableView = projectTable

	paths := p.getTreeHierarchy()
	projectTree := p.CreateProjectsTree(selectEnabled, title, paths)
	p.ProjectTreeView = projectTree

	if showTags {
		tagTitle := ""
		if showTitle && prefixNumber > 0 {
			tagTitle = fmt.Sprintf("[%d] Tags (%d)", prefixNumber, len(projectTags))
			prefixNumber += 1
		} else {
			tagTitle = fmt.Sprintf("Tags (%d)", len(projectTags))
		}

		tagsList := p.CreateProjectsTagsList(tagTitle)
		p.TagView = tagsList
	}

	if showPaths {
		pathTitle := ""
		if showTitle && prefixNumber > 0 {
			pathTitle = fmt.Sprintf("[%d] Paths (%d)", prefixNumber, len(projectPaths))
		} else {
			pathTitle = fmt.Sprintf("Paths (%d)", len(projectPaths))
		}

		pathsList := p.CreateProjectsPathsList(pathTitle)
		p.PathView = pathsList
	}

	// Events
	p.Emitter.Subscribe("remove_tag_path_filter", func(e misc.Event) {
		p.TagView.ClearFilter()
		p.PathView.ClearFilter()
	})
	p.Emitter.Subscribe("remove_tag_path_selections", func(e misc.Event) {
		p.unselectAllTags()
		p.unselectAllPaths()
	})
	p.Emitter.Subscribe("remove_project_filter", func(e misc.Event) {
		p.ProjectTableView.ClearFilter()
		p.ProjectTreeView.ClearFilter()
	})
	p.Emitter.Subscribe("remove_project_selections", func(event misc.Event) {
		p.unselectAllProjects()
	})
	p.Emitter.Subscribe("filter_projects", func(e misc.Event) {
		p.filterProjects()
	})

	return p
}

func (p *TProject) CreateProjectsTable(
	selectEnabled bool,
	title string,
	headers []string,
	rows [][]string,
) *components.TTable {
	table := &components.TTable{
		Title:         title,
		ToggleEnabled: selectEnabled,
		ShowHeaders:   p.ShowHeaders,
		FilterValue:   p.projectFilterValue,
	}
	table.Create()
	table.Update(headers, rows)

	// Methods
	table.IsRowSelected = func(name string) bool {
		return p.ProjectsSelected[name]
	}
	table.ToggleSelectRow = func(name string) {
		p.toggleSelectProject(name)
	}
	table.SelectAll = func() {
		p.selectAllProjects()
	}
	table.UnselectAll = func() {
		p.unselectAllProjects()
	}
	table.FilterRows = func() {
		p.filterProjects()
	}
	table.DescribeRow = func(projectName string) {
		if projectName != "" {
			p.showProjectDescModal(projectName)
		}
	}
	table.EditRow = func(projectName string) {
		if projectName != "" {
			p.editProject(projectName)
		}
	}
	return table
}

func (p *TProject) CreateProjectsTree(
	selectEnabled bool,
	title string,
	paths []dao.TNode,
) *components.TTree {
	tree := &components.TTree{
		Title:         title,
		RootTitle:     "",
		SelectEnabled: selectEnabled,
		FilterValue:   p.projectFilterValue,
	}
	tree.Create()
	tree.UpdateProjects(paths)

	tree.IsNodeSelected = func(name string) bool {
		return p.ProjectsSelected[name]
	}
	tree.ToggleSelectNode = func(name string) {
		p.toggleSelectProject(name)
	}
	tree.SelectAll = func() {
		p.selectAllProjects()
	}
	tree.UnselectAll = func() {
		p.unselectAllProjects()
	}
	tree.FilterNodes = func() {
		p.filterProjects()
	}
	tree.DescribeNode = func(projectName string) {
		if projectName != "" {
			p.showProjectDescModal(projectName)
		}
	}
	tree.EditNode = func(projectName string) {
		if projectName != "" {
			p.editProject(projectName)
		}
	}

	return tree
}

func (p *TProject) CreateProjectsTagsList(title string) *components.TList {
	list := &components.TList{
		Title:       title,
		FilterValue: p.projectTagFilterValue,
	}
	list.Create()
	list.Update(p.ProjectTags)

	// Methods
	list.IsItemSelected = func(name string) bool {
		return p.ProjectTagsSelected[name]
	}
	list.ToggleSelectItem = func(i int, tag string) {
		p.ProjectTagsSelected[tag] = !p.ProjectTagsSelected[tag]
		list.SetItemSelect(i, tag)
		p.filterProjects()
	}
	list.SelectAll = func() {
		p.selectAllTags()
		p.filterProjects()
	}
	list.UnselectAll = func() {
		p.unselectAllTags()
		p.filterProjects()
	}
	list.FilterItems = func() {
		p.filterTags()
	}

	return list
}

func (p *TProject) CreateProjectsPathsList(title string) *components.TList {
	list := &components.TList{
		Title:       title,
		FilterValue: p.projectPathFilterValue,
	}
	list.Create()
	list.Update(p.ProjectPaths)

	// Methods
	list.IsItemSelected = func(name string) bool {
		return p.ProjectPathsSelected[name]
	}
	list.ToggleSelectItem = func(i int, tag string) {
		p.ProjectPathsSelected[tag] = !p.ProjectPathsSelected[tag]
		list.SetItemSelect(i, tag)
		p.filterProjects()
	}
	list.SelectAll = func() {
		p.selectAllPaths()
		p.filterProjects()
	}
	list.UnselectAll = func() {
		p.unselectAllPaths()
		p.filterProjects()
	}
	list.FilterItems = func() {
		p.filterPaths()
	}

	return list
}

func (p *TProject) getTableRows() [][]string {
	var rows = make([][]string, len(p.ProjectsFiltered))
	for i, project := range p.ProjectsFiltered {
		rows[i] = make([]string, len(p.Headers))
		for j, header := range p.Headers {
			rows[i][j] = project.GetValue(header, 0)
		}
	}
	return rows
}

func (p *TProject) getTreeHierarchy() []dao.TNode {
	var paths = []dao.TNode{}
	for _, p := range p.ProjectsFiltered {
		node := dao.TNode{Name: p.Name, Path: p.RelPath}
		paths = append(paths, node)
	}

	return paths
}

func (p *TProject) toggleSelectProject(name string) {
	p.ProjectsSelected[name] = !p.ProjectsSelected[name]
	p.ProjectTableView.ToggleSelectCurrentRow(name)
	p.ProjectTreeView.ToggleSelectCurrentNode(name)
}

func (p *TProject) filterProjects() {
	projectTags := []string{}
	for key, filtered := range p.ProjectTagsSelected {
		if filtered {
			projectTags = append(projectTags, key)
		}
	}

	projectPaths := []string{}
	for key, filtered := range p.ProjectPathsSelected {
		if filtered {
			projectPaths = append(projectPaths, key)
		}
	}

	if len(projectTags) > 0 || len(projectPaths) > 0 {
		projects, _ := misc.Config.FilterProjects(false, false, []string{}, projectPaths, projectTags, "")
		p.ProjectsFiltered = projects
	} else {
		p.ProjectsFiltered = p.Projects
	}

	var finalProjects []dao.Project
	for _, project := range p.ProjectsFiltered {
		if strings.Contains(project.Name, *p.projectFilterValue) {
			finalProjects = append(finalProjects, project)
		}
	}
	p.ProjectsFiltered = finalProjects

	// Table
	rows := p.getTableRows()
	p.ProjectTableView.Update(p.Headers, rows)
	p.ProjectTableView.Table.ScrollToBeginning()
	p.ProjectTableView.Table.Select(1, 0)

	// Tree
	paths := p.getTreeHierarchy()
	p.ProjectTreeView.UpdateProjects(paths)
	p.ProjectTreeView.UpdateProjectsStyle()
	p.ProjectTreeView.FocusFirst()
}

func (p *TProject) filterTags() {
	var finalTags []string
	for _, tag := range p.ProjectTags {
		if strings.Contains(tag, *p.projectTagFilterValue) {
			finalTags = append(finalTags, tag)
		}
	}
	p.ProjectTagsFiltered = finalTags
	p.TagView.Update(p.ProjectTagsFiltered)
}

func (p *TProject) filterPaths() {
	var finalPaths []string
	for _, path := range p.ProjectPaths {
		if strings.Contains(path, *p.projectPathFilterValue) {
			finalPaths = append(finalPaths, path)
		}
	}
	p.ProjectPathsFiltered = finalPaths
	p.PathView.Update(p.ProjectPathsFiltered)
}

func (p *TProject) selectAllProjects() {
	for _, project := range p.ProjectsFiltered {
		p.ProjectsSelected[project.Name] = true
	}
	p.ProjectTableView.UpdateRowStyle()
	p.ProjectTreeView.UpdateProjectsStyle()
}

func (p *TProject) selectAllTags() {
	for _, tag := range p.ProjectTagsFiltered {
		p.ProjectTagsSelected[tag] = true
	}
	p.TagView.Update(p.ProjectTagsFiltered)
}

func (p *TProject) selectAllPaths() {
	for _, path := range p.ProjectPathsFiltered {
		p.ProjectPathsSelected[path] = true
	}
	p.PathView.Update(p.ProjectPathsFiltered)
}

func (p *TProject) unselectAllProjects() {
	for _, project := range p.ProjectsFiltered {
		p.ProjectsSelected[project.Name] = false
	}
	p.ProjectTableView.UpdateRowStyle()
	p.ProjectTreeView.UpdateProjectsStyle()
}

func (p *TProject) unselectAllTags() {
	for _, tag := range p.ProjectTagsFiltered {
		p.ProjectTagsSelected[tag] = false
	}
	p.TagView.Update(p.ProjectTagsFiltered)
}

func (p *TProject) unselectAllPaths() {
	for _, path := range p.ProjectPathsFiltered {
		p.ProjectPathsSelected[path] = false
	}
	p.PathView.Update(p.ProjectPathsFiltered)
}

func (p *TProject) showProjectDescModal(name string) {
	project, err := misc.Config.GetProject(name)
	if err != nil {
		return
	}
	description := print.PrintProjectBlocks([]dao.Project{*project}, true, *misc.BlockTheme, print.TviewFormatter{})
	descriptionNoColor := print.PrintProjectBlocks([]dao.Project{*project}, false, *misc.BlockTheme, print.TviewFormatter{})
	components.OpenTextModal("project-description-modal", description, descriptionNoColor, project.Name)
}

func (p *TProject) editProject(projectName string) {
	misc.App.Suspend(func() {
		err := misc.Config.EditProject(projectName)
		if err != nil {
			return
		}
	})
}
