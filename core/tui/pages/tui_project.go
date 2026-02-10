package pages

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TProjectPage struct {
	focusable []*misc.TItem
}

func CreateProjectsPage(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	p := &TProjectPage{}

	// Data
	projectData := views.CreateProjectsData(
		projects,
		projectTags,
		projectPaths,
		[]string{"Project", "Description", "Tag", "Url", "Path"},
		1,
		true,
		true,
		false,
		true,
		true,
	)

	// Views
	projectInfo := views.CreateProjectInfoView()
	projectTablePage := p.createProjectPage(projectData)

	// Context page (always show both panes, even when empty)
	projectData.ContextView = tview.NewFlex().SetDirection(tview.FlexRow)
	projectData.ContextView.AddItem(projectData.TagView.Root, 0, 1, true)
	projectData.ContextView.AddItem(projectData.PathView.Root, 0, 1, true)

	// Page
	projectData.Page = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(projectTablePage, 0, 1, true).
				AddItem(projectData.ContextView, 30, 1, false),
			0, 1, true).
		AddItem(projectInfo, 1, 0, false).
		AddItem(misc.Search, 1, 0, false)

	// Focusable
	p.focusable = p.updateProjectFocusable(projectData)
	misc.ProjectsLastFocus = &p.focusable[0].Primitive

	// Shortcuts
	projectData.Page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyTab:
			nextPrimitive := misc.FocusNext(p.focusable)
			misc.ProjectsLastFocus = nextPrimitive
			return nil
		case tcell.KeyBacktab:
			nextPrimitive := misc.FocusPrevious(p.focusable)
			misc.ProjectsLastFocus = nextPrimitive
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'C': // Clear filters
				projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_path_filter", Data: ""})
				projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_tag_path_selections", Data: ""})
				projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_project_filter", Data: ""})
				projectData.Emitter.PublishAndWait(misc.Event{Name: "remove_project_selections", Data: ""})
				projectData.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
				return nil
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				misc.FocusPage(event, p.focusable)
				return nil
			}
		}
		return event
	})

	return projectData.Page
}

func (p *TProjectPage) createProjectPage(projectData *views.TProject) *tview.Flex {
	isTable := projectData.ProjectStyle == "project-table"

	pages := tview.NewPages().
		AddPage("project-table", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(projectData.ProjectTableView.Root, 0, 1, true), true, isTable).
		AddPage("project-tree", tview.NewFlex().SetDirection(tview.FlexRow).AddItem(projectData.ProjectTreeView.Root, 0, 8, false), true, !isTable)

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyCtrlE:
			if projectData.ProjectStyle == "project-table" {
				projectData.ProjectStyle = "project-tree"
			} else {
				projectData.ProjectStyle = "project-table"
			}
			pages.SwitchToPage(projectData.ProjectStyle)
			p.focusable = p.updateProjectFocusable(projectData)
			misc.App.SetFocus(p.focusable[0].Primitive)
			misc.ProjectsLastFocus = &p.focusable[0].Primitive
			return nil

		}
		return event
	})

	return page
}

func (p *TProjectPage) updateProjectFocusable(
	data *views.TProject,
) []*misc.TItem {
	focusable := []*misc.TItem{}

	if data.ProjectStyle == "project-table" {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				data.ProjectTableView.Table,
				data.ProjectTableView.Table.Box,
			))
	} else {
		focusable = append(
			focusable,
			misc.GetTUIItem(
				data.ProjectTreeView.Tree,
				data.ProjectTreeView.Tree.Box,
			))
	}

	// Always include Tags and Paths panes (even when empty)
	focusable = append(
		focusable,
		misc.GetTUIItem(
			data.TagView.List,
			data.TagView.List.Box))
	focusable = append(
		focusable,
		misc.GetTUIItem(
			data.PathView.List,
			data.PathView.List.Box))

	return focusable
}
