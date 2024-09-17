package tui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

func isModalOpen() bool {
	frontPageName, _ := TUI.pages.GetFrontPage()
	return strings.Contains(frontPageName, "-modal")
}

func closeModal() {
	frontPageName, _ := TUI.pages.GetFrontPage()
	TUI.pages.RemovePage(frontPageName)

	if frontPageName == "help-modal" {
		TUI.helpBtn.SetLabelColor(THEME.FG)
	}

	// updateNavButtons(frontPageName)
	// Nav buttons
	if isPageVisible("projects") {
		TUI.projectBtn.SetLabelColor(THEME.BTN_FG_ACTIVE)
	} else if isPageVisible("tasks") {
		TUI.taskBtn.SetLabelColor(THEME.BTN_FG_ACTIVE)
	} else if isPageVisible("run") {
		TUI.runBtn.SetLabelColor(THEME.BTN_FG_ACTIVE)
	} else if isPageVisible("exec") {
		TUI.execBtn.SetLabelColor(THEME.BTN_FG_ACTIVE)
	}
}

func openModal(pageTitle string, text string, title string, width int, height int) {
	text = strings.TrimSpace(text)

	contentPane := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)
	contentPane.SetBorder(true).
		SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.TITLE_ACTIVE, title)).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(THEME.BORDER_COLOR_FOCUS).
		SetBorderPadding(1, 1, 2, 2)
	contentPane.SetBackgroundColor(THEME.BG)
	contentPane.SetTextColor(THEME.FG)

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(contentPane, width, 1, true).
				AddItem(nil, 0, 1, false),
			height, 1, true,
		).
		AddItem(nil, 0, 1, false)

	modal.SetFullScreen(true).SetBackgroundColor(THEME.BG)

	emptySearch()

	// Nav buttons
	TUI.projectBtn.SetLabelColor(THEME.TITLE)
	TUI.taskBtn.SetLabelColor(THEME.TITLE)
	TUI.runBtn.SetLabelColor(THEME.TITLE)
	TUI.execBtn.SetLabelColor(THEME.TITLE)

	TUI.pages.AddPage(pageTitle, modal, false, true)
	TUI.app.SetFocus(contentPane)
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

	// Nav buttons
	TUI.projectBtn.SetLabelColor(THEME.TITLE)
	TUI.taskBtn.SetLabelColor(THEME.TITLE)
	TUI.runBtn.SetLabelColor(THEME.TITLE)
	TUI.execBtn.SetLabelColor(THEME.TITLE)
	TUI.helpBtn.SetLabelColor(THEME.TITLE_ACTIVE)

	openModal("help-modal", helpText, "Help", 80, 30)
}
