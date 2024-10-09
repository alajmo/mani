package views

import (
	"fmt"

	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUISpec struct {
	// Spec
	Output            []string
	Parallel          bool
	IgnoreErrors      bool
	IgnoreNonExisting bool
}

// if isFocused {
// 	if isSelected {
// 		style = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
// 		selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
// 	} else {
// 		style = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED).Background(misc.THEME.BG_FOCUSED)
// 		selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED).Background(misc.THEME.BG_FOCUSED)
// 	}
// } else {
// 	if isSelected {
// 		style = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
// 		selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG_FOCUSED_SELECTED).Background(misc.THEME.BG_FOCUSED_SELECTED).Attributes(tcell.AttrBold)
// 	} else {
// 		style = tcell.StyleDefault.Foreground(misc.THEME.FG).Background(misc.THEME.BG)
// 		selectedStyle = tcell.StyleDefault.Foreground(misc.THEME.FG).Background(misc.THEME.BG)
// 	}
// }

func CreateSpecView(spec *TUISpec) *tview.Flex {
	view := tview.NewFlex().SetDirection(tview.FlexRow)
	view.SetTitle("Spec")
	view.SetBorder(true).SetBorderPadding(1, 0, 1, 1)
	view.SetBorderColor(misc.THEME.BORDER_COLOR)

	parallel := components.Checkbox("Parallel", &spec.Parallel)
	ignoreErrors := components.Checkbox("Ignore Errors", &spec.IgnoreErrors)
	ignoreNonExisting := components.Checkbox("Ignore Non Existing", &spec.IgnoreNonExisting)

	view.AddItem(parallel, 1, 0, false)
	view.AddItem(ignoreErrors, 1, 0, false)
	view.AddItem(ignoreNonExisting, 1, 0, false)

	checkboxes := []*tview.Box{parallel.Box, ignoreErrors.Box, ignoreNonExisting.Box}
	currentFocus := -1

	parallel.SetFocusFunc(func() {
		view.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		view.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, "Spec"))
	})
	ignoreErrors.SetFocusFunc(func() {
		view.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		view.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, "Spec"))
	})
	ignoreNonExisting.SetFocusFunc(func() {
		view.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		view.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, "Spec"))
	})

	// Events
	view.SetFocusFunc(func() {
		// view.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		// view.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, "Spec"))
		// currentFocus = 0
		// misc.App.SetFocus(parallel)
	})
	// view.SetBlurFunc(func() {
	// 	// TODO: This gets triggered before the h
	// 	// isChildrenFocused := misc.IsChildrenFocused(checkboxes)
	// 	if currentFocus < 0 {
	// 		view.SetBorderColor(misc.THEME.BORDER_COLOR)
	// 		view.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR, "Spec"))
	// 	}
	// })

	// checkboxes := []*tview.Box{parallel.Box, ignoreErrors.Box, ignoreNonExisting.Box}
	// currentFocus := 0

	// Input
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		numItems := view.GetItemCount()
		if numItems == 0 {
			return nil
		}

		switch event.Key() {
		case tcell.KeyDown:
			if currentFocus < (len(checkboxes) - 1) {
				currentFocus += 1
				misc.App.SetFocus(checkboxes[currentFocus])
			}
			return nil
		case tcell.KeyUp:
			if currentFocus > 0 {
				currentFocus -= 1
				misc.App.SetFocus(checkboxes[currentFocus])
			}
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g': // top
				currentFocus = 0
				misc.App.SetFocus(checkboxes[currentFocus])
				return nil
			case 'G': // bottom
				currentFocus = len(checkboxes) - 1
				misc.App.SetFocus(checkboxes[currentFocus])
				return nil
			case 'j': // down
				if currentFocus < (len(checkboxes) - 1) {
					currentFocus += 1
					misc.App.SetFocus(checkboxes[currentFocus])
				}
				return nil
			case 'k': // up
				if currentFocus > 0 {
					currentFocus -= 1
					misc.App.SetFocus(checkboxes[currentFocus])
				}
				return nil
			}
		}

		return event
	})

	// form.SetLabelColor(misc.THEME.FG)
	// form.SetFieldBackgroundColor(misc.THEME.BG)
	// form.SetItemPadding(0)

	// Output            []string
	// Parallel          bool
	// IgnoreErrors      bool
	// IgnoreNonExisting bool

	return view
}
