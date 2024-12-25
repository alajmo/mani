package components

import (
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

func Checkbox(label string, checked *bool, onFocus func(), onBlur func()) *tview.Checkbox {
	checkbox := tview.NewCheckbox().SetLabel(" " + label + " ")
	checkbox.SetChecked(*checked)
	checkbox.SetCheckedStyle(misc.STYLE_ITEM_SELECTED.Style)
	checkbox.SetUncheckedStyle(misc.STYLE_ITEM.Style)

	checkbox.SetFieldTextColor(misc.STYLE_ITEM_FOCUSED.Bg)
	checkbox.SetFieldBackgroundColor(misc.STYLE_ITEM.Bg)
	checkbox.SetCheckedString("")

	if *checked {
		checkbox.SetLabelStyle(misc.STYLE_ITEM_SELECTED.Style)
	} else {
		checkbox.SetLabelStyle(misc.STYLE_ITEM.Style)
	}

	// Callbacks
	checkbox.SetFocusFunc(func() {
		if *checked {
			checkbox.SetLabelColor(misc.STYLE_ITEM_SELECTED.Fg)
		} else {
			checkbox.SetLabelColor(misc.STYLE_ITEM_FOCUSED.Fg)
		}

		checkbox.SetBackgroundColor(misc.STYLE_ITEM_FOCUSED.Bg)
		onFocus()
	})
	checkbox.SetBlurFunc(func() {
		if *checked {
			checkbox.SetLabelColor(misc.STYLE_ITEM_SELECTED.Fg)
		} else {
			checkbox.SetLabelColor(misc.STYLE_ITEM.Fg)
		}

		checkbox.SetBackgroundColor(misc.STYLE_ITEM.Bg)
		onBlur()
	})
	checkbox.SetChangedFunc(func(isChecked bool) {
		if isChecked {
			checkbox.SetLabelStyle(misc.STYLE_ITEM_SELECTED.Style)
		} else {
			checkbox.SetLabelStyle(misc.STYLE_ITEM.Style)
		}
		*checked = !*checked
	})

	return checkbox
}
