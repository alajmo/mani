package components

import (
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

func CreateFilter() *tview.InputField {
	filter := tview.NewInputField().
		SetLabel("").
		SetLabelStyle(misc.STYLE_FILTER_LABEL.Style).
		SetFieldStyle(misc.STYLE_FILTER_TEXT.Style)

	return filter
}

func ShowFilter(filter *tview.InputField, text string) {
	filter.SetLabel(misc.Colorize("Filter:", *misc.TUITheme.FilterLabel))
	filter.SetText(text)
	misc.App.SetFocus(filter)
}

func CloseFilter(filter *tview.InputField) {
	filter.SetLabel("")
	filter.SetText("")
}

func InitFilter(filter *tview.InputField, text string) {
	if text != "" {
		filter.SetLabel(" Filter: ")
		filter.SetText(text)
	} else {
		filter.SetLabel("")
		filter.SetText("")
	}
}
