package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"

	"github.com/alajmo/mani/core/print"
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

	misc.App.SetFocus(misc.PreviousPage)
}

func OpenTextModal(pageTitle string, text string, title string, width int, height int) {
	width, height = getModalSize(text)
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

func shortcutString(shortcut string, text string) string {
	return fmt.Sprintf("[%s::b]%s[-::-]: %s \n", tcell.ColorGreen, shortcut, text)
}

func ShowHelpModal() {
	// versionString := fmt.Sprintf("Version: %s\n", Version)
	shortcutsHeader := fmt.Sprintf("[%s::b]Shortcuts\n", misc.THEME.TITLE)

	top := shortcutString("q", "Quit")
	top += shortcutString("Escape", "Close")
	top += shortcutString("?", "Show help")

	// Navigation
	navHeader := fmt.Sprintf("[%s::b]Navigation\n", misc.THEME.TITLE)
	middle := shortcutString("Tab", "Focus next pane")
	middle += shortcutString("Shift + Tab", "Focus previous pane")
	middle += shortcutString("1-9", "Focus pane")
	middle += shortcutString("r", "Switch to run page")
	middle += shortcutString("e", "Switch to exec page")
	middle += shortcutString("p", "Switch to projects page")
	middle += shortcutString("t", "Switch to tasks page")

	// Actions
	actionHeader := fmt.Sprintf("[%s::b]Actions\n", misc.THEME.TITLE)
	bottom := shortcutString("f", "Clear filters")
	bottom += shortcutString("a", "Select all")
	bottom += shortcutString("c", "Clear all selections")
	bottom += shortcutString("d", "Describe project or task")
	bottom += shortcutString("o", "Open project or task in editor")
	bottom += shortcutString("Ctrl + o", "Open task options")
	bottom += shortcutString("Ctrl + s", "Switch view")
	bottom += shortcutString("Ctrl + r", "Run tasks")

	helpText := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n",
		// versionString,
		shortcutsHeader,
		top,
		navHeader,
		middle,
		actionHeader,
		bottom,
	)

	// Nav buttons
	misc.ProjectBtn.SetLabelColor(misc.THEME.TITLE)
	misc.TaskBtn.SetLabelColor(misc.THEME.TITLE)
	misc.RunBtn.SetLabelColor(misc.THEME.TITLE)
	misc.ExecBtn.SetLabelColor(misc.THEME.TITLE)
	misc.HelpBtn.SetLabelColor(misc.THEME.TITLE_ACTIVE)

	OpenTextModal("help-modal", helpText, "Help", 80, 30)
}

func getModalSize(text string) (int, int) {
	termWidth, termHeight, _ := term.GetSize(0)
	textWidth, textHeight := print.GetTextDimensions(text)

	width := textWidth + 5
	height := textHeight + 3
	if termWidth < width {
		width = termWidth - 20
	}

	if termHeight < height {
		height = termHeight - 4
	}

	return width, height
}
