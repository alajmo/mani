package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func treeCmd(config *dao.Config, configErr *error) *cobra.Command {
	var treeFlags core.TreeFlags

	cmd := cobra.Command{
		Aliases: []string{"t", "tree"},
		Use:     "tree <projects|dirs>",
		Short:   "List dirs, projects in a tree-like format",
		Long:    "List dirs, projects in a tree-like format.",
		Example: `  # example
  mani tree projects`,
	}
	cmd.AddCommand(
		treeProjectsCmd(config, configErr, &treeFlags),
		treeDirsCmd(config, configErr, &treeFlags),
	)
	cmd.PersistentFlags().StringVar(&treeFlags.Theme, "theme", "default", "Specify theme")
	err := cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		names := config.GetThemeNames()

		return names, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.PersistentFlags().StringVarP(&treeFlags.Output, "output", "o", "tree", "Output tree|markdown|html")
	err = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"tree", "markdown", "html"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.PersistentFlags().StringSliceVarP(&treeFlags.Tags, "tags", "t", []string{}, "filter entity by their tag")
	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		options := config.GetTags()
		return options, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}
