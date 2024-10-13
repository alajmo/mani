package pages

import (
	"fmt"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/alajmo/mani/core/tui/views"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateProjectsPage(
	projects []dao.Project,
	projectTags []string,
	projectPaths []string,
) *tview.Flex {
	// Views
	data := views.CreateProjectsData(projects, projectTags, projectPaths, []string{"Project", "Description", "Tag"}, true)
	projectsTable := views.CreateProjectsTable(&data, false, "")
	tagsList := views.CreateProjectsTagsList(&data)
	pathsList := views.CreateProjectsPathsList(&data)

	// Context page
	data.ProjectsContextPage = tview.NewFlex().SetDirection(tview.FlexRow)
	if tagsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(tagsList.List, 0, 1, true)
	}
	if pathsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(pathsList.List, 0, 1, true)
	}

	// Page
	data.ProjectsPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(projectsTable.Table, 0, 1, true).
				AddItem(data.ProjectsContextPage, 30, 1, false),
			0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	// Focusable
	focusableElements := []*misc.TUIItem{misc.GetTUIItem("", projectsTable.Table, projectsTable.Table.Box)}
	if len(data.ProjectTags) > 0 {
		focusableElements = append(
			focusableElements,
			misc.GetTUIItem(
				fmt.Sprintf("Tags (%d)",
					len(data.ProjectTags)),
				tagsList.List,
				tagsList.List.Box))
	}
	if len(data.ProjectPaths) > 0 {
		focusableElements = append(
			focusableElements,
			misc.GetTUIItem(
				fmt.Sprintf("Paths (%d)",
					len(data.ProjectPaths)),
				pathsList.List,
				pathsList.List.Box))
	}
	focusableElements = append(focusableElements)

	// Shortcuts
	data.ProjectsPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyTab:
			misc.FocusNext(focusableElements)
			return nil
		case tcell.KeyBacktab:
			misc.FocusPrevious(focusableElements)
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i := int(event.Rune()-'0') - 1
				if i < len(focusableElements) {
					misc.App.SetFocus(focusableElements[i].Box)
				}
				return nil
			case 'f': // Clear filters
				data.Emitter.PublishAndWait(misc.Event{Name: "clear_filters", Data: ""})
				data.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
				return nil
			}
		}
		return event
	})

	return data.ProjectsPage
}
