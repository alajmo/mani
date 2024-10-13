package misc

import (
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUIItem struct {
	Title     string
	Primitive tview.Primitive
	Box       *tview.Box
}

func GetTUIItem(title string, primitive tview.Primitive, box *tview.Box) *TUIItem {
	return &TUIItem{
		Title:     title,
		Primitive: primitive,
		Box:       box,
	}
}

func SwitchToPage(pageName string) {
	MainPage.SwitchToPage(pageName)

	switch pageName {
	case "projects":
		SetActiveButtonStyle(ProjectBtn)

		SetInactiveButtonStyle(HelpBtn)
		SetInactiveButtonStyle(RunBtn)
		SetInactiveButtonStyle(TaskBtn)
		SetInactiveButtonStyle(ExecBtn)
	case "tasks":
		SetActiveButtonStyle(TaskBtn)

		SetInactiveButtonStyle(HelpBtn)
		SetInactiveButtonStyle(ProjectBtn)
		SetInactiveButtonStyle(RunBtn)
		SetInactiveButtonStyle(ExecBtn)
	case "run":
		SetActiveButtonStyle(RunBtn)

		SetInactiveButtonStyle(HelpBtn)
		SetInactiveButtonStyle(ProjectBtn)
		SetInactiveButtonStyle(TaskBtn)
		SetInactiveButtonStyle(ExecBtn)
	case "exec":
		SetActiveButtonStyle(ExecBtn)

		SetInactiveButtonStyle(HelpBtn)
		SetInactiveButtonStyle(ProjectBtn)
		SetInactiveButtonStyle(TaskBtn)
		SetInactiveButtonStyle(RunBtn)
	}

	_, page := MainPage.GetFrontPage()
	App.SetFocus(page)
}

func IsPageVisible(pageName string) bool {
	if page, _ := MainPage.GetFrontPage(); page == pageName {
		return true
	}
	return false
}

func SetActiveButtonStyle(button *tview.Button) {
	button.
		SetStyle(tcell.StyleDefault.
			Background(THEME.BTN_BG_ACTIVE).
			Foreground(THEME.BTN_FG_ACTIVE).
			Bold(true)).
		SetActivatedStyle(tcell.StyleDefault.
			Background(THEME.BTN_BG_ACTIVE).
			Foreground(THEME.BTN_FG_ACTIVE).
			Bold(true))
}

func SetInactiveButtonStyle(button *tview.Button) {
	button.
		SetStyle(tcell.StyleDefault.
			Background(THEME.BTN_BG).
			Foreground(THEME.BTN_FG).
			Bold(true)).
		SetActivatedStyle(tcell.StyleDefault.
			Background(THEME.BTN_BG).
			Foreground(THEME.BTN_FG).
			Bold(true))
}

func CreateButton(label string) *tview.Button {
	button := tview.NewButton(label)
	return button
}

func GetProject(projects []dao.Project, projectName string) dao.Project {
	for index, project := range projects {
		if project.Name == projectName {
			return projects[index]
		}
	}
	return dao.Project{}
}

func RemoveProject(projects []dao.Project, projectName string) []dao.Project {
	for index, project := range projects {
		if project.Name == projectName {
			return append(projects[:index], projects[index+1:]...)
		}
	}
	return projects
}

func IsProjectSelected(projects []dao.Project, projectName string) bool {
	for _, project := range projects {
		if project.Name == projectName {
			return true
		}
	}
	return false
}

func GetTask(tasks []dao.Task, taskName string) dao.Task {
	for index, project := range tasks {
		if project.Name == taskName {
			return tasks[index]
		}
	}
	return dao.Task{}
}

func RemoveTask(tasks []dao.Task, taskName string) []dao.Task {
	for index, project := range tasks {
		if project.Name == taskName {
			return append(tasks[:index], tasks[index+1:]...)
		}
	}
	return tasks
}

func IsTaskSelected(tasks []dao.Task, taskName string) bool {
	for _, task := range tasks {
		if task.Name == taskName {
			return true
		}
	}
	return false
}

func GetCurrentFocusIndex(focusableElements []tview.Primitive) int {
	for i, elem := range focusableElements {
		if elem.HasFocus() {
			return i
		}
	}

	return 0
}

func FocusPreviousPage() {
	App.SetFocus(PreviousPage)
}

func CalculateTextHeight(text string) int {
	lines := strings.Split(text, "\n")
	return Max(len(lines), 1)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func FocusNext(elements []*TUIItem) {
	currentFocus := App.GetFocus()
	nextIndex := -1
	var nextFocusItem TUIItem
	for i, element := range elements {
		if element.Primitive == currentFocus {
			element.Box.SetBorderColor(THEME.BORDER_COLOR)

			if element.Title != "" {
				element.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR, element.Title))
			}

			nextIndex = (i + 1) % len(elements)
			nextFocusItem = *elements[nextIndex]
		} else {
			element.Box.SetBorderColor(THEME.BORDER_COLOR)

			if element.Title != "" {
				element.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR, element.Title))
			}
		}
	}

	// In-case no nextIndex is found, use the previous page as base to find nextFocusItem
	if nextIndex < 0 {
		for i, element := range elements {
			if element.Primitive == PreviousPage {
				nextIndex = (i + 1) % len(elements)
				nextFocusItem = *elements[nextIndex]
			}
		}
	}

	// Set border and focus
	nextFocusItem.Box.SetBorderColor(THEME.BORDER_COLOR_FOCUS)
	if nextFocusItem.Title != "" {
		nextFocusItem.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR_FOCUS, nextFocusItem.Title))
	}
	App.SetFocus(nextFocusItem.Primitive)
}

func FocusPrevious(elements []*TUIItem) {
	currentFocus := App.GetFocus()
	prevIndex := -1
	var nextFocusItem TUIItem
	for i, element := range elements {
		if element.Primitive == currentFocus {
			element.Box.SetBorderColor(THEME.BORDER_COLOR)

			if element.Title != "" {
				element.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR, element.Title))
			}

			prevIndex = (i - 1 + len(elements)) % len(elements)
			nextFocusItem = *elements[prevIndex]
		} else {
			element.Box.SetBorderColor(THEME.BORDER_COLOR)

			if element.Title != "" {
				element.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR, element.Title))
			}
		}
	}

	// In-case no prevIndex is found, use the previous page as base to find nextFocusItem
	for i, element := range elements {
		if element.Primitive == PreviousPage {
			prevIndex = (i - 1 + len(elements)) % len(elements)
			nextFocusItem = *elements[prevIndex]
		}
	}

	// Set border and focus
	nextFocusItem.Box.SetBorderColor(THEME.BORDER_COLOR_FOCUS)
	if nextFocusItem.Title != "" {
		nextFocusItem.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR_FOCUS, nextFocusItem.Title))
	}
	App.SetFocus(nextFocusItem.Primitive)
}

func IsChildrenFocused(children []*tview.Box) bool {
	for _, box := range children {
		if box.HasFocus() {
			return true
		}
	}

	return false
}

func SetActive(box *tview.Box, title string, active bool) {
	if active {
		box.SetBorderColor(THEME.BORDER_COLOR_FOCUS)
		if title != "" {
			box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR_FOCUS, title))
		}
	} else {
		box.SetBorderColor(THEME.BORDER_COLOR)
		if title != "" {
			box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR, title))
		}
	}
}
