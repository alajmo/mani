package components

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateButton(label string) *tview.Button {
	label = dao.StyleFormat(label, misc.STYLE_BUTTON.FormatStr)
	button := tview.NewButton(label)
	SetInactiveButtonStyle(button)
	return button
}

func SetActiveButtonStyle(button *tview.Button) {
	label := button.GetLabel()
	button.SetLabel(dao.StyleFormat(label, misc.STYLE_BUTTON_ACTIVE.FormatStr))
	button.
		SetStyle(tcell.StyleDefault.
			Foreground(misc.STYLE_BUTTON_ACTIVE.Fg).
			Background(misc.STYLE_BUTTON_ACTIVE.Bg).
			Attributes(misc.STYLE_BUTTON_ACTIVE.Attr)).
		SetActivatedStyle(tcell.StyleDefault.
			Foreground(misc.STYLE_BUTTON_ACTIVE.Fg).
			Background(misc.STYLE_BUTTON_ACTIVE.Bg).
			Attributes(misc.STYLE_BUTTON_ACTIVE.Attr))
}

func SetInactiveButtonStyle(button *tview.Button) {
	label := button.GetLabel()
	button.SetLabel(dao.StyleFormat(label, misc.STYLE_BUTTON.FormatStr))
	button.
		SetStyle(tcell.StyleDefault.
			Foreground(misc.STYLE_BUTTON.Fg).
			Background(misc.STYLE_BUTTON.Bg).
			Attributes(misc.STYLE_BUTTON.Attr)).
		SetActivatedStyle(tcell.StyleDefault.
			Foreground(misc.STYLE_BUTTON.Fg).
			Background(misc.STYLE_BUTTON.Bg).
			Attributes(misc.STYLE_BUTTON.Attr))
}
