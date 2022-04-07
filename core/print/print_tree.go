package print

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/list"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func PrintTree(config *dao.Config, listFlags *core.ListFlags, tree []core.TreeNode) {
	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	switch theme.Tree {
	case "bullet-square":
		dao.TreeStyle = list.StyleBulletSquare
	case "bullet-circle":
		dao.TreeStyle = list.StyleBulletCircle
	case "bullet-star":
		dao.TreeStyle = list.StyleBulletStar
	case "connected-bold":
		dao.TreeStyle = list.StyleConnectedBold
	default: // connected-light
		dao.TreeStyle = list.StyleConnectedLight
	}

	l := list.NewWriter()
	l.SetStyle(dao.TreeStyle)
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
