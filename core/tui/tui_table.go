package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func createTable() *tview.Table {
	table := tview.NewTable()
	table.SetFixed(1, 0)           // Fixed header
	table.SetEvaluateAllRows(true) // Avoid resizing of headers when scrolling
	table.SetBorder(true).SetBorderPadding(0, 0, 2, 2)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorDefault)

	table.Select(1, 0)

	return table
}
