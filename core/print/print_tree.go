package print

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/list"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func PrintTree(config *dao.Config, theme dao.Theme, listFlags *core.ListFlags, tree []dao.TreeNode) {
	// Style
	var treeStyle list.Style
	switch theme.Tree.Style {
	case "light":
		treeStyle = list.StyleConnectedLight
	case "bullet-flower":
		treeStyle = list.StyleBulletFlower
	case "bullet-square":
		treeStyle = list.StyleBulletSquare
	case "bullet-star":
		treeStyle = list.StyleBulletStar
	case "bullet-triangle":
		treeStyle = list.StyleBulletTriangle
	case "bold":
		treeStyle = list.StyleConnectedBold
	case "double":
		treeStyle = list.StyleConnectedDouble
	case "rounded":
		treeStyle = list.StyleConnectedRounded
	case "markdown":
		treeStyle = list.StyleMarkdown
	default:
		treeStyle = list.StyleDefault
	}

	// Print
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
		for range depth {
			l.Indent()
		}

		l.AppendItem(n.Path)
		printTreeNodes(l, n.Children, depth+1)

		for range depth {
			l.UnIndent()
		}
	}
}

func printTree(content string) {
	for line := range strings.SplitSeq(content, "\n") {
		fmt.Printf("%s\n", line)
	}
	fmt.Println()
}
