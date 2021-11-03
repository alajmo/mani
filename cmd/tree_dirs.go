package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func treeDirsCmd(config *dao.Config, configErr *error, treeFlags *core.TreeFlags) *cobra.Command {
	var dirFlags core.DirFlags

	cmd := cobra.Command{
		Aliases: []string{"dir", "dr", "r"},
		Use:     "dirs [flags]",
		Short:   "list dirs in a tree-like format",
		Long:    "list dirs in a tree-like format.",
		Example: `  # example
  mani tree dirs`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runTreeDirs(config, treeFlags, &dirFlags)
		},
	}

	cmd.Flags().StringSliceVarP(&dirFlags.Paths, "paths", "p", []string{}, "filter dirs by their path")
	err := cmd.RegisterFlagCompletionFunc("paths", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetProjectPaths()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func runTreeDirs(
	config *dao.Config,
	treeFlags *core.TreeFlags,
	dirFlags *core.DirFlags,
) {
	tree := config.GetDirsTree(dirFlags.Paths, treeFlags.Tags)
	dao.PrintTree(config, treeFlags, tree)
}
