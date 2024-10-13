package components

import (
	// "github.com/alajmo/mani/core/tui/misc"

	"fmt"

	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

// type TUIText struct {
// 	Grid    *tview.Flex
// 	Headers *tview.Grid
// 	Rows    *tview.Grid
// 	Border  bool
// }

func CreateTextView(title string) *tview.TextView {
	textview := tview.NewTextView()
	textview.SetBorder(true)
	textview.SetDynamicColors(true)
	if title != "" {
		textview.SetTitle(fmt.Sprintf("[::b] %s ", title))
	}

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
