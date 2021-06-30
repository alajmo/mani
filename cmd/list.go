package cmd

import (
	"github.com/spf13/cobra"
)

func listCmd(configFile *string) *cobra.Command {
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
		listCommandsCmd(configFile),
		listProjectsCmd(configFile),
		listTagsCmd(configFile),
	)

	return &cmd
}
