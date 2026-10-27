package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/term"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
)

// OpenModal Used for when a custom tview Flex is passed to a modal.
func OpenModal(pageTitle string, title string, contentPane *tview.Flex, width int, height int) {
	termWidth, termHeight, _ := term.GetSize(0)
	if width > termWidth {
		width = termWidth - 5
	}
	if height > termHeight {
		height = termHeight - 5
	}

	formattedTitle := misc.ColorizeTitle(dao.StyleFormat(title, misc.STYLE_TITLE_ACTIVE.FormatStr), *misc.TUITheme.TitleActive)
	contentPane.SetTitle(formattedTitle)

	background := tview.NewBox()
	containerFlex := tview.NewFlex().
		AddItem(contentPane, 0, 1, true)
	containerFlex.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		background.SetRect(x, y, width, height)
		background.Draw(screen)
		contentPane.SetRect(x, y, width, height)
		contentPane.Draw(screen)
		return x, y, width, height
	})

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

	modal.SetFullScreen(true)

	EmptySearch()

	misc.Pages.AddPage(pageTitle, modal, false, true)
	misc.App.SetFocus(containerFlex)
}

// OpenTextModal Used for when text is passed to a modal.
func OpenTextModal(pageTitle string, textColor string, textNoColor string, title string) {
	width, height := misc.GetTexztModalSize(textNoColor)
	textColor = strings.TrimSpace(textColor)

	// Text
	contentPane := tview.NewTextView().
		SetText(textColor).
		SetTextAlign(tview.AlignLeft).
		SetDynamicColors(true)

		// Border

	formattedTitle := misc.ColorizeTitle(dao.StyleFormat(title, misc.STYLE_TITLE_ACTIVE.FormatStr), *misc.TUITheme.TitleActive)
	contentPane.SetBorder(true).
		SetTitle(formattedTitle).
		SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorderColor(misc.STYLE_BORDER_FOCUS.Fg).
		SetBorderPadding(1, 1, 2, 2)

		// Colors
	contentPane.SetBackgroundColor(misc.STYLE_DEFAULT.Bg)
	contentPane.SetTextColor(misc.STYLE_DEFAULT.Fg)

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

	modal.SetFullScreen(true).SetBackgroundColor(misc.STYLE_DEFAULT.Fg)

	EmptySearch()

	misc.Pages.AddPage(pageTitle, modal, false, true)
	misc.App.SetFocus(contentPane)
}

func CloseModal() {
	// Need to store before removing, because otherwise
	// the first pane gets focused and so misc.PreviousPage
	// doesn't work as intended.
	previousPane := misc.PreviousPane
	frontPageName, _ := misc.Pages.GetFrontPage()
	misc.Pages.RemovePage(frontPageName)
	misc.App.SetFocus(previousPane)
}

func IsModalOpen() bool {
	frontPageName, _ := misc.Pages.GetFrontPage()
	return strings.Contains(frontPageName, "-modal")
}
