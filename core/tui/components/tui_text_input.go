package components

import (
	// "github.com/alajmo/mani/core/tui/misc"

	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

// type TUIText struct {
// 	Grid    *tview.Flex
// 	Headers *tview.Grid
// 	Rows    *tview.Grid
// 	Border  bool
// }

func CreateTextInputView(title string) *tview.TextView {
	textview := tview.NewTextView()
	textview.SetBorder(true)
	textview.SetDynamicColors(true)
	// streamView.SetChangedFunc(func() {
	// 	misc.App.Draw()
	// })
	textview.SetTitle(title)

	textview.SetFocusFunc(func() {
		misc.PreviousPage = textview
		textview.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		misc.SetActive(textview.Box, title, true)
	})
	textview.SetBlurFunc(func() {
		misc.PreviousPage = textview
		misc.SetActive(textview.Box, title, false)
	})

	return textview
}
