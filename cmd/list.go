package cmd

import (
	"github.com/spf13/cobra"
)

func listCmd(configFile *string) *cobra.Command {
	var noHeaders bool
	var noBorders bool

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
		listProjectsCmd(configFile, &noHeaders, &noBorders),
		listCommandsCmd(configFile),
		listTagsCmd(configFile),
	)

	cmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "Remove table headers")
	cmd.PersistentFlags().BoolVar(&noBorders, "no-borders", false, "Remove table borders")

	return &cmd
}
