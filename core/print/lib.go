package print

import (
	"bufio"
	"strings"
	"unicode/utf8"
)

func GetMaxTextWidth(text string) int {
	scanner := bufio.NewScanner(strings.NewReader(text))
	maxWidth := 0

	for scanner.Scan() {
		lineWidth := utf8.RuneCountInString(scanner.Text())
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	return maxWidth
}

func GetTextDimensions(text string) (int, int) {
	// TODO: Seems it also counts color codes, so need to skip that
	scanner := bufio.NewScanner(strings.NewReader(text))
	maxWidth := 0
	height := 0

	for scanner.Scan() {
		height++
		lineWidth := utf8.RuneCountInString(scanner.Text())
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	return maxWidth, height
}
