package misc

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/rivo/tview"
	"golang.org/x/term"
)

func SetActive(box *tview.Box, title string, active bool) {
	if active {
		box.SetBorderColor(STYLE_BORDER_FOCUS.Fg)
		box.SetTitleAlign(STYLE_TITLE_ACTIVE.Align)
		title = dao.StyleFormat(title, STYLE_TITLE_ACTIVE.FormatStr)
		if title != "" {
			title = ColorizeTitle(title, *TUITheme.TitleActive)
			box.SetTitle(title)
		}
	} else {
		box.SetBorderColor(STYLE_BORDER.Fg)
		box.SetTitleAlign(STYLE_TITLE.Align)
		title = dao.StyleFormat(title, STYLE_TITLE.FormatStr)
		if title != "" {
			title = ColorizeTitle(title, *TUITheme.Title)
			box.SetTitle(title)
		}
	}
}

func GetTexztModalSize(text string) (int, int) {
	termWidth, termHeight, _ := term.GetSize(0)
	textWidth, textHeight := print.GetTextDimensions(text)

	width := textWidth
	height := textHeight

	// Min Width - sane minimum default width
	if width < 45 {
		width = 45
	}

	// Max Width - can't be wider than terminal width
	if width > termWidth {
		width = termWidth - 20 // Add some margin left/right
		height = height + 4    // Since text wraps, add some margin to height
	}

	// Max Height - can't be taller than terminal width
	if height > termHeight {
		height = termHeight - 5 // Add some margin top/bottom
	}

	width += 8  // Add some padding
	height += 2 // Add some padding

	return width, height
}
