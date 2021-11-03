package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func treeProjectsCmd(config *dao.Config, configErr *error, treeFlags *core.TreeFlags) *cobra.Command {
	var projectFlags core.ProjectFlags

	cmd := cobra.Command{
		Aliases: []string{"project", "proj", "p"},
		Use:     "projects [flags]",
		Short:   "list projects in a tree-like format",
		Long:    "list projects in a tree-like format.",
		Example: `  # example
  mani tree projects`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runTreeProjects(config, treeFlags, &projectFlags)
		},
	}

	cmd.Flags().StringSliceVarP(&projectFlags.Paths, "paths", "p", []string{}, "filter projects by their path")
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

func runTreeProjects(
	config *dao.Config,
	treeFlags *core.TreeFlags,
	projectFlags *core.ProjectFlags,
) {
	tree := config.GetProjectsTree(projectFlags.Paths, treeFlags.Tags)
	dao.PrintTree(config, treeFlags, tree)
}
