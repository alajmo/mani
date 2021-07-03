package cmd

import (
	"github.com/spf13/cobra"
)

func describeCmd(configFile *string) *cobra.Command {
	cmd := cobra.Command {
		Aliases: []string { "desc" },
		Use:   "describe <projects|commands>",
		Short: "Describe projects and commands",
		Long:  "Describe projects and commands.",
		Example: `  # Describe projects
  mani describe projects

  # Describe commands
  mani describe commands`,
	}

	cmd.AddCommand(
		describeCommandsCmd(configFile),
		describeProjectsCmd(configFile),
	)

	return &cmd
}
