package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func listTasksCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var taskFlags core.TaskFlags

	cmd := cobra.Command{
		Aliases: []string{"task", "tasks", "tsk", "tsks"},
		Use:     "tasks [flags]",
		Short:   "List tasks",
		Long:    "List tasks.",
		Example: `  # List tasks
  mani list tasks`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			listTasks(config, args, listFlags, &taskFlags)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVar(&taskFlags.Headers, "headers", []string{"name", "description"}, "Specify headers, defaults to name, description")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"name", "description"}
		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTasks(
	config *dao.Config,
	args []string,
	listFlags *core.ListFlags,
	taskFlags *core.TaskFlags,
) {
	tasks := config.GetTasksByNames(args)
	dao.PrintTasks(tasks, *listFlags, *taskFlags)
}
