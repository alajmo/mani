package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/dao"
)

func describeCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"desc"},
		Use:     "describe <projects|tasks>",
		Short:   "Describe projects and tasks",
		Long:    "Describe projects and tasks.",
		Example: `  # Describe projects
  mani describe projects

  # Describe tasks
  mani describe tasks`,
	}

	cmd.AddCommand(
		describeProjectsCmd(config, configErr),
		describeTasksCmd(config, configErr),
	)

	return &cmd
}
