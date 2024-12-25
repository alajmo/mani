package components

import (
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TToggleText struct {
	Value    *string
	Option1  string
	Option2  string
	Label1   string
	Label2   string
	Data1    string
	Data2    string
	TextView *tview.TextView
}

func (t *TToggleText) Create() {
	textview := tview.NewTextView()
	textview.SetTitle("")
	if *t.Value == t.Option1 {
		textview.SetText(t.Label1)
	} else {
		textview.SetText(t.Label2)
	}
	textview.SetSize(1, 18)
	textview.SetBorder(false)
	textview.SetBorderPadding(0, 0, 0, 0)
	textview.SetBackgroundColor(misc.STYLE_ITEM.Bg)

	toggleOutput := func() {

		if *t.Value == t.Option1 {
			*t.Value = t.Option2
			textview.SetText(t.Label2)
		} else {
			*t.Value = t.Option1
			textview.SetText(t.Label1)
		}
	}

	textview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			toggleOutput()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case ' ': // space
				toggleOutput()
				return nil
			}
		}

		return event
	})

	textview.SetFocusFunc(func() {
		textview.SetTextColor(misc.STYLE_ITEM_FOCUSED.Fg)
		textview.SetBackgroundColor(misc.STYLE_ITEM_FOCUSED.Bg)
	})

	textview.SetBlurFunc(func() {
		textview.SetTextColor(misc.STYLE_ITEM.Fg)
		textview.SetBackgroundColor(misc.STYLE_ITEM.Bg)
	})

	t.TextView = textview
}
