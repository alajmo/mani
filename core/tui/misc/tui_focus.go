package misc

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TItem struct {
	Primitive tview.Primitive
	Box       *tview.Box
}

func FocusNext(elements []*TItem) *tview.Primitive {
	if len(elements) == 0 {
		return nil
	}

	currentFocus := App.GetFocus()
	nextIndex := -1
	var nextFocusItem TItem
	for i, element := range elements {
		if element.Primitive == currentFocus {
			nextIndex = (i + 1) % len(elements)
			nextFocusItem = *elements[nextIndex]
		}
		element.Box.SetBorderColor(STYLE_BORDER.Fg)
	}

	// In-case no nextIndex is found, use the previous page as base to find nextFocusItem
	if nextIndex < 0 {
		for i, element := range elements {
			if element.Primitive == PreviousPane {
				nextIndex = (i + 1) % len(elements)
				nextFocusItem = *elements[nextIndex]
			}
		}
	}

	// Fallback to first element if still not found
	if nextIndex < 0 {
		nextFocusItem = *elements[0]
	}

	// Set border and focus
	nextFocusItem.Box.SetBorderColor(STYLE_BORDER_FOCUS.Fg)
	App.SetFocus(nextFocusItem.Primitive)

	return &nextFocusItem.Primitive
}

func FocusPrevious(elements []*TItem) *tview.Primitive {
	if len(elements) == 0 {
		return nil
	}

	currentFocus := App.GetFocus()
	prevIndex := -1
	var nextFocusItem TItem
	for i, element := range elements {
		if element.Primitive == currentFocus {
			prevIndex = (i - 1 + len(elements)) % len(elements)
			nextFocusItem = *elements[prevIndex]
		}
		element.Box.SetBorderColor(STYLE_BORDER.Fg)
	}

	// In-case no prevIndex is found, use the previous page as base to find nextFocusItem
	if prevIndex < 0 {
		for i, element := range elements {
			if element.Primitive == PreviousPane {
				prevIndex = (i - 1 + len(elements)) % len(elements)
				nextFocusItem = *elements[prevIndex]
			}
		}
	}

	// Fallback to first element if still not found
	if prevIndex < 0 {
		nextFocusItem = *elements[0]
	}

	// Set border and focus
	nextFocusItem.Box.SetBorderColor(STYLE_BORDER_FOCUS.Fg)
	App.SetFocus(nextFocusItem.Primitive)

	return &nextFocusItem.Primitive
}

func FocusPage(event *tcell.EventKey, focusable []*TItem) {
	i := int(event.Rune()-'0') - 1
	if i < len(focusable) {
		App.SetFocus(focusable[i].Box)
	}
}

func FocusPreviousPage() {
	App.SetFocus(PreviousPane)
}

func GetTUIItem(primitive tview.Primitive, box *tview.Box) *TItem {
	return &TItem{
		Primitive: primitive,
		Box:       box,
	}
}
