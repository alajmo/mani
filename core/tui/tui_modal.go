package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func isModalOpen() bool {
	frontPageName, _ := TUI.pages.GetFrontPage()
	switch frontPageName {
	case "help":
		return true
	case "project-description":
		return true
	}

	return false
}

func closeModal() {
	hideSearch()

	frontPageName, _ := TUI.pages.GetFrontPage()
	switch frontPageName {
	case "help":
		TUI.helpBtn.SetLabelColor(tcell.ColorWhite)
		TUI.pages.RemovePage("help")
	case "project-description":
		TUI.pages.RemovePage("project-description")
	}

	if isPageVisible("projects") {
		TUI.projectBtn.SetLabelColor(tcell.ColorYellow)
	} else if isPageVisible("tasks") {
		TUI.taskBtn.SetLabelColor(tcell.ColorYellow)
	} else if isPageVisible("run") {
		TUI.runBtn.SetLabelColor(tcell.ColorYellow)
	}
}

func openModal(pageTitle string, text string, title string, width int) {
	textView := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	textView.SetBorder(true).
		SetTitle(fmt.Sprintf("[yellow::b] %s ", title)).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorYellow).
		SetBorderPadding(1, 1, 2, 2)
	textView.SetBackgroundColor(tcell.ColorDefault)
	textView.SetTextColor(tcell.ColorWhite)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(textView, width, 1, true).
				AddItem(nil, 0, 1, false),
			0, 1, true,
		).
		AddItem(nil, 0, 1, false)
	flex.SetFullScreen(true).SetBackgroundColor(tcell.ColorBlack)

	hideSearch()

	TUI.projectBtn.SetLabelColor(tcell.ColorWhite)
	TUI.taskBtn.SetLabelColor(tcell.ColorWhite)
	TUI.runBtn.SetLabelColor(tcell.ColorWhite)

	TUI.pages.AddPage(pageTitle, flex, false, true)
	TUI.app.SetFocus(textView)
}

func showHelpModal() {
	helpText := "\n" +
		fmt.Sprintf("Version: %s\n\n", version) +
		"q: Quit\n" +
		"esc: Close Help\n" +
		"?: Show this Help\n" +
		"\n" +
		"1 | p: Switch to Projects\n" +
		"2 | t: Switch to Tasks\n" +
		"3 | r: Switch to Run\n" +
		"\n" +
		"Tab: Next pane\n" +
		"Shift + Tab: Previous pane\n" +
		"\n" +
		"Shift + v: Toggle project view (table|tree)\n" +
		"d: View project\n" +
		"Ctrl + a: Toggle select all\n" +
		"Shift + c: Clear all selections\n"

	TUI.projectBtn.SetLabelColor(tcell.ColorWhite)
	TUI.taskBtn.SetLabelColor(tcell.ColorWhite)
	TUI.runBtn.SetLabelColor(tcell.ColorWhite)
	TUI.helpBtn.SetLabelColor(tcell.ColorYellow)

	openModal("help", helpText, "Help", 80)
}
