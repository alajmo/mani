package dao

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/list"

	"github.com/alajmo/mani/core"
)

func PrintTree(config *Config, treeFlags *core.TreeFlags, tree []core.TreeNode) {
	theme, err := config.GetTheme(treeFlags.Theme)
	core.CheckIfError(err)

	switch theme.Tree {
	case "square":
		core.TreeStyle = list.StyleBulletSquare
	case "circle":
		core.TreeStyle = list.StyleBulletCircle
	case "star":
		core.TreeStyle = list.StyleBulletStar
	case "line-bold":
		core.TreeStyle = list.StyleConnectedBold
	default:
		core.TreeStyle = list.StyleConnectedLight
	}

	l := list.NewWriter()
	l.SetStyle(core.TreeStyle)
	printTreeNodes(l, tree, 0)

	switch treeFlags.Output {
	case "markdown":
		printTree(l.RenderMarkdown())
	case "html":
		printTree(l.RenderHTML())
	default:
		printTree(l.Render())
	}
}

func printTreeNodes(l list.Writer, tree []core.TreeNode, depth int) {
	for _, n := range tree {
		for i := 0; i < depth; i++ {
			l.Indent()
		}

		l.AppendItem(n.Name)
		printTreeNodes(l, n.Children, depth+1)

		for i := 0; i < depth; i++ {
			l.UnIndent()
		}
	}
}

func printTree(content string) {
	for _, line := range strings.Split(content, "\n") {
		fmt.Printf("%s\n", line)
	}
	fmt.Println()
}
