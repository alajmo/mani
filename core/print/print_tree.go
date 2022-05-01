package print

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/list"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func PrintTree(config *dao.Config, theme dao.Theme, listFlags *core.ListFlags, tree []dao.TreeNode) {
	var treeStyle list.Style
	switch theme.Tree.Style {
	case "bullet-square":
		treeStyle = list.StyleBulletSquare
	case "bullet-circle":
		treeStyle = list.StyleBulletCircle
	case "bullet-star":
		treeStyle = list.StyleBulletStar
	case "connected-bold":
		treeStyle = list.StyleConnectedBold
	default: // connected-light
		treeStyle = list.StyleConnectedLight
	}

	l := list.NewWriter()
	l.SetStyle(treeStyle)
	printTreeNodes(l, tree, 0)

	switch listFlags.Output {
	case "markdown":
		printTree(l.RenderMarkdown())
	case "html":
		printTree(l.RenderHTML())
	default:
		printTree(l.Render())
	}
}

func printTreeNodes(l list.Writer, tree []dao.TreeNode, depth int) {
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
