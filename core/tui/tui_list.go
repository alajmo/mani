package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUIList struct {
	Title      string
	List       *tview.List
	Count      int
	OnFocus    func()
	OnBlur     func()
	SelectItem func(i int, mainText string, SecondaryText string)
}

func (l *TUIList) createList() {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		SetMainTextColor(tcell.ColorWhite)

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
			case 'g':
				list.SetCurrentItem(0)
				return nil
			case 'G':
				list.SetCurrentItem(numItems - 1)
				return nil
			case 'j':
				nextItem := currentItem + 1
				if nextItem < numItems {
					list.SetCurrentItem(nextItem)
				}
				return nil
			case 'k':
				nextItem := currentItem - 1
				if nextItem >= 0 {
					list.SetCurrentItem(nextItem)
				}
				return nil
			case ' ': // Space
				l.SelectItem(currentItem, mainText, secondaryText)
				return nil
			}
		}

		return event
	})

	list.SetFocusFunc(func() {
		l.OnFocus()
	})
	list.SetBlurFunc(func() {
		TUI.previousPage = list
		l.OnBlur()
	})

	l.List = list
}

func (l *TUIList) getTitle() string {
	if l.Count > 0 {
		return fmt.Sprintf("%s (%d)", l.Title, l.Count)
	}

	return l.Title
}
