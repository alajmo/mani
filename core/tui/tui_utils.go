package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func switchToPage(page string) {
	TUI.mainPage.SwitchToPage(page)
	hideSearch()

	switch page {
	case "projects":
		setActiveButtonStyle(TUI.projectBtn)

		setInactiveButtonStyle(TUI.helpBtn)
		setInactiveButtonStyle(TUI.runBtn)
		setInactiveButtonStyle(TUI.taskBtn)
	case "tasks":
		setActiveButtonStyle(TUI.taskBtn)

		setInactiveButtonStyle(TUI.helpBtn)
		setInactiveButtonStyle(TUI.projectBtn)
		setInactiveButtonStyle(TUI.runBtn)
	case "run":
		setActiveButtonStyle(TUI.runBtn)

		setInactiveButtonStyle(TUI.helpBtn)
		setInactiveButtonStyle(TUI.projectBtn)
		setInactiveButtonStyle(TUI.taskBtn)
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
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorYellow).
			Bold(true)).
		SetActivatedStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorYellow).
			Bold(true))
}

func setInactiveButtonStyle(button *tview.Button) {
	button.
		SetStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorWhite).
			Bold(true)).
		SetActivatedStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorWhite).
			Bold(true))
}

func createButton(label string) *tview.Button {
	button := tview.NewButton(label).
		SetStyle(tcell.StyleDefault.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorWhite).
			Bold(true)).
		SetLabelColor(tcell.ColorWhite).
		SetLabelColorActivated(tcell.ColorWhite).
		SetBackgroundColorActivated(tcell.ColorDefault)

	return button
}

func setActive(box *tview.Box, title string, active bool) {
	if active {
		box.SetBorderColor(tcell.ColorYellow)
		box.SetTitle(fmt.Sprintf("[yellow::b] %s ", title))
	} else {
		box.SetBorderColor(tcell.ColorWhite)
		box.SetTitle(fmt.Sprintf("[white::b] %s ", title))
	}
}

func printList(title string, items []string) string {
	str := title
	for _, item := range items {
		str += fmt.Sprintf("%4s%s\n", " ", strings.Replace(strings.TrimSuffix(item, "\n"), "=", ": ", 1))
	}
	return str
}
