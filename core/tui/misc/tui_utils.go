package misc

import (
	"strings"

	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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

func FocusNext(elements []tview.Primitive) {
	currentFocus := App.GetFocus()
	for i, element := range elements {
		if element == currentFocus {
			nextIndex := (i + 1) % len(elements)
			App.SetFocus(elements[nextIndex])
			return
		}
	}
	// If current focus is not in the list, focus the first element
	if len(elements) > 0 {
		App.SetFocus(elements[0])
	}
}

func FocusPrevious(elements []tview.Primitive) {
	currentFocus := App.GetFocus()
	for i, element := range elements {
		if element == currentFocus {
			prevIndex := (i - 1 + len(elements)) % len(elements)
			App.SetFocus(elements[prevIndex])
			return
		}
	}
	// If current focus is not in the list, focus the last element
	if len(elements) > 0 {
		App.SetFocus(elements[len(elements)-1])
	}
}
