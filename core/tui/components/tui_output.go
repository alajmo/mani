package components

import (
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

func CreateOutputView(title string) (*tview.TextView, *misc.ThreadSafeWriter) {
	streamView := CreateText(title)
	ansiWriter := misc.NewThreadSafeWriter(streamView)

	return streamView, ansiWriter
}
