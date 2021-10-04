package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func editTask(config *dao.Config, configErr *error) *cobra.Command {
	cmd := cobra.Command{
		Use:   "task",
		Short: "Edit mani task",
		Long:  `Edit mani task`,

		Example: `  # Edit a task called status
  mani edit task status

  # Edit task in specific mani config
  mani edit task status --config path/to/mani/config`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runEditTask(args, *config)
		},
		Args: cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil || len(args) == 1 {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return &cmd
}

func runEditTask(args []string, config dao.Config) {
	if len(args) > 0 {
		config.EditTask(args[0])
	} else {
		config.EditTask("")
	}
}
