package tui

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUIList struct {
	Title      string
	List       *tview.List
	Items      map[string]bool
	SelectItem func(i int, mainText string, SecondaryText string)
}

func (l *TUIList) createList() {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(THEME.FG).
		SetSelectedStyle(tcell.StyleDefault.Foreground(THEME.FG).Background(THEME.BG_FOCUSED_SELECTED)).
		SetMainTextColor(THEME.FG)
	l.List = list

	// Items
	var items []string
	for item := range l.Items {
		items = append(items, item)
	}
	sort.Strings(items)
	for _, item := range items {
		list.AddItem(item, item, 0, nil)
	}

	list.
		SetTitle(fmt.Sprintf("[::b] %s ", l.getTitle())).
		SetBorder(true).
		SetBorderPadding(1, 1, 1, 1)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		numItems := list.GetItemCount()
		if numItems == 0 {
			return nil
		}

		currentItem := list.GetCurrentItem()
		mainText, secondaryText := list.GetItemText(currentItem)

		switch event.Key() {
		case tcell.KeyEnter:
			l.SelectItem(currentItem, mainText, secondaryText)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g': // top
				list.SetCurrentItem(0)
				return nil
			case 'G': // bottom
				list.SetCurrentItem(numItems - 1)
				return nil
			case 'j': // down
				nextItem := currentItem + 1
				if nextItem < numItems {
					list.SetCurrentItem(nextItem)
				}
				return nil
			case 'k': // up
				nextItem := currentItem - 1
				if nextItem >= 0 {
					list.SetCurrentItem(nextItem)
				}
				return nil
			case ' ': // Select (Space)
				l.SelectItem(currentItem, mainText, secondaryText)
				return nil
			}
		}

		return event
	})

	list.SetFocusFunc(func() {
		l.setActive(true)
	})
	list.SetBlurFunc(func() {
		TUI.previousPage = list
		l.setActive(false)
	})
}

// Called inside SelectItem
func (l *TUIList) handleSelectItem(i int, mainText string, secondaryText string) {
	l.Items[secondaryText] = !l.Items[secondaryText]
	if l.Items[secondaryText] {
		l.List.SetItemText(i, "[blue::b]"+mainText, secondaryText)
	} else {
		l.List.SetItemText(i, secondaryText, secondaryText)
	}
}

func (l *TUIList) getTitle() string {
	l.List.GetItemCount()
	count := l.List.GetItemCount()
	if count > 0 {
		return fmt.Sprintf("%s (%d)", l.Title, count)
	}

	return l.Title
}

func (l *TUIList) setActive(active bool) {
	title := l.getTitle()

	if active {
		l.List.Box.SetBorderColor(THEME.BORDER_COLOR_FOCUS)
		l.List.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR_FOCUS, title))
	} else {
		l.List.Box.SetBorderColor(THEME.BORDER_COLOR)
		l.List.Box.SetTitle(fmt.Sprintf("[%s::b] %s ", THEME.BORDER_COLOR, title))
	}
}

func (l *TUIList) clearItems(itemsMap map[string]bool) {
	for key, _ := range itemsMap {
		itemsMap[key] = false
	}

	for row := 0; row < l.List.GetItemCount(); row++ {
		_, secondaryText := l.List.GetItemText(row)
		l.List.SetItemText(row, secondaryText, secondaryText)
	}
}
