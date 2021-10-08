package cmd

import (
	"github.com/spf13/cobra"

	// "github.com/jedib0t/go-pretty/v6/list"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func treeDirsCmd(config *dao.Config, configErr *error, treeFlags *print.TreeFlags) *cobra.Command {
	var dirPaths []string

	cmd := cobra.Command{
		Aliases: []string{"dir", "dr", "r"},
		Use:     "dirs [flags]",
		Short:   "list dirs in a tree-like format",
		Long:    "list dirs in a tree-like format.",
		Example: `  # example
  mani tree dirs`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runTreeDirs(config, treeFlags, &dirPaths)
		},
	}

	cmd.Flags().StringSliceVar(&dirPaths, "dir-paths", []string{}, "filter dirs by their path")
	err := cmd.RegisterFlagCompletionFunc("dir-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectDirs()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func runTreeDirs(
	config *dao.Config,
	treeFlags *print.TreeFlags,
	dirPaths *[]string,
) {
	// switch config.Theme.Tree {
	// case "square":
	// 	core.TreeStyle = list.StyleBulletSquare
	// case "circle":
	// 	core.TreeStyle = list.StyleBulletCircle
	// case "star":
	// 	core.TreeStyle = list.StyleBulletStar
	// case "line-bold":
	// 	core.TreeStyle = list.StyleConnectedBold
	// default:
	// 	core.TreeStyle = list.StyleConnectedLight
	// }

	tree := config.GetDirsTree(*dirPaths, treeFlags.Tags)
	print.PrintTree(treeFlags.Output, tree)
}
