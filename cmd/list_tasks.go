package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func listTasksCmd(config *dao.Config, configErr *error, listFlags *core.ListFlags) *cobra.Command {
	var taskFlags core.TaskFlags

	cmd := cobra.Command{
		Aliases: []string{"task", "tsk", "tsks"},
		Use:     "tasks [tasks]",
		Short:   "List tasks",
		Long:    "List tasks.",
		Example: `  # List all tasks
  mani list tasks

  # List task <task>
  mani list task <task>`,
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
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringSliceVar(&taskFlags.Headers, "headers", []string{"task", "description"}, "set headers. Available headers: task, description")
	err := cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string{"task", "description"}
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
	tasks, err := config.GetTasksByNames(args)
	core.CheckIfError(err)

	theme, err := config.GetTheme(listFlags.Theme)
	core.CheckIfError(err)

	if len(tasks) == 0 {
		fmt.Println("No tasks")
	} else {
		options := print.PrintTableOptions{
			Output:               listFlags.Output,
			Theme:                *theme,
			Tree:                 listFlags.Tree,
			OmitEmpty:            false,
			SuppressEmptyColumns: true,
		}

		print.PrintTable(tasks, options, taskFlags.Headers, []string{})
	}
}
