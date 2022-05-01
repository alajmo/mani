package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/dao"
)

func describeCmd(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Aliases: []string{"desc"},
		Use:     "describe",
		Short:   "Describe projects and tasks",
		Long:    "Describe projects and tasks.",
		Example: `  # Describe all projects
  mani describe projects

  # Describe all tasks
  mani describe tasks`,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		describeProjectsCmd(config, configErr),
		describeTasksCmd(config, configErr),
	)

	return &cmd
}
