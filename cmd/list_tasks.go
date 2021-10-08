package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listTasksCmd(config *dao.Config, configErr *error, listFlags *print.ListFlags) *cobra.Command {
	var taskFlags print.ListTaskFlags

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
	listFlags *print.ListFlags,
	taskFlags *print.ListTaskFlags,
) {
	// Table Style
	// switch config.Theme.Table {
	// case "ascii":
	// 	core.ManiList.Box = core.StyleBoxASCII
	// default:
	// 	core.ManiList.Box = core.StyleBoxDefault
	// }

	tasks := config.GetTasksByNames(args)
	print.PrintTasks(tasks, *listFlags, *taskFlags)
}
