package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/alajmo/mani/core/tui/misc"
)

type TList struct {
	Root   *tview.Flex
	List   *tview.List
	Filter *tview.InputField

	Title       string
	FilterValue *string

	IsItemSelected   func(item string) bool
	ToggleSelectItem func(i int, itemName string)
	SelectAll        func()
	UnselectAll      func()
	FilterItems      func()
}

func (l *TList) Create() {
	// Init
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedStyle(misc.STYLE_ITEM_FOCUSED.Style).
		SetMainTextColor(misc.STYLE_ITEM.Fg)
	filter := CreateFilter()

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(filter, 1, 0, false)
	root.SetTitleColor(misc.STYLE_TITLE.Fg)
	root.SetTitleAlign(misc.STYLE_TITLE.Align).
		SetBorder(true).
		SetBorderPadding(1, 0, 1, 1)

	l.Filter = filter
	l.Root = root
	l.List = list

	if l.Title != "" {
		misc.SetActive(l.Root.Box, l.Title, false)
	}

	l.IsItemSelected = func(item string) bool { return false }
	l.ToggleSelectItem = func(i int, itemName string) {}
	l.SelectAll = func() {}
	l.UnselectAll = func() {}
	l.FilterItems = func() {}

	// Filter
	l.Filter.SetChangedFunc(func(_ string) {
		l.applyFilter()
		l.FilterItems()
	})

	l.Filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentFocus := misc.App.GetFocus()
		if currentFocus == filter {
			switch event.Key() {
			case tcell.KeyEscape:
				l.ClearFilter()
				misc.App.SetFocus(list)
				return nil
			case tcell.KeyEnter:
				l.applyFilter()
				misc.App.SetFocus(list)
			}
			return event
		}
		return event
	})

	// Input
	l.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Need to check filter in-case list is empty
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'f': // Filter
				ShowFilter(filter, *l.FilterValue)
				return nil
			case 'F': // Remove filter
				CloseFilter(filter)
				*l.FilterValue = ""
				return nil
			}
		}

		numItems := l.List.GetItemCount()
		if numItems == 0 {
			return nil
		}

		currentItemIndex := l.List.GetCurrentItem()
		_, secondaryText := l.List.GetItemText(currentItemIndex)
		switch event.Key() {
		case tcell.KeyEnter:
			l.ToggleSelectItem(currentItemIndex, secondaryText)
			return nil
		case tcell.KeyCtrlD:
			current := list.GetCurrentItem()
			_, _, _, height := list.GetInnerRect()
			newPos := min(current+height/2, list.GetItemCount()-1)
			list.SetCurrentItem(newPos)
			return nil
		case tcell.KeyCtrlU:
			current := list.GetCurrentItem()
			_, _, _, height := list.GetInnerRect()
			newPos := max(current-height/2, 0)
			list.SetCurrentItem(newPos)
			return nil
		case tcell.KeyCtrlF:
			current := list.GetCurrentItem()
			_, _, _, height := list.GetInnerRect()
			newPos := min(current+height, list.GetItemCount()-1)
			list.SetCurrentItem(newPos)
			return nil
		case tcell.KeyCtrlB:
			current := list.GetCurrentItem()
			_, _, _, height := list.GetInnerRect()
			newPos := max(current-height, 0)
			list.SetCurrentItem(newPos)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g': // top
				l.List.SetCurrentItem(0)
				return nil
			case 'G': // bottom
				l.List.SetCurrentItem(numItems - 1)
				return nil
			case 'j': // down
				nextItem := currentItemIndex + 1
				if nextItem < numItems {
					l.List.SetCurrentItem(nextItem)
				}
				return nil
			case 'k': // up
				nextItem := currentItemIndex - 1
				if nextItem >= 0 {
					l.List.SetCurrentItem(nextItem)
				}
				return nil
			case 'a': // Select all
				l.SelectAll()
				return nil
			case 'c': // Unselect all
				l.UnselectAll()
				return nil
			case ' ': // Select (Space)
				l.ToggleSelectItem(currentItemIndex, secondaryText)
				return nil
			}
		}

		return event
	})

	// Events
	l.List.SetFocusFunc(func() {
		misc.PreviousPane = l.List
		misc.SetActive(l.Root.Box, l.Title, true)
	})
	l.List.SetBlurFunc(func() {
		misc.PreviousPane = l.List
		misc.SetActive(l.Root.Box, l.Title, false)
	})
}

func (l *TList) Update(items []string) {
	l.List.Clear()
	for _, name := range items {
		l.List.AddItem(l.getItemText(name), name, 0, nil)
	}
}

func (l *TList) SetItemSelect(i int, item string) {
	if l.IsItemSelected(item) {
		value := misc.Colorize(item, *misc.TUITheme.ItemSelected)
		l.List.SetItemText(i, value, item)
	} else {
		value := misc.Colorize(item, *misc.TUITheme.Item)
		l.List.SetItemText(i, value, item)
	}
}

func (l *TList) ClearFilter() {
	CloseFilter(l.Filter)
	*l.FilterValue = ""
}

func (l *TList) applyFilter() {
	*l.FilterValue = l.Filter.GetText()
}

func (l *TList) getItemText(item string) string {
	if l.IsItemSelected(item) {
		value := misc.Colorize(item, *misc.TUITheme.ItemSelected)
		return value
	}
	return misc.PadString(item)
}
