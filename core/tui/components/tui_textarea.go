package components

import (
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

func CreateTextArea(title string) *tview.TextArea {
	textarea := tview.NewTextArea()
	textarea.SetBorder(true)
	textarea.SetWrap(true)
	textarea.SetTitle(title)
	textarea.SetTitleAlign(misc.STYLE_TITLE.Align)
	textarea.SetTitleColor(misc.STYLE_DEFAULT.Fg)
	textarea.SetBackgroundColor(misc.STYLE_DEFAULT.Bg)
	textarea.SetBorderPadding(0, 0, 1, 1)

	// Callbacks
	textarea.SetFocusFunc(func() {
		misc.PreviousPane = textarea
		misc.SetActive(textarea.Box, title, true)
	})
	textarea.SetBlurFunc(func() {
		misc.PreviousPane = textarea
		misc.SetActive(textarea.Box, title, false)
	})

	return textarea
}
