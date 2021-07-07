package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/dao"
)

func describeCmd(config *dao.Config, configErr *error) *cobra.Command {
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
		describeProjectsCmd(config, configErr),
		describeCommandsCmd(config, configErr),
	)

	return &cmd
}
