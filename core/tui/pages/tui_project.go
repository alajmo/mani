package pages

import (
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
	data := views.CreateProjectsData(projects, projectTags, projectPaths)
	projectsTable := views.CreateProjectsTable(&data)
	tagsList := views.CreateProjectsTagsList(&data)
	pathsList := views.CreateProjectsPathsList(&data)
	selectedList := views.CreateProjectsSelectedList(&data)

	// Projects context
	data.ProjectsContextPage = tview.NewFlex().SetDirection(tview.FlexRow)
	if tagsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(tagsList.List, 0, 1, true)
	}
	if pathsList.List.GetItemCount() > 0 {
		data.ProjectsContextPage.AddItem(pathsList.List, 0, 1, true)
	}
	data.ProjectsContextPage.AddItem(selectedList.List, 0, 1, true)

	data.ProjectsPage = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(projectsTable.Table, 0, 1, true).
				AddItem(data.ProjectsContextPage, 30, 1, false),
			0, 1, true).
		AddItem(misc.Search, 1, 0, false)

	focusableElements := []tview.Primitive{projectsTable.Table}
	if len(data.ProjectTags) > 0 {
		focusableElements = append(focusableElements, tagsList.List)
	}
	if len(data.ProjectPaths) > 0 {
		focusableElements = append(focusableElements, pathsList.List)
	}
	focusableElements = append(focusableElements, selectedList.List)

	currentFocus := 0
	// Handle global shortcuts
	data.ProjectsPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if misc.App.GetFocus() == misc.Search {
			return event
		}

		switch event.Key() {
		case tcell.KeyTab:
			currentFocus = (currentFocus + 1) % len(focusableElements)
			misc.App.SetFocus(focusableElements[currentFocus])
			return nil
		case tcell.KeyBacktab:
			currentFocus = (currentFocus - 1 + len(focusableElements)) % len(focusableElements)
			misc.App.SetFocus(focusableElements[currentFocus])
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case '1': // Table focus
				misc.App.SetFocus(projectsTable.Table)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '2': // Tags focus
				misc.App.SetFocus(tagsList.List)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '3': // Paths focus
				misc.App.SetFocus(pathsList.List)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case '4': // Selected focus
				misc.App.SetFocus(selectedList.List)
				currentFocus = misc.GetCurrentFocusIndex(focusableElements)
				return nil
			case 'a': // Select all
				misc.Emitter.Publish(misc.Event{Name: "select_all_projects", Data: ""})
				return nil
			case 'c': // Unselect all all
				misc.Emitter.Publish(misc.Event{Name: "deselect_all_projects", Data: ""})
				return nil
			case 'f': // Clear filters
				misc.Emitter.PublishAndWait(misc.Event{Name: "clear_filters", Data: ""})
				misc.Emitter.Publish(misc.Event{Name: "filter_projects", Data: ""})
				return nil
			}
		}
		return event
	})

	return data.ProjectsPage
}
