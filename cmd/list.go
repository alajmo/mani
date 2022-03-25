package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func listCmd(config *dao.Config, configErr *error) *cobra.Command {
	var listFlags core.ListFlags

	cmd := cobra.Command{
		Aliases: []string{"ls", "l"},
		Use:     "list",
		Short:   "List projects, tasks and tags",
		Long:    "List projects, tasks and tags.",
		Example: `  # List all projects
  mani list projects

  # List all tasks
  mani list tasks

  # List all tags
  mani list tags`,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		listProjectsCmd(config, configErr, &listFlags),
		listTasksCmd(config, configErr, &listFlags),
		listTagsCmd(config, configErr, &listFlags),
	)

	cmd.PersistentFlags().StringVar(&listFlags.Theme, "theme", "default", "set theme")
	err := cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		names := config.GetThemeNames()

		return names, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.PersistentFlags().StringVarP(&listFlags.Output, "output", "o", "table", "set output [table|markdown|html]")
	err = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string{"table", "markdown", "html"}
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}
