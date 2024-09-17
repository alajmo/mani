package tui

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func switchToPage(pageName string) {
	TUI.mainPage.SwitchToPage(pageName)

	switch pageName {
	case "projects":
		setActiveButtonStyle(TUI.projectBtn)

		setInactiveButtonStyle(TUI.helpBtn)
		setInactiveButtonStyle(TUI.runBtn)
		setInactiveButtonStyle(TUI.taskBtn)
		setInactiveButtonStyle(TUI.execBtn)
	case "tasks":
		setActiveButtonStyle(TUI.taskBtn)

		setInactiveButtonStyle(TUI.helpBtn)
		setInactiveButtonStyle(TUI.projectBtn)
		setInactiveButtonStyle(TUI.runBtn)
		setInactiveButtonStyle(TUI.execBtn)
	case "run":
		setActiveButtonStyle(TUI.runBtn)

		setInactiveButtonStyle(TUI.helpBtn)
		setInactiveButtonStyle(TUI.projectBtn)
		setInactiveButtonStyle(TUI.taskBtn)
		setInactiveButtonStyle(TUI.execBtn)
	case "exec":
		setActiveButtonStyle(TUI.execBtn)

		setInactiveButtonStyle(TUI.helpBtn)
		setInactiveButtonStyle(TUI.projectBtn)
		setInactiveButtonStyle(TUI.taskBtn)
		setInactiveButtonStyle(TUI.runBtn)
	}
}

func isPageVisible(pageName string) bool {
	if page, _ := TUI.mainPage.GetFrontPage(); page == pageName {
		return true
	}
	return false
}

func setActiveButtonStyle(button *tview.Button) {
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

func setInactiveButtonStyle(button *tview.Button) {
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

func createButton(label string) *tview.Button {
	button := tview.NewButton(label)
	return button
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

func getTask(tasks []dao.Task, taskName string) dao.Task {
	for index, project := range tasks {
		if project.Name == taskName {
			return tasks[index]
		}
	}
	return dao.Task{}
}

func removeTask(tasks []dao.Task, taskName string) []dao.Task {
	for index, project := range tasks {
		if project.Name == taskName {
			return append(tasks[:index], tasks[index+1:]...)
		}
	}
	return tasks
}

func isTaskSelected(tasks []dao.Task, taskName string) bool {
	for _, task := range tasks {
		if task.Name == taskName {
			return true
		}
	}
	return false
}

func getCurrentFocusIndex(focusableElements []tview.Primitive) int {
	for i, elem := range focusableElements {
		if elem.HasFocus() {
			return i
		}
	}

	return 0
}
