package views

import (
	"fmt"

	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUISpec struct {
	View      *tview.Flex
	items     []*tview.Box
	onNoFocus func()

	// Spec
	Output            string
	ClearBeforeRun    bool
	Parallel          bool
	IgnoreErrors      bool
	IgnoreNonExisting bool
}

func (spec *TUISpec) AddCheckbox(title string, checked *bool) *tview.Checkbox {
	onFocus := func() {
		spec.View.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		spec.View.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, "Spec"))
	}
	onBlur := func() {
		spec.checkFocus()
	}

	checkbox := components.Checkbox(title, checked, onFocus, onBlur)
	spec.items = append(spec.items, checkbox.Box)
	return checkbox
}

func (spec *TUISpec) checkFocus() {
	go func() {
		misc.App.QueueUpdateDraw(func() {
			for _, cb := range spec.items {
				if cb.HasFocus() {
					return
				}
			}

			if spec.onNoFocus != nil {
				spec.onNoFocus()
			}
		})
	}()
}

func CreateSpecView(emitter *misc.EventEmitter, data *TUISpec) *tview.Flex {
	// Main view
	view := tview.NewFlex().SetDirection(tview.FlexRow)
	view.SetTitle("Spec")
	view.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	view.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
	view.SetBackgroundColor(tcell.ColorGreen)
	view.Box.SetBackgroundColor(tcell.ColorGreen)
	data.View = view

	// Create checkboxes
	outputType := TextView(emitter, &data.Output)
	clearBeforeRun := data.AddCheckbox("Clear Before Run", &data.ClearBeforeRun)
	parallel := data.AddCheckbox("Parallel", &data.Parallel)
	ignoreErrors := data.AddCheckbox("Ignore Errors", &data.IgnoreErrors)
	ignoreNonExisting := data.AddCheckbox("Ignore Non Existing", &data.IgnoreNonExisting)

	// Add checkboxes
	view.AddItem(outputType, 1, 0, false)
	view.AddItem(clearBeforeRun, 1, 0, false)
	view.AddItem(parallel, 1, 0, false)
	view.AddItem(ignoreErrors, 1, 0, false)
	view.AddItem(ignoreNonExisting, 1, 0, false)

	// Input
	checkboxes := []*tview.Box{outputType.Box, clearBeforeRun.Box, parallel.Box, ignoreErrors.Box, ignoreNonExisting.Box}
	currentFocus := 0
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

	// Events
	data.onNoFocus = func() {
		view.SetBorderColor(misc.THEME.BORDER_COLOR_FOCUS)
		view.SetTitle(fmt.Sprintf("[%s::b] %s ", misc.THEME.BORDER_COLOR_FOCUS, "Spec"))
	}
	view.SetFocusFunc(func() {
		currentFocus = 0
		misc.App.SetFocus(outputType)
	})

	return view
}

func TextView(emitter *misc.EventEmitter, output *string) *tview.TextView {
	textview := tview.NewTextView()

	textview.SetTitle("")
	textview.SetText("Output text")
	textview.SetSize(1, 18)
	textview.SetBorder(false)
	textview.SetBorderPadding(0, 0, 0, 0)
	textview.SetBackgroundColor(misc.THEME.BG)

	toggleOutput := func() {
		if *output == "text" {
			*output = "table"
			textview.SetText("Output table")
			emitter.Publish(misc.Event{Name: "toggle_output", Data: "exec-table"})
		} else {
			*output = "text"
			textview.SetText("Output text")
			emitter.Publish(misc.Event{Name: "toggle_output", Data: "exec-text"})
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
		textview.SetTextColor(tcell.ColorWhite)
		textview.SetBackgroundColor(misc.THEME.BG_FOCUSED)
	})

	textview.SetBlurFunc(func() {
		textview.SetBackgroundColor(misc.THEME.BG)
	})

	return textview
}
