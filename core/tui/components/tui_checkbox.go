package components

import (
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func Checkbox(label string, checked *bool) *tview.Checkbox {
	// Label Style
	selectedStyle := tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG).Attributes(tcell.AttrBold)
	nonSelectedStyle := tcell.StyleDefault.Foreground(misc.THEME.FG).Background(misc.THEME.BG).Attributes(tcell.AttrNone)

	// Checkbox marker style
	checkedStyle := tcell.StyleDefault.Background(misc.THEME.BG).Foreground(misc.THEME.FG_FOCUSED_SELECTED)
	uncheckedStyle := tcell.StyleDefault.Background(misc.THEME.BG).Foreground(misc.THEME.FG)

	checkbox := tview.NewCheckbox().SetLabel(label)
	checkbox.SetChecked(*checked)
	checkbox.SetCheckedStyle(checkedStyle)
	checkbox.SetUncheckedStyle(uncheckedStyle)
	if *checked {
		checkbox.SetLabelStyle(selectedStyle)
	} else {
		checkbox.SetLabelStyle(nonSelectedStyle)
	}
	checkbox.SetFieldTextColor(misc.THEME.BG_FOCUSED)
	checkbox.SetFieldBackgroundColor(misc.THEME.BG)
	checkbox.SetCheckedString("")

	checkbox.SetFocusFunc(func() {
		checkbox.SetBackgroundColor(misc.THEME.BG_FOCUSED)
	})
	checkbox.SetBlurFunc(func() {
		checkbox.SetBackgroundColor(misc.THEME.BG)
	})
	checkbox.SetChangedFunc(func(checked bool) {
		if checked {
			checkbox.SetLabelStyle(selectedStyle)
		} else {
			checkbox.SetLabelStyle(nonSelectedStyle)
		}
	})

	return checkbox
}
