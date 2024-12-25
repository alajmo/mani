package components

import (
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateText(title string) *tview.TextView {
	textview := tview.NewTextView()
	textview.SetBorder(true)
	textview.SetBorderPadding(0, 0, 2, 1)
	textview.SetDynamicColors(true)
	textview.SetWrap(false)
	textTitle := title

	if textTitle != "" {
		textTitle = misc.Colorize(title, *misc.TUITheme.Title)
		textview.SetTitle(textTitle)
	}

	textview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		_, _, _, height := textview.GetInnerRect()
		row, _ := textview.GetScrollOffset()
		switch {
		case event.Key() == tcell.KeyCtrlD || event.Rune() == 'd':
			textview.ScrollTo(row+height/2, 0)
			return nil
		case event.Key() == tcell.KeyCtrlU || event.Rune() == 'u':
			textview.ScrollTo(row-height/2, 0)
			return nil
		case event.Key() == tcell.KeyCtrlF || event.Rune() == 'f':
			textview.ScrollTo(row+height, 0)
			return nil
		case event.Key() == tcell.KeyCtrlB || event.Rune() == 'b':
			textview.ScrollTo(row-height, 0)
			return nil
		}
		return event
	})

	// Callbacks
	textview.SetFocusFunc(func() {
		misc.PreviousPane = textview
		misc.SetActive(textview.Box, title, true)
	})
	textview.SetBlurFunc(func() {
		misc.PreviousPane = textview
		misc.SetActive(textview.Box, title, false)
	})

	return textview
}
