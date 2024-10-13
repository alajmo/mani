package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

var Version = "dev"

func IsModalOpen() bool {
	frontPageName, _ := misc.Pages.GetFrontPage()
	return strings.Contains(frontPageName, "-modal")
}

func CloseModal() {
	frontPageName, _ := misc.Pages.GetFrontPage()
	misc.Pages.RemovePage(frontPageName)

	if frontPageName == "help-modal" {
		misc.HelpBtn.SetLabelColor(misc.THEME.FG)
	}

	// Nav buttons
	if misc.IsPageVisible("projects") {
		misc.ProjectBtn.SetLabelColor(misc.THEME.BTN_FG_ACTIVE)
	} else if misc.IsPageVisible("tasks") {
		misc.TaskBtn.SetLabelColor(misc.THEME.BTN_FG_ACTIVE)
	} else if misc.IsPageVisible("run") {
		misc.RunBtn.SetLabelColor(misc.THEME.BTN_FG_ACTIVE)
	} else if misc.IsPageVisible("exec") {
		misc.ExecBtn.SetLabelColor(misc.THEME.BTN_FG_ACTIVE)
	}
}

func OpenTextModal(pageTitle string, text string, title string, width int, height int) {
	text = strings.TrimSpace(text)

	// Text
	contentPane := tview.NewTextView().
		SetText(text).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

		// Border
	contentPane.SetBorder(true).
		SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.TITLE_ACTIVE, title)).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS).
		SetBorderPadding(1, 1, 2, 2)

		// Colors
	contentPane.SetBackgroundColor(misc.THEME.BG)
	contentPane.SetTextColor(misc.THEME.FG)

	// Container
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

	modal.SetFullScreen(true).SetBackgroundColor(misc.THEME.BG)

	EmptySearch()

	// Nav buttons
	misc.ProjectBtn.SetLabelColor(misc.THEME.TITLE)
	misc.TaskBtn.SetLabelColor(misc.THEME.TITLE)
	misc.RunBtn.SetLabelColor(misc.THEME.TITLE)
	misc.ExecBtn.SetLabelColor(misc.THEME.TITLE)

	misc.Pages.AddPage(pageTitle, modal, false, true)
	misc.App.SetFocus(contentPane)
}

func OpenModal(pageTitle string, title string, contentPane *tview.Flex, width int, height int) {
	contentPane.SetTitle(title)
	contentPane.SetTitleAlign(tview.AlignCenter)
	contentPane.SetBackgroundColor(misc.THEME.BG)

	background := tview.NewBox().SetBackgroundColor(misc.THEME.BG)
	containerFlex := tview.NewFlex().
		AddItem(contentPane, 0, 1, true)
	containerFlex.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		background.SetRect(x, y, width, height)
		background.Draw(screen)
		contentPane.SetRect(x, y, width, height)
		contentPane.Draw(screen)
		return x, y, width, height
	})

	containerFlex.SetBorder(true).
		SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.TITLE_ACTIVE, title)).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS).
		SetBorderPadding(1, 1, 2, 2)

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(nil, 0, 1, false).
				AddItem(containerFlex, width, 1, true).
				AddItem(nil, 0, 1, false),
			height, 1, true,
		).
		AddItem(nil, 0, 1, false)

	modal.SetFullScreen(true).SetBackgroundColor(tcell.ColorPurple)

	modal.Box.SetBackgroundColor(tcell.ColorYellow)

	EmptySearch()

	// Nav buttons
	misc.ProjectBtn.SetLabelColor(misc.THEME.TITLE)
	misc.TaskBtn.SetLabelColor(misc.THEME.TITLE)
	misc.RunBtn.SetLabelColor(misc.THEME.TITLE)
	misc.ExecBtn.SetLabelColor(misc.THEME.TITLE)

	misc.Pages.AddPage(pageTitle, modal, false, true)
	misc.App.SetFocus(containerFlex)
}

func ShowHelpModal() {
	helpText := "\n" +
		fmt.Sprintf("Version: %s\n\n", Version) +
		"q: Quit\n" +
		"esc: Close Modals\n" +
		"?: Show this Help\n" +
		"\n" +
		"p: Switch to Projects\n" +
		"t: Switch to Tasks\n" +
		"r: Switch to Run\n" +
		"e: Switch to Exec\n" +
		"\n" +
		"Navigation\n" +
		"Tab: Next pane\n" +
		"Shift + Tab: Previous pane\n" +
		"\n" +
		"Shift + v: Toggle project view (table|tree)\n" +
		"d: View project\n" +
		"Ctrl + a: Toggle select all\n" +
		"Shift + c: Clear all selections\n"

	// Nav buttons
	misc.ProjectBtn.SetLabelColor(misc.THEME.TITLE)
	misc.TaskBtn.SetLabelColor(misc.THEME.TITLE)
	misc.RunBtn.SetLabelColor(misc.THEME.TITLE)
	misc.ExecBtn.SetLabelColor(misc.THEME.TITLE)
	misc.HelpBtn.SetLabelColor(misc.THEME.TITLE_ACTIVE)

	OpenTextModal("help-modal", helpText, "Help", 80, 30)
}
