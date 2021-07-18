package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/alajmo/mani/core/dao"
)

func listCmd(config *dao.Config, configErr *error) *cobra.Command {
	var listFlags print.ListFlags

	cmd := cobra.Command {
		Use:   "list <projects|commands|tags>",
		Short: "List projects, commands and tags",
		Long:  "List projects, commands and tags.",
		Example: `  # List projects
  mani list projects

  # List commands
  mani list commands`,
	}

	cmd.AddCommand(
		listProjectsCmd(config, configErr, &listFlags),
		listCommandsCmd(config, configErr, &listFlags),
		listTagsCmd(config, configErr, &listFlags),
	)

	cmd.PersistentFlags().BoolVar(&listFlags.NoHeaders, "no-headers", false, "Remove table headers")
	cmd.PersistentFlags().BoolVar(&listFlags.NoBorders, "no-borders", false, "Remove table borders")
	cmd.PersistentFlags().StringVarP(&listFlags.Output, "output", "o", "table", "Output table|markdown|html")
	err := cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string { "table", "markdown", "html" }
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}
