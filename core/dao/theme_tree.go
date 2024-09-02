package dao

import (
	"strings"
)

type Tree struct {
	Style string `yaml:"style"`
}

var DefaultTree = Tree{
	Style: "light",
}

func LoadTreeTheme(tree *Tree) {
	style := strings.ToLower(tree.Style)
	switch style {
	case "light":
		tree.Style = "light"
	case "bullet-flower":
		tree.Style = "bullet-flower"
	case "bullet-square":
		tree.Style = "bullet-square"
	case "bullet-star":
		tree.Style = "bullet-star"
	case "bullet-triangle":
		tree.Style = "bullet-triangle"
	case "bold":
		tree.Style = "bold"
	case "double":
		tree.Style = "double"
	case "rounded":
		tree.Style = "rounded"
	case "markdown":
		tree.Style = "markdown"
	default:
		tree.Style = "ascii"
	}
}
