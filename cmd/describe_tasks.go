package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func describeTasksCmd(config *dao.Config, configErr *error) *cobra.Command {
	var edit bool

	cmd := cobra.Command{
		Aliases: []string{"task", "tasks"},
		Use:     "tasks [tasks] [flags]",
		Short:   "Describe tasks",
		Long:    "Describe tasks.",
		Example: `  # Describe tasks
  mani describe tasks`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			describe(config, args, edit)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			values := config.GetTaskNames()
			return values, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVarP(&edit, "edit", "e", false, "Edit task")

	return &cmd
}

func describe(config *dao.Config, args []string, editFlag bool) {
	if editFlag {
		if len(args) > 0 {
			config.EditTask(args[0])
		} else {
			config.EditTask("")
		}
	} else {
		tasks := config.GetTasksByNames(args)

		for i := range tasks {
			var userEnv []string
			if len(args) > 1 {
				userEnv = args[1:]
			}

			tasks[i].SetEnvList(userEnv, []string{}, config.GetEnv())
			for j := range tasks[i].Commands {
				tasks[i].Commands[j].SetEnvList(userEnv, tasks[i].EnvList, config.GetEnv())
			}
		}

		print.PrintTaskBlock(tasks)
	}
}
