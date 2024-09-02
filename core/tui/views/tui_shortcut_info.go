package views

import (
	"fmt"
	"strings"

	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

type Shortcut struct {
	shortcut string
	label    string
}

func getShortcutInfo(shortcuts []Shortcut) string {
	var formattedShortcuts []string
	for _, s := range shortcuts {
		value := fmt.Sprintf("[%s:%s:%s]%s[-:-:-] [%s:%s:%s]%s[-:-:-]",
			misc.STYLE_SHORTCUT_LABEL.Fg, misc.STYLE_SHORTCUT_LABEL.Bg, misc.STYLE_SHORTCUT_LABEL.AttrStr, s.label,
			misc.STYLE_SHORTCUT_TEXT.Fg, misc.STYLE_SHORTCUT_TEXT.Bg, misc.STYLE_SHORTCUT_TEXT.AttrStr, s.shortcut,
		)
		formattedShortcuts = append(formattedShortcuts, value)
	}
	return strings.Join(formattedShortcuts, "  ")
}

func CreateRunInfoVIew() *tview.TextView {
	shortcuts := []Shortcut{
		{"Ctrl-r", "Run"},
		{"Ctrl-s", "Toggle View"},
		{"Ctrl-e", "Toggle Table/Tree"},
		{"Ctrl-o", "Options"},
	}
	text := getShortcutInfo(shortcuts)

	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func CreateExecInfoView() *tview.TextView {
	shortcuts := []Shortcut{
		{"Ctrl-r", "Run"},
		{"Ctrl-x", "Clear"},
		{"Ctrl-s", "Toggle View"},
		{"Ctrl-o", "Options"},
	}
	text := getShortcutInfo(shortcuts)

	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func CreateProjectInfoView() *tview.TextView {
	shortcuts := []Shortcut{
		{"Ctrl-e", "Toggle Table/Tree"},
	}
	text := getShortcutInfo(shortcuts)

	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}

func CreateTaskInfoView() *tview.TextView {
	shortcuts := []Shortcut{
		{"Ctrl-e", "Toggle Table/Tree"},
	}
	text := getShortcutInfo(shortcuts)
	helpInfo := tview.NewTextView().
		SetDynamicColors(true).
		SetText(text)
	helpInfo.SetTextAlign(tview.AlignRight)
	helpInfo.SetBorderPadding(0, 0, 0, 1)
	return helpInfo
}
