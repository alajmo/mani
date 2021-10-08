package cmd

import (
	"github.com/spf13/cobra"

	// "github.com/jedib0t/go-pretty/v6/list"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func treeProjectsCmd(config *dao.Config, configErr *error, treeFlags *print.TreeFlags) *cobra.Command {
	var projectPaths []string

	cmd := cobra.Command{
		Aliases: []string{"project", "proj", "p"},
		Use:     "projects [flags]",
		Short:   "list projects in a tree-like format",
		Long:    "list projects in a tree-like format.",
		Example: `  # example
  mani tree projects`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runTreeProjects(config, treeFlags, &projectPaths)
		},
	}

	cmd.Flags().StringSliceVar(&projectPaths, "project-paths", []string{}, "filter projects by their path")
	err := cmd.RegisterFlagCompletionFunc("project-paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectDirs()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func runTreeProjects(
	config *dao.Config,
	treeFlags *print.TreeFlags,
	projectPaths *[]string,
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

	tree := config.GetProjectsTree(*projectPaths, treeFlags.Tags)
	print.PrintTree(treeFlags.Output, tree)
}
