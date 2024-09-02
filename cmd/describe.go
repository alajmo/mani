package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func describeCmd(config *dao.Config, configErr *error) *cobra.Command {
	var describeFlags core.DescribeFlags

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
		describeProjectsCmd(config, configErr, &describeFlags),
		describeTasksCmd(config, configErr, &describeFlags),
	)

	cmd.PersistentFlags().StringVar(&describeFlags.Theme, "theme", "default", "set theme")
	err := cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}
		names := config.GetThemeNames()
		return names, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}
