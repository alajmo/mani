package views

import (
	"os"

	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TSpec struct {
	View  *tview.Flex
	items []*tview.Box

	// Spec
	Output            string
	ClearBeforeRun    bool
	Parallel          bool
	IgnoreErrors      bool
	IgnoreNonExisting bool
	OmitEmptyRows     bool
	OmitEmptyColumns  bool
}

func CreateSpecView() *TSpec {
	defSpec, err := misc.Config.GetSpec("default")
	if err != nil {
		os.Exit(0)
	}

	spec := &TSpec{
		Output:            defSpec.Output,
		ClearBeforeRun:    defSpec.ClearOutput,
		Parallel:          defSpec.Parallel,
		IgnoreErrors:      defSpec.IgnoreErrors,
		IgnoreNonExisting: defSpec.IgnoreNonExisting,
		OmitEmptyRows:     defSpec.OmitEmptyRows,
		OmitEmptyColumns:  defSpec.OmitEmptyColumns,
	}

	view := tview.NewFlex().SetDirection(tview.FlexRow)
	view.SetBorder(true).SetBorderPadding(1, 1, 1, 1).
		SetBorderColor(misc.STYLE_BORDER_FOCUS.Fg).
		SetBorderPadding(1, 1, 2, 2)
	spec.View = view

	// Output type
	outputType := &components.TToggleText{
		Value:   &spec.Output,
		Option1: "stream",
		Option2: "table",
		Label1:  " Output stream ",
		Label2:  " Output table ",
		Data1:   "exec-stream",
		Data2:   "exec-table",
	}
	outputType.Create()

	clearBeforeRun := spec.AddCheckbox("Clear Before Run", &spec.ClearBeforeRun)
	parallel := spec.AddCheckbox("Parallel", &spec.Parallel)
	ignoreErrors := spec.AddCheckbox("Ignore Errors", &spec.IgnoreErrors)
	ignoreNonExisting := spec.AddCheckbox("Ignore Non Existing", &spec.IgnoreNonExisting)
	omitEmptyRows := spec.AddCheckbox("Omit Empty Rows", &spec.OmitEmptyRows)
	omitEmptyColumns := spec.AddCheckbox("Omit Empty Columns", &spec.OmitEmptyColumns)

	// Add checkboxes
	view.AddItem(outputType.TextView, 1, 0, false)
	view.AddItem(clearBeforeRun, 1, 0, false)
	view.AddItem(parallel, 1, 0, false)
	view.AddItem(ignoreErrors, 1, 0, false)
	view.AddItem(ignoreNonExisting, 1, 0, false)
	view.AddItem(omitEmptyRows, 1, 0, false)
	view.AddItem(omitEmptyColumns, 1, 0, false)

	checkboxes := []*tview.Box{
		outputType.TextView.Box,
		clearBeforeRun.Box,
		parallel.Box,
		ignoreErrors.Box,
		ignoreNonExisting.Box,
		omitEmptyRows.Box,
		omitEmptyColumns.Box,
	}

	// Input
	currentFocus := 0
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

	view.SetFocusFunc(func() {
		currentFocus = 0
		misc.App.SetFocus(outputType.TextView)
	})

	return spec
}

func (spec *TSpec) AddCheckbox(title string, checked *bool) *tview.Checkbox {
	onFocus := func() {}
	onBlur := func() {}

	checkbox := components.Checkbox(title, checked, onFocus, onBlur)
	spec.items = append(spec.items, checkbox.Box)
	return checkbox
}
