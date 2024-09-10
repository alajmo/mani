package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUIList struct {
	List       *tview.List
	OnFocus    func()
	OnBlur     func()
	SelectItem func()
}

func (l *TUIList) createList(title string) {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorBlue).
		SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)).
		SetMainTextColor(tcell.ColorWhite)

	list.
		SetTitle(fmt.Sprintf("[::b] %s ", title)).
		SetBorder(true).
		SetBorderPadding(1, 1, 1, 1)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		numItems := list.GetItemCount()
		currentItem := list.GetCurrentItem()

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
		case ' ':
			l.SelectItem()
			return nil
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
